// Package signatures cung cấp tính năng cập nhật YARA rules từ yara-forge.
package util

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	yaraForgeURL = "https://github.com/YARAHQ/yara-forge/releases/latest/download/yara-forge-rules-core.zip"
	yaraDestDir  = "signatures/yara"
	yaraTmpFile  = "yara-forge-rules-core.zip"
)

// UpdateSignatures tải và cập nhật YARA rules từ yara-forge.
// Được gọi khi chương trình chạy với flag --update-signatures.
func UpdateSignatures() error {
	fmt.Println("→ Đang tải YARA rules từ yara-forge...")

	start := time.Now()
	if err := downloadFile(yaraForgeURL, yaraTmpFile); err != nil {
		return fmt.Errorf("tải file thất bại: %w", err)
	}
	fmt.Printf("  ✓ Tải xong (%s)\n", time.Since(start).Round(time.Millisecond))

	defer func() {
		os.Remove(yaraTmpFile)
	}()

	if err := os.MkdirAll(yaraDestDir, 0755); err != nil {
		return fmt.Errorf("không tạo được thư mục '%s': %w", yaraDestDir, err)
	}

	fmt.Printf("→ Giải nén *.yar vào %s ...\n", yaraDestDir)
	count, err := extractYarFiles(yaraTmpFile, yaraDestDir)
	if err != nil {
		return fmt.Errorf("giải nén thất bại: %w", err)
	}

	fmt.Printf("✓ Cập nhật hoàn tất: %d file .yar\n", count)
	return nil
}

// downloadFile tải URL về localPath với progress bar.
func downloadFile(url, localPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	pr := &progressReader{r: resp.Body, total: resp.ContentLength}
	_, err = io.Copy(out, pr)
	fmt.Println()
	return err
}

// extractYarFiles giải nén tất cả *.yar từ zipPath vào destDir.
func extractYarFiles(zipPath, destDir string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(f.Name), ".yar") {
			continue
		}

		destPath := filepath.Join(destDir, filepath.Base(f.Name))
		if err := extractEntry(f, destPath); err != nil {
			return count, fmt.Errorf("copy '%s': %w", filepath.Base(f.Name), err)
		}
		count++
	}

	if count == 0 {
		return 0, fmt.Errorf("không tìm thấy file .yar nào trong zip")
	}
	return count, nil
}

// extractEntry ghi một entry zip ra file đích.
func extractEntry(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// progressReader in tiến trình download ra stdout.
type progressReader struct {
	r       io.Reader
	total   int64
	current int64
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	p.current += int64(n)

	if p.total > 0 {
		pct := float64(p.current) / float64(p.total) * 100
		filled := int(pct / 5)
		fmt.Printf("\r  [%-20s] %5.1f%%  %s / %s",
			strings.Repeat("█", filled)+strings.Repeat("░", 20-filled),
			pct,
			humanSize(p.current),
			humanSize(p.total),
		)
	} else {
		fmt.Printf("\r  Đã tải: %s", humanSize(p.current))
	}
	return n, err
}

func humanSize(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

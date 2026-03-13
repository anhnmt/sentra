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

	"github.com/anhnmt/sentra/internal/progress"
)

const (
	yaraForgeURL = "https://github.com/YARAHQ/yara-forge/releases/latest/download/yara-forge-rules-core.zip"
	yaraDestDir  = "signatures/yara"
	yaraTmpFile  = "yara-forge-rules-core.zip"
)

func UpdateSignatures() error {
	fmt.Println("→ Downloading YARA rules from yara-forge...")

	start := time.Now()
	if err := downloadFile(yaraForgeURL, yaraTmpFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Printf("  ✓ Downloaded (%s)\n", time.Since(start).Round(time.Millisecond))
	defer os.Remove(yaraTmpFile)

	if err := os.MkdirAll(yaraDestDir, 0755); err != nil {
		return fmt.Errorf("mkdir '%s': %w", yaraDestDir, err)
	}

	fmt.Printf("→ Extracting *.yar into %s...\n", yaraDestDir)
	count, err := extractYarFiles(yaraTmpFile, yaraDestDir)
	if err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	fmt.Printf("✓ Updated: %d .yar files\n", count)
	return nil
}

func downloadFile(url, localPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// wrap body với progress bar
	proxy, wait := progress.NewDownloadBar(resp.Body, resp.ContentLength, "downloading")
	defer wait()
	defer proxy.Close()

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, proxy)
	return err
}

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
		return 0, fmt.Errorf("no .yar files found in zip")
	}
	return count, nil
}

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

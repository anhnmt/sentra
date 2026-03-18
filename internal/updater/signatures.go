package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/anhnmt/sentra/internal/progress"
)

const (
	yaraForgeURL    = "https://github.com/YARAHQ/yara-forge/releases/latest/download/yara-forge-rules-core.zip"
	yaraForgeAPIURL = "https://api.github.com/repos/YARAHQ/yara-forge/releases/latest"
	yaraDestDir     = "signatures/yara"
	yaraTmpFile     = "yara-forge-rules-core.zip"
	yaraVersionFile = "signatures/yara/.version"
)

func UpdateSignatures() error {
	log.Info().Msg("Checking latest yara-forge release")

	latest, err := fetchLatestTag()
	if err != nil {
		return fmt.Errorf("version check failed: %w", err)
	}

	if cached, _ := readCachedVersion(); cached == latest {
		log.Info().Str("version", latest).Msg("Already up-to-date")
		return nil
	}

	log.Info().Str("version", latest).Msg("Downloading YARA rules")

	start := time.Now()
	if err := downloadFile(yaraForgeURL, yaraTmpFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	log.Info().Dur("elapsed", time.Since(start).Round(time.Millisecond)).Msg("Downloaded")
	defer os.Remove(yaraTmpFile)

	if err := os.MkdirAll(yaraDestDir, 0755); err != nil {
		return fmt.Errorf("mkdir '%s': %w", yaraDestDir, err)
	}

	log.Info().Str("dest", yaraDestDir).Msg("Extracting .yar files")
	count, err := extractYarFiles(yaraTmpFile, yaraDestDir)
	if err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	if err := writeCachedVersion(latest); err != nil {
		log.Warn().Err(err).Msg("Could not save version cache")
	}

	log.Info().Str("version", latest).Int("count", count).Msg("Updated .yar files")
	return nil
}

func fetchLatestTag() (string, error) {
	req, err := http.NewRequest(http.MethodGet, yaraForgeAPIURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "sentra-updater")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if release.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response")
	}
	return release.TagName, nil
}

func readCachedVersion() (string, error) {
	b, err := os.ReadFile(yaraVersionFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func writeCachedVersion(tag string) error {
	return os.WriteFile(yaraVersionFile, []byte(tag+"\n"), 0644)
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

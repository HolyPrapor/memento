package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const repoAPI = "https://api.github.com/repos/HolyPrapor/memento/releases/latest"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpdate() {
	current := strings.TrimPrefix(version, "v")

	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: fetch latest release: %s\n", err)
		os.Exit(1)
	}

	latest := strings.TrimPrefix(release.TagName, "v")

	if current != "dev" && !versionLess(current, latest) {
		fmt.Println("Already up to date.")
		return
	}

	if current != "dev" {
		fmt.Printf("Updating from %s to %s\n", current, latest)
	} else {
		fmt.Printf("Updating to %s\n", latest)
	}

	assetName := assetNameForPlatform(latest)
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		fmt.Fprintf(os.Stderr, "Error: no asset found for %s/%s (expected %s)\n", runtime.GOOS, runtime.GOARCH, assetName)
		os.Exit(1)
	}

	tmpFile, err := downloadBinary(downloadURL, assetName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: download: %s\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	newExe, err := extractBinary(tmpFile, assetName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: extract: %s\n", err)
		os.Exit(1)
	}
	defer os.Remove(newExe)

	if err := replaceSelf(newExe); err != nil {
		fmt.Fprintf(os.Stderr, "Error: replace: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Updated. Restart to use the new version.")
}

func fetchLatestRelease() (*ghRelease, error) {
	resp, err := http.Get(repoAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func assetNameForPlatform(version string) string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("memento_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, ext)
}

func downloadBinary(url, name string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "memento-*"+filepath.Ext(name))
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

func extractBinary(archivePath, archiveName string) (string, error) {
	if strings.HasSuffix(archiveName, ".zip") {
		return extractFromZip(archivePath)
	}
	return extractFromTarGz(archivePath)
}

func extractFromZip(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			base := filepath.Base(f.Name)
			if base == "memento" || base == "memento.exe" {
				return copyZipFile(f)
			}
		}
		if f.FileInfo().IsDir() {
			for _, f2 := range r.File {
				base := filepath.Base(f2.Name)
				if base == "memento" || base == "memento.exe" {
					return copyZipFile(f2)
				}
			}
		}
	}

	return "", fmt.Errorf("memento binary not found in archive")
}

func copyZipFile(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	tmp, err := os.CreateTemp("", "memento-exe-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, rc); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

func extractFromTarGz(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		base := filepath.Base(hdr.Name)
		if hdr.Typeflag == tar.TypeReg && (base == "memento" || base == "memento.exe") {
			tmp, err := os.CreateTemp("", "memento-exe-*")
			if err != nil {
				return "", err
			}
			defer tmp.Close()

			if _, err := io.Copy(tmp, tr); err != nil {
				os.Remove(tmp.Name())
				return "", err
			}

			return tmp.Name(), nil
		}
	}

	return "", fmt.Errorf("memento binary not found in archive")
}

func replaceSelf(newPath string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	oldPath := self + ".old"

	if err := os.Rename(self, oldPath); err != nil {
		return fmt.Errorf("rename current binary: %w", err)
	}

	if err := os.Rename(newPath, self); err != nil {
		os.Rename(oldPath, self)
		return fmt.Errorf("install new binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(self, 0755); err != nil {
			return fmt.Errorf("chmod: %w", err)
		}
	}

	os.Remove(oldPath)
	return nil
}

func versionLess(a, b string) bool {
	ap := parseVersion(a)
	bp := parseVersion(b)
	for i := 0; i < 3; i++ {
		if ap[i] < bp[i] {
			return true
		}
		if ap[i] > bp[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	var parts [3]int
	for i, s := range strings.SplitN(v, ".", 3) {
		fmt.Sscanf(s, "%d", &parts[i])
	}
	return parts
}

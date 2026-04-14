package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"dxtrity/super/src/config"
	"dxtrity/super/src/util"

	"github.com/pelletier/go-toml"
)

const githubRepo = "kkam05/super"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func CmdUpdate(args []string) {
	var local bool
	for _, a := range args {
		if a == "--local" {
			local = true
		}
	}

	if local {
		updateLocal()
	} else {
		updateRemote()
	}
}

// ── local install ──────────────────────────────────────────────────────────────

func updateLocal() {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		util.PrintError(err.Error())
		os.Exit(1)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	srcPath := filepath.Join(projectRoot, "build", "super"+ext)

	if _, err := os.Stat(srcPath); err != nil {
		util.PrintError(fmt.Sprintf("binary not found at %s — run 'super run build' first", filepath.Join("build", "super"+ext)))
		os.Exit(1)
	}

	// Read project version from project.settings
	settingsPath := filepath.Join(projectRoot, "project.settings")
	cfg, err := config.LoadSettings(settingsPath)
	newVersion := "unknown"
	if err == nil && cfg.Project.Version != "" {
		newVersion = cfg.Project.Version
	}

	home, err := os.UserHomeDir()
	if err != nil {
		util.PrintError("could not determine home directory: " + err.Error())
		os.Exit(1)
	}

	binDir := filepath.Join(home, ".super", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		util.PrintError("could not create ~/.super/bin: " + err.Error())
		os.Exit(1)
	}

	dstPath := filepath.Join(binDir, "super"+ext)
	if err := copyFile(srcPath, dstPath, 0755); err != nil {
		util.PrintError("could not install binary: " + err.Error())
		os.Exit(1)
	}
	util.PrintStep("installed", dstPath)

	writeGlobalSettings(home, newVersion, "local")

	fmt.Println()
	util.PrintSuccess(fmt.Sprintf("super v%s installed to %s", newVersion, dstPath))
	util.PrintInfo(fmt.Sprintf("make sure %s is on your PATH", binDir))
}

// ── remote install ─────────────────────────────────────────────────────────────

func updateRemote() {
	util.PrintInfo("checking for latest release...")

	release, err := fetchLatestRelease()
	if err != nil {
		util.PrintError("could not fetch release info: " + err.Error())
		os.Exit(1)
	}

	assetName := fmt.Sprintf("super-%s-%s.zip", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		util.PrintError(fmt.Sprintf("no release asset found for %s/%s (expected %s)", runtime.GOOS, runtime.GOARCH, assetName))
		os.Exit(1)
	}

	newVersion := strings.TrimPrefix(release.TagName, "v")

	home, err := os.UserHomeDir()
	if err != nil {
		util.PrintError("could not determine home directory: " + err.Error())
		os.Exit(1)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	binDir := filepath.Join(home, ".super", "bin")
	tmpDir := filepath.Join(home, ".super", "tmp")
	backupDir := filepath.Join(home, ".super", "backup")
	for _, d := range []string{binDir, tmpDir, backupDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			util.PrintError("could not create directory " + d + ": " + err.Error())
			os.Exit(1)
		}
	}

	dstPath := filepath.Join(binDir, "super"+ext)

	// Backup current binary if it exists
	if _, err := os.Stat(dstPath); err == nil {
		curVersion := currentInstalledVersion(home)
		backupName := fmt.Sprintf("super-v%s-%s-%s.zip", curVersion, runtime.GOOS, runtime.GOARCH)
		backupPath := filepath.Join(backupDir, backupName)
		if err := zipBinary(dstPath, backupPath, "super"+ext); err != nil {
			util.PrintWarn("could not create backup: " + err.Error())
		} else {
			util.PrintStep("backed up", backupPath)
		}
	}

	// Download release zip
	tmpZip := filepath.Join(tmpDir, assetName)
	util.PrintInfo(fmt.Sprintf("downloading %s...", release.TagName))
	if err := downloadFile(downloadURL, tmpZip); err != nil {
		util.PrintError("download failed: " + err.Error())
		os.Exit(1)
	}
	util.PrintStep("downloaded", tmpZip)

	// Extract binary from zip
	binaryName := "super" + ext
	if err := extractFromZip(tmpZip, binaryName, dstPath, 0755); err != nil {
		util.PrintError("could not extract binary: " + err.Error())
		os.Exit(1)
	}
	util.PrintStep("installed", dstPath)

	// Clean up tmp
	_ = os.Remove(tmpZip)

	writeGlobalSettings(home, newVersion, "remote")

	fmt.Println()
	util.PrintSuccess(fmt.Sprintf("super v%s installed to %s", newVersion, dstPath))
	util.PrintInfo(fmt.Sprintf("make sure %s is on your PATH", binDir))
}

// ── helpers ────────────────────────────────────────────────────────────────────

func fetchLatestRelease() (*githubRelease, error) {
	url := "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "super-cli/"+config.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func downloadFile(url, dst string) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d downloading asset", resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// zipBinary creates a zip archive at dst containing src as nameInZip.
func zipBinary(src, dst, nameInZip string) error {
	zf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zf.Close()

	w := zip.NewWriter(zf)
	defer w.Close()

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	hdr.Name = nameInZip
	hdr.Method = zip.Deflate

	wr, err := w.CreateHeader(hdr)
	if err != nil {
		return err
	}

	_, err = io.Copy(wr, f)
	return err
}

// extractFromZip finds binaryName (by base name) inside zipPath and writes it to dst.
func extractFromZip(zipPath, binaryName, dst string, perm os.FileMode) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == binaryName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("binary %q not found in zip", binaryName)
}

// currentInstalledVersion reads the version from ~/.super/super.settings.
func currentInstalledVersion(home string) string {
	settingsPath := filepath.Join(home, ".super", "super.settings")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return "unknown"
	}
	var gcfg config.GlobalConfig
	if err := toml.Unmarshal(data, &gcfg); err != nil {
		return "unknown"
	}
	if gcfg.Super.Version == "" {
		return "unknown"
	}
	return gcfg.Super.Version
}

func writeGlobalSettings(home, newVersion, method string) {
	globalSettingsPath := filepath.Join(home, ".super", "super.settings")
	gcfg := &config.GlobalConfig{
		Super: config.GlobalSuperSection{
			Version:       newVersion,
			InstallMethod: method,
			UpdatedAt:     time.Now().Format("2006-01-02"),
		},
	}
	b, err := toml.Marshal(gcfg)
	if err == nil {
		if err := os.WriteFile(globalSettingsPath, b, 0644); err != nil {
			util.PrintWarn("could not write super.settings: " + err.Error())
		} else {
			util.PrintStep("updated", globalSettingsPath)
		}
	}
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pelletier/go-toml"
)

type globalConfig struct {
	Super globalSuperSection `toml:"super"`
}

type globalSuperSection struct {
	Version       string `toml:"version"`
	InstallMethod string `toml:"install_method"`
	UpdatedAt     string `toml:"updated_at"`
}

func cmdUpdate(args []string) {
	var local bool
	for _, a := range args {
		if a == "--local" {
			local = true
		}
	}

	if !local {
		printError("usage: super update --local")
		os.Exit(1)
	}

	updateLocal()
}

func updateLocal() {
	projectRoot, err := findProjectRoot()
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	srcPath := filepath.Join(projectRoot, "build", "super"+ext)

	if _, err := os.Stat(srcPath); err != nil {
		printError(fmt.Sprintf("binary not found at %s — run 'super run build' first", filepath.Join("build", "super"+ext)))
		os.Exit(1)
	}

	// Read project version from project.settings
	settingsPath := filepath.Join(projectRoot, "project.settings")
	cfg, err := loadSettings(settingsPath)
	newVersion := "unknown"
	if err == nil && cfg.Project.Version != "" {
		newVersion = cfg.Project.Version
	}

	home, err := os.UserHomeDir()
	if err != nil {
		printError("could not determine home directory: " + err.Error())
		os.Exit(1)
	}

	binDir := filepath.Join(home, ".super", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		printError("could not create ~/.super/bin: " + err.Error())
		os.Exit(1)
	}

	dstPath := filepath.Join(binDir, "super"+ext)
	if err := copyFile(srcPath, dstPath, 0755); err != nil {
		printError("could not install binary: " + err.Error())
		os.Exit(1)
	}
	printStep("installed", dstPath)

	// Write ~/.super/super.settings
	globalSettingsPath := filepath.Join(home, ".super", "super.settings")
	gcfg := &globalConfig{
		Super: globalSuperSection{
			Version:       newVersion,
			InstallMethod: "local",
			UpdatedAt:     time.Now().Format("2006-01-02"),
		},
	}
	b, err := toml.Marshal(gcfg)
	if err == nil {
		if err := os.WriteFile(globalSettingsPath, b, 0644); err != nil {
			printWarn("could not write super.settings: " + err.Error())
		} else {
			printStep("updated", globalSettingsPath)
		}
	}

	fmt.Println()
	printSuccess(fmt.Sprintf("super v%s installed to %s", newVersion, dstPath))
	printInfo(fmt.Sprintf("make sure %s is on your PATH", binDir))
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

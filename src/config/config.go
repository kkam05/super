package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

// Version is set by main at startup (via -ldflags -X main.version=...).
var Version string

// ── project-level types ────────────────────────────────────────────────────────

type ProjectConfig struct {
	Project ProjectSection    `toml:"project"`
	Scripts map[string]string `toml:"scripts"`
}

type ProjectSection struct {
	Name         string `toml:"name"`
	Version      string `toml:"version,omitempty"`
	Description  string `toml:"description,omitempty"`
	SuperVersion string `toml:"super_version,omitempty"`
}

// ── global (installed super) types ────────────────────────────────────────────

type GlobalConfig struct {
	Super GlobalSuperSection `toml:"super"`
}

type GlobalSuperSection struct {
	Version       string `toml:"version"`
	InstallMethod string `toml:"install_method"`
	UpdatedAt     string `toml:"updated_at"`
}

// ── project root discovery ─────────────────────────────────────────────────────

func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "project.settings")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a super project (no project.settings found)")
		}
		dir = parent
	}
}

// ── settings I/O ──────────────────────────────────────────────────────────────

func LoadSettings(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Scripts == nil {
		cfg.Scripts = make(map[string]string)
	}
	return &cfg, nil
}

func SaveSettings(path string, cfg *ProjectConfig) error {
	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

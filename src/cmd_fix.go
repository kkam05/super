package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func cmdFix(_ []string) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		cwd, _ := os.Getwd()
		projectRoot = cwd
		printInfo("no project.settings found — fixing in current directory")
	}

	var fixed int

	// ── 1. directory structure ─────────────────────────────────────────────────
	dirs := []string{
		filepath.Join(".super", "scripts"),
		"build",
		"lib",
		"src",
	}
	for _, d := range dirs {
		full := filepath.Join(projectRoot, d)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			if err := os.MkdirAll(full, 0755); err != nil {
				printError(fmt.Sprintf("could not create %s: %v", d, err))
				os.Exit(1)
			}
			printStep("created", d+string(filepath.Separator))
			fixed++
		}
	}

	// ── 2. project.settings ────────────────────────────────────────────────────
	settingsPath := filepath.Join(projectRoot, "project.settings")
	cfg, loadErr := loadSettings(settingsPath)
	settingsChanged := false

	if loadErr != nil {
		cfg = &projectConfig{
			Project: projectSection{
				Name:         filepath.Base(projectRoot),
				SuperVersion: version,
			},
			Scripts: make(map[string]string),
		}
		settingsChanged = true
	}

	if cfg.Project.Name == "" {
		cfg.Project.Name = filepath.Base(projectRoot)
		settingsChanged = true
	}

	if cfg.Project.Version == "" {
		cfg.Project.Version = "0.1.0"
		settingsChanged = true
	}

	if cfg.Project.SuperVersion != version {
		cfg.Project.SuperVersion = version
		settingsChanged = true
	}

	for _, s := range []string{"build", "run", "dev"} {
		if _, ok := cfg.Scripts[s]; !ok {
			cfg.Scripts[s] = ".super/scripts"
			settingsChanged = true
		}
	}

	if settingsChanged {
		if err := saveSettings(settingsPath, cfg); err != nil {
			printError("could not write project.settings: " + err.Error())
			os.Exit(1)
		}
		if loadErr != nil {
			printStep("created", "project.settings")
		} else {
			printStep("updated", "project.settings")
		}
		fixed++
	}

	// ── 3. .super/version ─────────────────────────────────────────────────────
	versionFilePath := filepath.Join(projectRoot, ".super", "version")
	existing, readErr := os.ReadFile(versionFilePath)
	if readErr != nil || strings.TrimSpace(string(existing)) != version {
		if err := os.WriteFile(versionFilePath, []byte(version+"\n"), 0644); err != nil {
			printError("could not write .super/version: " + err.Error())
			os.Exit(1)
		}
		if readErr != nil {
			printStep("created", filepath.Join(".super", "version"))
		} else {
			printStep("updated", filepath.Join(".super", "version"))
		}
		fixed++
	}

	// ── 4. default scripts ─────────────────────────────────────────────────────
	isWindows := runtime.GOOS == "windows"
	scriptsDir := filepath.Join(projectRoot, ".super", "scripts")
	name := cfg.Project.Name

	type scriptEntry struct {
		filename string
		content  string
		perm     os.FileMode
	}

	var scripts []scriptEntry
	if isWindows {
		scripts = []scriptEntry{
			{"build.ps1", buildPS1(name), 0644},
			{"run.ps1", runPS1(name), 0644},
			{"dev.ps1", devPS1(), 0644},
		}
	} else {
		scripts = []scriptEntry{
			{"build.sh", buildSH(name), 0755},
			{"run.sh", runSH(name), 0755},
			{"dev.sh", devSH(), 0755},
		}
	}

	for _, s := range scripts {
		p := filepath.Join(scriptsDir, s.filename)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, []byte(s.content), s.perm); err != nil {
				printError(fmt.Sprintf("could not write %s: %v", s.filename, err))
				os.Exit(1)
			}
			printStep("created", filepath.Join(".super", "scripts", s.filename))
			fixed++
		}
	}

	fmt.Println()
	if fixed == 0 {
		printSuccess("project is already up to date.")
	} else {
		printSuccess(fmt.Sprintf("fixed %d issue(s).", fixed))
	}
}

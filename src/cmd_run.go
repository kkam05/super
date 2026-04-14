package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml"
)

func cmdRun(args []string) {
	if len(args) == 0 {
		printError("usage: super run <script> [args...]")
		os.Exit(1)
	}
	scriptName := args[0]
	passthrough := args[1:]

	projectRoot, err := findProjectRoot()
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	settingsPath := filepath.Join(projectRoot, "project.settings")
	cfg, err := loadSettings(settingsPath)
	if err != nil {
		printError("could not read project.settings: " + err.Error())
		os.Exit(1)
	}

	scriptValue, ok := cfg.Scripts[scriptName]
	if !ok {
		// Check if a script file exists in .super/scripts/ and auto-register it.
		if autoPath := findScriptFile(projectRoot, scriptName); autoPath != "" {
			cfg.Scripts[scriptName] = ".super/scripts"
			if err := saveSettings(settingsPath, cfg); err != nil {
				printWarn("could not update project.settings: " + err.Error())
			} else {
				printInfo(fmt.Sprintf("registered %q in [scripts] in project.settings", scriptName))
			}
			scriptValue = ".super/scripts"
		} else {
			printError(fmt.Sprintf("unknown script %q — add it to [scripts] in project.settings", scriptName))
			os.Exit(1)
		}
	}

	runScript(projectRoot, scriptName, scriptValue, passthrough)
}

// ── project root discovery ─────────────────────────────────────────────────────

func findProjectRoot() (string, error) {
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

func loadSettings(path string) (*projectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg projectConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Scripts == nil {
		cfg.Scripts = make(map[string]string)
	}
	return &cfg, nil
}

func saveSettings(path string, cfg *projectConfig) error {
	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// ── script resolution & execution ─────────────────────────────────────────────

// findScriptFile returns the path to <name>.sh / <name>.ps1 inside
// .super/scripts/ if one exists, otherwise "".
func findScriptFile(root, name string) string {
	ext := ".sh"
	if runtime.GOOS == "windows" {
		ext = ".ps1"
	}
	p := filepath.Join(root, ".super", "scripts", name+ext)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// resolveScriptTarget works out what to execute for a given value from
// [scripts]. Returns (kind, resolved) where kind is "dir", "file", or "inline".
func resolveScriptTarget(projectRoot, value string) (kind, resolved string) {
	// If the value contains spaces it can't be a bare path – treat as inline.
	if strings.Contains(value, " ") {
		return "inline", value
	}
	abs := filepath.Join(projectRoot, value)
	info, err := os.Stat(abs)
	if err != nil {
		return "inline", value
	}
	if info.IsDir() {
		return "dir", abs
	}
	return "file", abs
}

func runScript(projectRoot, scriptName, value string, passthrough []string) {
	isWindows := runtime.GOOS == "windows"
	kind, resolved := resolveScriptTarget(projectRoot, value)

	var cmd *exec.Cmd

	switch kind {
	case "dir":
		// Expect <dir>/<scriptName>.sh (or .ps1)
		ext := ".sh"
		if isWindows {
			ext = ".ps1"
		}
		scriptPath := filepath.Join(resolved, scriptName+ext)
		if _, err := os.Stat(scriptPath); err != nil {
			printError(fmt.Sprintf("script file not found: %s", scriptPath))
			os.Exit(1)
		}
		cmd = buildScriptCmd(isWindows, scriptPath, passthrough)

	case "file":
		cmd = buildScriptCmd(isWindows, resolved, passthrough)

	case "inline":
		// Append passthrough args to the inline command string and run via shell.
		full := value
		if len(passthrough) > 0 {
			full += " " + strings.Join(passthrough, " ")
		}
		if isWindows {
			cmd = exec.Command("cmd", "/C", full)
		} else {
			cmd = exec.Command("sh", "-c", full)
		}
	}

	cmd.Dir = projectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	printStep("run", scriptName)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		printError("script failed: " + err.Error())
		os.Exit(1)
	}
}

func buildScriptCmd(isWindows bool, scriptPath string, passthrough []string) *exec.Cmd {
	if isWindows {
		args := append([]string{"-File", scriptPath}, passthrough...)
		return exec.Command("powershell", args...)
	}
	args := append([]string{scriptPath}, passthrough...)
	return exec.Command("bash", args...)
}

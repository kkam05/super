package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func cmdPath(_ []string) {
	home, err := os.UserHomeDir()
	if err != nil {
		printError("could not determine home directory: " + err.Error())
		os.Exit(1)
	}

	binDir := filepath.Join(home, ".super", "bin")

	if isOnPath(binDir) {
		printSuccess(fmt.Sprintf("%s is already on your PATH", binDir))
		return
	}

	printInfo(fmt.Sprintf("%s is not on your PATH — adding it now...", binDir))

	if runtime.GOOS == "windows" {
		addToPathWindows(binDir)
	} else {
		addToPathUnix(binDir)
	}
}

// isOnPath reports whether dir appears in the current PATH.
func isOnPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	for _, p := range strings.Split(pathEnv, sep) {
		if filepath.Clean(p) == filepath.Clean(dir) {
			return true
		}
	}
	return false
}

// addToPathUnix appends an export line to the user's shell config file.
func addToPathUnix(binDir string) {
	home, _ := os.UserHomeDir()
	line := fmt.Sprintf(`export PATH="%s:$PATH"`, binDir)

	cfg := detectShellConfig(home)
	if cfg == "" {
		printWarn("could not detect shell config file — add the following line manually:")
		fmt.Printf("  %s\n", line)
		return
	}

	// Check if the line is already present (e.g. from a previous run that
	// didn't affect the running shell's PATH).
	data, _ := os.ReadFile(cfg)
	if strings.Contains(string(data), binDir) {
		printInfo(fmt.Sprintf("PATH entry already exists in %s", cfg))
		printInfo("restart your terminal or run: source " + cfg)
		return
	}

	f, err := os.OpenFile(cfg, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		printError("could not open " + cfg + ": " + err.Error())
		os.Exit(1)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n# added by super\n%s\n", line); err != nil {
		printError("could not write to " + cfg + ": " + err.Error())
		os.Exit(1)
	}

	printStep("updated", cfg)
	fmt.Println()
	printSuccess(fmt.Sprintf("%s added to PATH in %s", binDir, cfg))
	printInfo("restart your terminal or run: source " + cfg)
}

// detectShellConfig returns the most appropriate shell rc file for the user.
func detectShellConfig(home string) string {
	shell := os.Getenv("SHELL")

	candidates := []string{}

	switch {
	case strings.Contains(shell, "zsh"):
		candidates = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".zprofile"),
		}
	case strings.Contains(shell, "bash"):
		candidates = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".profile"),
		}
	case strings.Contains(shell, "fish"):
		candidates = []string{
			filepath.Join(home, ".config", "fish", "config.fish"),
		}
	default:
		candidates = []string{
			filepath.Join(home, ".profile"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
		}
	}

	// Return the first candidate that already exists; otherwise return the
	// first candidate so we can create it.
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

// addToPathWindows adds binDir to the user PATH via setx.
func addToPathWindows(binDir string) {
	// Read the current user PATH from the registry via PowerShell so we can
	// append rather than overwrite (setx truncates values > 1024 chars if
	// given directly, so we expand first).
	psGetPath := `[Environment]::GetEnvironmentVariable('Path','User')`
	out, err := exec.Command("powershell", "-NoProfile", "-Command", psGetPath).Output()
	if err != nil {
		printError("could not read user PATH from registry: " + err.Error())
		os.Exit(1)
	}

	currentPath := strings.TrimSpace(string(out))

	// Check whether binDir is already in the persistent PATH (the running
	// process PATH may differ from the registry value).
	for _, p := range strings.Split(currentPath, ";") {
		if strings.EqualFold(filepath.Clean(p), filepath.Clean(binDir)) {
			printInfo("PATH entry already exists in user environment")
			printInfo("restart your terminal for the change to take effect")
			return
		}
	}

	newPath := currentPath
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += binDir

	psSetPath := fmt.Sprintf(`[Environment]::SetEnvironmentVariable('Path','%s','User')`, strings.ReplaceAll(newPath, "'", "''"))
	if err := exec.Command("powershell", "-NoProfile", "-Command", psSetPath).Run(); err != nil {
		printError("could not update user PATH: " + err.Error())
		printWarn("add the following to your PATH manually:")
		fmt.Printf("  %s\n", binDir)
		os.Exit(1)
	}

	fmt.Println()
	printSuccess(fmt.Sprintf("%s added to user PATH", binDir))
	printInfo("restart your terminal for the change to take effect")
}

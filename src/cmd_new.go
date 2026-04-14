package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml"
)

func cmdNew(args []string) {
	// parse flags
	var force bool
	var rest []string
	for _, a := range args {
		if a == "-y" || a == "--yes" {
			force = true
		} else {
			rest = append(rest, a)
		}
	}
	args = rest

	cwd, err := os.Getwd()
	if err != nil {
		printError("could not determine current directory: " + err.Error())
		os.Exit(1)
	}

	var projectName, projectRoot string

	if len(args) > 0 {
		projectName = args[0]
		projectRoot = filepath.Join(cwd, projectName)

		if err := os.MkdirAll(projectRoot, 0755); err != nil {
			printError("could not create project directory: " + err.Error())
			os.Exit(1)
		}
		printStep("created", projectName+"/")
	} else {
		projectRoot = cwd
		projectName = filepath.Base(cwd)
		printInfo(fmt.Sprintf("scaffolding in current directory as \"%s\"", projectName))
	}

	if !force {
		if err := checkDirEmpty(projectRoot); err != nil {
			printError(fmt.Sprintf("directory %q is not empty — use -y to overwrite", filepath.Base(projectRoot)))
			os.Exit(1)
		}
	}

	scaffold(projectRoot, projectName)
}

func checkDirEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("not empty")
	}
	return nil
}

func scaffold(root, name string) {
	isWindows := runtime.GOOS == "windows"

	// ── directories ────────────────────────────────────────────────────────────
	dirs := []string{
		filepath.Join(".super", "scripts"),
		"build",
		"lib",
		"src",
	}
	for _, d := range dirs {
		full := filepath.Join(root, d)
		if err := os.MkdirAll(full, 0755); err != nil {
			printError(fmt.Sprintf("could not create %s: %v", d, err))
			os.Exit(1)
		}
		printStep("mkdir", d+string(filepath.Separator))
	}

	// ── src/main.go ────────────────────────────────────────────────────────────
	writeFile(filepath.Join(root, "src", "main.go"), srcMainGo())

	// ── project.settings ───────────────────────────────────────────────────────
	writeFile(filepath.Join(root, "project.settings"), projectSettings(name))

	// ── scripts ────────────────────────────────────────────────────────────────
	scriptsDir := filepath.Join(root, ".super", "scripts")
	if isWindows {
		writeFile(filepath.Join(scriptsDir, "build.ps1"), buildPS1(name))
		writeFile(filepath.Join(scriptsDir, "run.ps1"), runPS1(name))
		writeFile(filepath.Join(scriptsDir, "dev.ps1"), devPS1())
	} else {
		writeFileExec(filepath.Join(scriptsDir, "build.sh"), buildSH(name))
		writeFileExec(filepath.Join(scriptsDir, "run.sh"), runSH(name))
		writeFileExec(filepath.Join(scriptsDir, "dev.sh"), devSH())
	}

	// ── go mod init ────────────────────────────────────────────────────────────
	runGoModInit(root, name)

	fmt.Println()
	printSuccess(fmt.Sprintf("project \"%s\" is ready.", name))
}

// ── file helpers ───────────────────────────────────────────────────────────────

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		printError(fmt.Sprintf("could not write %s: %v", filepath.Base(path), err))
		os.Exit(1)
	}
	printStep("wrote", relativeLast(path))
}

func writeFileExec(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		printError(fmt.Sprintf("could not write %s: %v", filepath.Base(path), err))
		os.Exit(1)
	}
	printStep("wrote", relativeLast(path))
}

func relativeLast(path string) string {
	dir := filepath.Base(filepath.Dir(path))
	base := filepath.Base(path)
	return dir + string(filepath.Separator) + base
}

// ── go mod init ────────────────────────────────────────────────────────────────

func runGoModInit(root, name string) {
	cmd := exec.Command("go", "mod", "init", name)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		printWarn(fmt.Sprintf("go mod init failed: %v", err))
		if len(out) > 0 {
			printWarn(string(out))
		}
		return
	}
	printStep("exec", "go mod init "+name)
}

// ── file content generators ────────────────────────────────────────────────────

func srcMainGo() string {
	return "package main\n\nfunc main() {\n}\n"
}

type projectConfig struct {
	Project projectSection    `toml:"project"`
	Scripts map[string]string `toml:"scripts"`
}

type projectSection struct {
	Name        string `toml:"name"`
	Description string `toml:"description,omitempty"`
}

func projectSettings(name string) string {
	cfg := projectConfig{
		Project: projectSection{Name: name},
		Scripts: map[string]string{
			"build": ".super/scripts",
			"run":   ".super/scripts",
			"dev":   ".super/scripts",
		},
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		printError("could not marshal project.settings: " + err.Error())
		os.Exit(1)
	}
	return string(b)
}

// -- bash scripts ---------------------------------------------------------------

func buildSH(name string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Read project name from project.settings
NAME="%s"

cd "$PROJECT_ROOT"
echo "[super] building $NAME..."
go build -o "build/$NAME" src/main.go
echo "[super] build complete -> build/$NAME"
`, name)
}

func runSH(name string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

NAME="%s"
BINARY="$PROJECT_ROOT/build/$NAME"

if [ ! -f "$BINARY" ]; then
  echo "[super] binary not found, run build first."
  exit 1
fi

exec "$BINARY" "$@"
`, name)
}

func devSH() string {
	return `#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"
exec go run src/main.go "$@"
`
}

// -- PowerShell scripts ---------------------------------------------------------

func buildPS1(name string) string {
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "..\..")

$Name = "%s"

Set-Location $ProjectRoot
Write-Host "[super] building $Name..."
go build -o "build\$Name.exe" src\main.go
Write-Host "[super] build complete -> build\$Name.exe"
`, name)
}

func runPS1(name string) string {
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "..\..")

$Name = "%s"
$Binary = Join-Path $ProjectRoot "build\$Name.exe"

if (-not (Test-Path $Binary)) {
    Write-Host "[super] binary not found, run build first."
    exit 1
}

& $Binary @args
`, name)
}

func devPS1() string {
	return `$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "..\..")

Set-Location $ProjectRoot
go run src\main.go @args
`
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"dxtrity/super/src/config"
	"dxtrity/super/src/util"

	"github.com/pelletier/go-toml"
)

func CmdNew(args []string) {
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
		util.PrintError("could not determine current directory: " + err.Error())
		os.Exit(1)
	}

	var projectName, projectRoot string

	if len(args) > 0 {
		projectName = args[0]
		projectRoot = filepath.Join(cwd, projectName)

		if err := os.MkdirAll(projectRoot, 0755); err != nil {
			util.PrintError("could not create project directory: " + err.Error())
			os.Exit(1)
		}
		util.PrintStep("created", projectName+"/")
	} else {
		projectRoot = cwd
		projectName = filepath.Base(cwd)
		util.PrintInfo(fmt.Sprintf("scaffolding in current directory as \"%s\"", projectName))
	}

	if !force {
		if err := checkDirEmpty(projectRoot); err != nil {
			util.PrintError(fmt.Sprintf("directory %q is not empty — use -y to overwrite", filepath.Base(projectRoot)))
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
			util.PrintError(fmt.Sprintf("could not create %s: %v", d, err))
			os.Exit(1)
		}
		util.PrintStep("mkdir", d+string(filepath.Separator))
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

	// ── .super/version ─────────────────────────────────────────────────────────
	writeFile(filepath.Join(root, ".super", "version"), config.Version+"\n")

	// ── go mod init ────────────────────────────────────────────────────────────
	runGoModInit(root, name)

	fmt.Println()
	util.PrintSuccess(fmt.Sprintf("project \"%s\" is ready.", name))
}

// ── file helpers ───────────────────────────────────────────────────────────────

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		util.PrintError(fmt.Sprintf("could not write %s: %v", filepath.Base(path), err))
		os.Exit(1)
	}
	util.PrintStep("wrote", relativeLast(path))
}

func writeFileExec(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		util.PrintError(fmt.Sprintf("could not write %s: %v", filepath.Base(path), err))
		os.Exit(1)
	}
	util.PrintStep("wrote", relativeLast(path))
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
		util.PrintWarn(fmt.Sprintf("go mod init failed: %v", err))
		if len(out) > 0 {
			util.PrintWarn(string(out))
		}
		return
	}
	util.PrintStep("exec", "go mod init "+name)
}

// ── file content generators ────────────────────────────────────────────────────

func srcMainGo() string {
	return "package main\n\nvar version = \"dev\"\n\nfunc main() {\n}\n"
}

func projectSettings(name string) string {
	cfg := config.ProjectConfig{
		Project: config.ProjectSection{Name: name, Version: "0.1.0", SuperVersion: config.Version},
		Scripts: map[string]string{
			"build": ".super/scripts",
			"run":   ".super/scripts",
			"dev":   ".super/scripts",
		},
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		util.PrintError("could not marshal project.settings: " + err.Error())
		os.Exit(1)
	}
	return string(b)
}

// ── bash scripts ───────────────────────────────────────────────────────────────

func buildSH(name string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

NAME="%s"
SETTINGS="$PROJECT_ROOT/project.settings"

# Auto-increment patch version in project.settings
CURRENT=$(grep -E '^\s+version = ' "$SETTINGS" 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$CURRENT" ]; then
  MAJOR=$(echo "$CURRENT" | cut -d. -f1)
  MINOR=$(echo "$CURRENT" | cut -d. -f2)
  PATCH=$(echo "$CURRENT" | cut -d. -f3)
  NEW_VERSION="$MAJOR.$MINOR.$((PATCH + 1))"
  sed -i.bak "s/^\(  version = \"\)[0-9]*\.[0-9]*\.[0-9]*/\1$NEW_VERSION/" "$SETTINGS" && rm -f "$SETTINGS.bak"
  echo "[super] version: $CURRENT -> $NEW_VERSION"
else
  NEW_VERSION="dev"
fi

cd "$PROJECT_ROOT"
echo "[super] building $NAME..."
go build -ldflags "-X main.version=$NEW_VERSION" -o "build/$NAME" ./src/
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
exec go run ./src/ "$@"
`
}

// ── PowerShell scripts ─────────────────────────────────────────────────────────

func buildPS1(name string) string {
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "..\..")

$Name = "%s"
$Settings = Join-Path $ProjectRoot "project.settings"

# Auto-increment patch version in project.settings
$NewVersion = "dev"
$OldVer = $null
foreach ($line in (Get-Content $Settings)) {
    if ($line -match '^\s+version = "(\d+)\.(\d+)\.(\d+)"') {
        $OldVer = $Matches[1] + '.' + $Matches[2] + '.' + $Matches[3]
        $NewVersion = $Matches[1] + '.' + $Matches[2] + '.' + ([int]$Matches[3] + 1)
        break
    }
}
if ($OldVer) {
    $updated = (Get-Content $Settings) | ForEach-Object {
        if ($_ -match '^\s+version = "\d+\.\d+\.\d+"') {
            $_ -replace [regex]::Escape('"' + $OldVer + '"'), ('"' + $NewVersion + '"')
        } else { $_ }
    }
    Set-Content $Settings $updated
    Write-Host "[super] version: $OldVer -> $NewVersion"
}

Set-Location $ProjectRoot
Write-Host "[super] building $Name..."
go build -ldflags "-X main.version=$NewVersion" -o "build\$Name.exe" .\src\
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
go run .\src\ @args
`
}

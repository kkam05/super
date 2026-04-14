package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func cmdDev(args []string) {
	if len(args) == 0 {
		printError("usage: super dev <subcommand> [flags]")
		fmt.Fprintln(os.Stderr, "  subcommands: release")
		os.Exit(1)
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "release":
		cmdDevRelease(rest)
	default:
		printError(fmt.Sprintf("unknown dev subcommand: %q", sub))
		os.Exit(1)
	}
}

// cmdDevRelease packages the local build binary into a release zip that is
// ready to upload to GitHub Releases.
//
// Flags:
//
//	--local   use the binary from build/ in the current project (required for now)
func cmdDevRelease(args []string) {
	var local bool
	for _, a := range args {
		if a == "--local" {
			local = true
		}
	}

	if !local {
		printError("super dev release requires --local (only local builds are supported)")
		os.Exit(1)
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	// Read project name and version from project.settings.
	settingsPath := filepath.Join(projectRoot, "project.settings")
	cfg, err := loadSettings(settingsPath)
	if err != nil {
		printError("could not read project.settings: " + err.Error())
		os.Exit(1)
	}

	name := cfg.Project.Name
	if name == "" {
		name = filepath.Base(projectRoot)
	}
	ver := cfg.Project.Version
	if ver == "" {
		ver = "unknown"
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// Locate the built binary.
	binarySrc := filepath.Join(projectRoot, "build", name+ext)
	if _, err := os.Stat(binarySrc); err != nil {
		printError(fmt.Sprintf("binary not found at %s — run 'super run build' first", filepath.Join("build", name+ext)))
		os.Exit(1)
	}

	// Create release/ directory.
	releaseDir := filepath.Join(projectRoot, "release")
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		printError("could not create release/ directory: " + err.Error())
		os.Exit(1)
	}

	// Zip name matches what super update expects: <name>-<GOOS>-<GOARCH>.zip
	zipName := fmt.Sprintf("%s-%s-%s.zip", name, runtime.GOOS, runtime.GOARCH)
	zipDst := filepath.Join(releaseDir, zipName)

	if err := zipBinary(binarySrc, zipDst, name+ext); err != nil {
		printError("could not create release zip: " + err.Error())
		os.Exit(1)
	}

	printStep("created", filepath.Join("release", zipName))
	fmt.Println()
	printSuccess(fmt.Sprintf("%s v%s packaged as %s", name, ver, zipName))
	printInfo(fmt.Sprintf("upload %s to your GitHub release tagged v%s", zipName, ver))
}

package main

import (
	"fmt"
	"os"

	"dxtrity/super/src/cmd"
	"dxtrity/super/src/config"
	"dxtrity/super/src/util"
)

var version = "0.1.6"

func main() {
	config.Version = version

	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	command := args[0]
	rest := args[1:]

	switch command {
	case "new":
		cmd.CmdNew(rest)
	case "run":
		cmd.CmdRun(rest)
	case "fix":
		cmd.CmdFix(rest)
	case "update":
		cmd.CmdUpdate(rest)
	case "path":
		cmd.CmdPath(rest)
	case "dev":
		cmd.CmdDev(rest)
	case "version", "--version", "-v":
		fmt.Printf("super v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		util.PrintError(fmt.Sprintf("unknown command: %q", command))
		fmt.Fprintln(os.Stderr)
		printUsage()
		os.Exit(1)
	}
}

const asciiArt = `
 ____  _   _ ____  _____ ____
/ ___|| | | |  _ \| ____|  _ \
\___ \| | | | |_) |  _| | |_) |
 ___) | |_| |  __/| |___|  _ <
|____/ \___/|_|   |_____|_| \_\
`

func printUsage() {
	fmt.Println(util.CyanBold.Sprint(asciiArt))
	fmt.Printf("%s — an npm-style project manager for Go (%s)\n\n", util.Tag(), util.Bold.Sprintf("v%s", version))
	fmt.Printf("%s\n\n", util.Gray.Sprint("Usage:"))
	fmt.Printf("  super %s\n\n", util.Bold.Sprint("<command> [arguments]"))
	fmt.Printf("%s\n\n", util.Gray.Sprint("Commands:"))
	fmt.Printf("  %-28s Scaffold a new Go project.\n", util.Bold.Sprint("new")+" [project-name] [-y]")
	fmt.Printf("  %-28s If project-name is omitted, uses the current directory.\n", "")
	fmt.Printf("  %-28s Use -y to overwrite a non-empty directory.\n", "")
	fmt.Printf("  %-28s Run a script defined in project.settings.\n", util.Bold.Sprint("run")+" <script> [args...]")
	fmt.Printf("  %-28s Scripts can be paths (.super/scripts) or inline commands.\n", "")
	fmt.Printf("  %-28s Repair a project to match the expected super structure.\n", util.Bold.Sprint("fix"))
	fmt.Printf("  %-28s Ensures dirs, project.settings, scripts, and version file are correct.\n", "")
	fmt.Printf("  %-28s Pull and install the latest release from GitHub.\n", util.Bold.Sprint("update"))
	fmt.Printf("  %-28s Use --local to install from build/super instead.\n", "")
	fmt.Printf("  %-28s Check PATH and add ~/.super/bin if missing.\n", util.Bold.Sprint("path"))
	fmt.Printf("  %-28s Developer utilities (e.g. packaging releases).\n", util.Bold.Sprint("dev")+" <subcommand>")
	fmt.Printf("  %-28s Package the local build into a release zip.\n", "  "+util.Bold.Sprint("release")+" --local")
	fmt.Printf("  %-28s Print the super version.\n", util.Bold.Sprint("version"))
	fmt.Printf("  %-28s Show this help message.\n\n", util.Bold.Sprint("help"))
}

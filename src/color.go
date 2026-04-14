package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	cyanBold = color.New(color.FgCyan, color.Bold)
	green    = color.New(color.FgGreen)
	red      = color.New(color.FgRed)
	yellow   = color.New(color.FgYellow)
	gray     = color.New(color.FgHiBlack)
	bold     = color.New(color.Bold)
)

func tag() string {
	return cyanBold.Sprint("[super]")
}

func printInfo(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", tag(), msg)
}

func printStep(action, target string) {
	fmt.Fprintf(os.Stdout, "%s %s %s\n", tag(), gray.Sprintf("%-10s", action), target)
}

func printSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", tag(), green.Sprint(msg))
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", tag(), red.Sprintf("error: %s", msg))
}

func printWarn(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", tag(), yellow.Sprintf("warn: %s", msg))
}

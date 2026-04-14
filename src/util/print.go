package util

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	CyanBold = color.New(color.FgCyan, color.Bold)
	Green    = color.New(color.FgGreen)
	Red      = color.New(color.FgRed)
	Yellow   = color.New(color.FgYellow)
	Gray     = color.New(color.FgHiBlack)
	Bold     = color.New(color.Bold)
)

func Tag() string {
	return CyanBold.Sprint("[super]")
}

func PrintInfo(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Tag(), msg)
}

func PrintStep(action, target string) {
	fmt.Fprintf(os.Stdout, "%s %s %s\n", Tag(), Gray.Sprintf("%-10s", action), target)
}

func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Tag(), Green.Sprint(msg))
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Tag(), Red.Sprintf("error: %s", msg))
}

func PrintWarn(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Tag(), Yellow.Sprintf("warn: %s", msg))
}

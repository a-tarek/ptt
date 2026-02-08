package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
)

// FormatError formats an error message with color and help footer
// The cmd parameter is the subcommand name (e.g., "goto", "delete")
// When cmd is empty, the help footer is omitted
func FormatError(err error, cmd string) string {
	if err == nil {
		return ""
	}

	msg := red.Sprintf("Error: %s", err.Error())

	if cmd != "" {
		msg += fmt.Sprintf("\nRun 'wt help %s' for details", cmd)
	}

	return msg
}

// Warn prints a warning message in yellow to stderr
func Warn(msg string) string {
	return yellow.Sprintf("Warning: %s", msg)
}

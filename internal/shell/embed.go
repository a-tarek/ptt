package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed templates/wrapper.bash
var wrapperBash string

//go:embed templates/wrapper.zsh
var wrapperZsh string

//go:embed templates/wrapper.fish
var wrapperFish string

const binPlaceholder = "__PTT_BIN__"

// GetWrapper returns the embedded wrapper script for the given shell,
// with the binary path injected.
// Supported shells: bash, zsh, fish.
func GetWrapper(shellName string, binaryPath string) (string, error) {
	var template string
	switch shellName {
	case "bash":
		template = wrapperBash
	case "zsh":
		template = wrapperZsh
	case "fish":
		template = wrapperFish
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shellName)
	}
	return strings.ReplaceAll(template, binPlaceholder, binaryPath), nil
}

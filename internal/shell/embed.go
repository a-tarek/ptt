package shell

import (
	_ "embed"
	"fmt"
)

//go:embed templates/wrapper.bash
var wrapperBash string

//go:embed templates/wrapper.zsh
var wrapperZsh string

//go:embed templates/wrapper.fish
var wrapperFish string

// GetWrapper returns the embedded wrapper script for the given shell.
// Supported shells: bash, zsh, fish.
func GetWrapper(shellName string) (string, error) {
	switch shellName {
	case "bash":
		return wrapperBash, nil
	case "zsh":
		return wrapperZsh, nil
	case "fish":
		return wrapperFish, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shellName)
	}
}

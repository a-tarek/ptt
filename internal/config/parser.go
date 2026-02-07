package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseFile reads a config file and returns a slice of actions
func ParseFile(path string) ([]Action, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var actions []Action
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Split into action type and path (use SplitN to preserve spaces in run commands)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line %d: missing space separator", lineNum)
		}

		actionType := strings.TrimSpace(parts[0])
		actionPath := strings.TrimSpace(parts[1])

		actions = append(actions, Action{
			Type: actionType,
			Path: actionPath,
			Line: lineNum,
		})
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return actions, nil
}

package installer

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	// MarkerBegin is the start marker for the wt configuration block
	MarkerBegin = "# >>> wt >>>"
	// MarkerEnd is the end marker for the wt configuration block
	MarkerEnd = "# <<< wt <<<"
)

var (
	// v1SourcePattern matches lines that source wt.zsh
	// Matches patterns like: source ~/.oh-my-zsh/custom/plugins/wt/wt.zsh
	// or: source /path/to/wt.zsh
	v1SourcePattern = regexp.MustCompile(`^\s*source\s+.*wt\.zsh`)
	// v1DotPattern matches lines that use dot-source for wt.zsh
	// Matches patterns like: . ~/wt.zsh or . /path/to/wt.zsh
	v1DotPattern = regexp.MustCompile(`^\s*\.\s+.*wt\.zsh`)
)

// EvalLine returns the eval line for a given shell type.
func EvalLine(shellType string) string {
	switch shellType {
	case "fish":
		return `wt shell-init | source`
	default: // bash, zsh
		return `eval "$(wt shell-init)"`
	}
}

// MarkerBlock returns the full marker block string for a given shell type.
func MarkerBlock(shellType string) string {
	return fmt.Sprintf(`%s
# Managed by wt installer — do not edit this block manually
%s
%s
`, MarkerBegin, EvalLine(shellType), MarkerEnd)
}

// HasMarkerBlock returns true if the content contains the marker block.
func HasMarkerBlock(content string) bool {
	return strings.Contains(content, MarkerBegin)
}

// FindV1Lines returns 0-indexed line numbers that appear to be wt v1 entries.
// It scans each line and returns indices of non-commented lines that match
// v1 source patterns (source wt.zsh or . wt.zsh).
func FindV1Lines(content string) []int {
	lines := strings.Split(content, "\n")
	var indices []int

	for i, line := range lines {
		// Skip commented lines
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for v1 patterns
		if v1SourcePattern.MatchString(line) || v1DotPattern.MatchString(line) {
			indices = append(indices, i)
		}
	}

	return indices
}

// CommentOutLines returns a copy of lines with the given indices commented out.
// Each line gets "# [wt v2 migration] " prepended.
func CommentOutLines(lines []string, indices []int) []string {
	// Create a copy
	result := make([]string, len(lines))
	copy(result, lines)

	// Create a set of indices for O(1) lookup
	indexSet := make(map[int]bool)
	for _, idx := range indices {
		indexSet[idx] = true
	}

	// Comment out the specified lines
	for i := range result {
		if indexSet[i] {
			result[i] = "# [wt v2 migration] " + result[i]
		}
	}

	return result
}

// InsertMarkerBlock adds the marker block to the end of content.
func InsertMarkerBlock(content string, shellType string) string {
	// Ensure content ends with newline
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Append marker block
	return content + MarkerBlock(shellType)
}

// RemoveMarkerBlock removes everything from MarkerBegin through MarkerEnd (inclusive).
// Returns content unchanged if marker block is not found.
// Leaves a blank line in place of the removed block to maintain file structure.
func RemoveMarkerBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false
	foundBlock := false

	for _, line := range lines {
		if strings.Contains(line, MarkerBegin) {
			inBlock = true
			foundBlock = true
			// Add blank line in place of the block
			result = append(result, "")
			continue
		}
		if strings.Contains(line, MarkerEnd) {
			inBlock = false
			continue
		}
		if !inBlock {
			result = append(result, line)
		}
	}

	// If no block was found, return content unchanged
	if !foundBlock {
		return content
	}

	return strings.Join(result, "\n")
}

// BackupFile copies filePath to filePath + ".wt-backup".
// Returns the backup path on success.
func BackupFile(filePath string) (string, error) {
	backupPath := filePath + ".wt-backup"

	// Read original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for backup: %w", err)
	}

	// Get original file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file for backup: %w", err)
	}

	// Write backup with same permissions
	if err := os.WriteFile(backupPath, content, info.Mode()); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// RestoreBackup copies backup back to original path and removes backup file.
func RestoreBackup(backupPath, originalPath string) error {
	// Read backup
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Get backup file permissions
	info, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("failed to stat backup file: %w", err)
	}

	// Restore original
	if err := os.WriteFile(originalPath, content, info.Mode()); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Remove backup
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}

	return nil
}

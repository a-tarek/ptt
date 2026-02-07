package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahmedelarabyy/wt/internal/installer"
	"github.com/ahmedelarabyy/wt/internal/shell"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up wt shell integration",
	Long: `Set up wt shell integration by adding the wt configuration block to your shell's RC file.

This command will:
  • Detect your current shell
  • Check for existing wt installations
  • Show you exactly what will be added
  • Migrate from wt v1 if needed
  • Safely modify your RC file with backup

The installation is idempotent - you can run it multiple times safely.`,
	Args: cobra.NoArgs,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("wt install — Shell Integration Setup")
	fmt.Println("=====================================")
	fmt.Println()

	// Step 1: Detect shell
	fmt.Println("Step 1: Detecting shell...")
	shellType, err := shell.DetectShell()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "To set up manually, add this to your shell's RC file:\n\n")
		fmt.Fprintf(os.Stderr, "  eval \"$(wt shell-init)\"\n\n")
		return err
	}
	fmt.Printf("  Detected: %s\n\n", shellType)

	fmt.Printf("Continue with %s? [Y/n] ", shellType)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "n" || response == "no" {
		fmt.Println("\nNo changes made. To set up manually, add this to your RC file:")
		fmt.Printf("\n  %s\n\n", installer.EvalLine(shellType))
		return nil
	}

	// Step 2: Check RC file
	fmt.Println("Step 2: Checking configuration...")
	rcFilePath, err := installer.RCFilePath(shellType)
	if err != nil {
		return fmt.Errorf("failed to determine RC file path: %w", err)
	}

	// Display path with ~ for home directory
	displayPath := rcFilePath
	if homeDir, err := os.UserHomeDir(); err == nil {
		displayPath = strings.Replace(rcFilePath, homeDir, "~", 1)
	}
	fmt.Printf("  RC file: %s\n\n", displayPath)

	// Read existing content (or empty if file doesn't exist)
	var content string
	if data, err := os.ReadFile(rcFilePath); err == nil {
		content = string(data)
	}

	// Step 3: Check for existing installation
	if installer.HasMarkerBlock(content) {
		fmt.Printf("wt is already configured in %s\n", displayPath)
		fmt.Println("No changes needed.")
		return nil
	}

	// Step 4: Detect v1 entries
	v1Lines := installer.FindV1Lines(content)
	if len(v1Lines) > 0 {
		fmt.Println("Step 3: Migrating from wt v1...")
		lines := strings.Split(content, "\n")
		for _, lineNum := range v1Lines {
			if lineNum < len(lines) {
				fmt.Printf("  Found v1 entry on line %d: %s\n", lineNum+1, lines[lineNum])
			}
		}
		fmt.Println("  This will be commented out and replaced with the v2 configuration.")
		fmt.Println()
	}

	// Step 5: Show changes
	stepNum := 4
	if len(v1Lines) > 0 {
		stepNum = 4
	} else {
		stepNum = 3
	}
	fmt.Printf("Step %d: The following will be added to %s:\n\n", stepNum, displayPath)

	// Show the marker block with indentation
	blockLines := strings.Split(strings.TrimSuffix(installer.MarkerBlock(shellType), "\n"), "\n")
	for _, line := range blockLines {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println()

	if len(v1Lines) > 0 {
		fmt.Println("  The following line(s) will be commented out:")
		lines := strings.Split(content, "\n")
		for _, lineNum := range v1Lines {
			if lineNum < len(lines) {
				fmt.Printf("    Line %d: %s\n", lineNum+1, lines[lineNum])
			}
		}
		fmt.Println()
	}

	fmt.Print("Proceed? [Y/n] ")
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "n" || response == "no" {
		fmt.Printf("\nNo changes made. To set up manually, add this to your %s:\n\n", displayPath)
		fmt.Printf("  %s\n\n", installer.EvalLine(shellType))
		return nil
	}

	// Step 6: Apply changes
	fmt.Printf("\nStep %d: Applying changes...\n", stepNum+1)

	// Backup existing file if it exists
	var backupPath string
	if _, err := os.Stat(rcFilePath); err == nil {
		backupPath, err = installer.BackupFile(rcFilePath)
		if err != nil {
			return fmt.Errorf("failed to backup RC file: %w", err)
		}
		fmt.Printf("  Backing up %s → %s\n", displayPath, filepath.Base(backupPath))
	}

	// Apply v1 migration if needed
	if len(v1Lines) > 0 {
		lines := strings.Split(content, "\n")
		lines = installer.CommentOutLines(lines, v1Lines)
		content = strings.Join(lines, "\n")
	}

	// Insert marker block
	content = installer.InsertMarkerBlock(content, shellType)

	// Write the new content
	if err := os.WriteFile(rcFilePath, []byte(content), 0644); err != nil {
		// Rollback on failure
		if backupPath != "" {
			_ = installer.RestoreBackup(backupPath, rcFilePath)
		}
		return fmt.Errorf("failed to write RC file: %w", err)
	}

	// Remove backup on success (keep it if user wants to revert manually)
	// Actually, keep the backup - users might want it
	// if backupPath != "" {
	// 	_ = os.Remove(backupPath)
	// }

	// Step 7: Success
	fmt.Println()
	fmt.Printf("Done! wt has been configured for %s.\n\n", shellType)
	fmt.Println("To activate, either:")
	fmt.Println("  • Restart your terminal")

	sourceCmd := "source " + displayPath
	if shellType == "fish" {
		sourceCmd = "source " + displayPath
	}
	fmt.Printf("  • Run: %s\n", sourceCmd)

	return nil
}

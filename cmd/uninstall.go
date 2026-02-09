package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/a-tarek/ptt/internal/installer"
	"github.com/a-tarek/ptt/internal/shell"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove ptt shell integration",
	Long: `Remove ptt shell integration by removing the ptt configuration block from your shell's RC file.

This command will:
  • Detect your current shell
  • Check for existing ptt installation
  • Show you exactly what will be removed
  • Safely modify your RC file with backup
  • Provide instructions for complete uninstallation

Note: This only removes the RC file integration. To fully uninstall ptt,
you'll need to also run: npm uninstall -g @atarek/ptt`,
	Args: cobra.NoArgs,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("ptt uninstall — Remove Shell Integration")
	fmt.Println("=========================================")
	fmt.Println()

	// Step 1: Detect shell and rc file
	fmt.Print("Detecting shell... ")
	shellType, err := shell.DetectShell()
	if err != nil {
		return fmt.Errorf("failed to detect shell: %w", err)
	}
	fmt.Println(shellType)

	rcFilePath, err := installer.RCFilePath(shellType)
	if err != nil {
		return fmt.Errorf("failed to determine RC file path: %w", err)
	}

	// Display path with ~ for home directory
	displayPath := rcFilePath
	if homeDir, err := os.UserHomeDir(); err == nil {
		displayPath = strings.Replace(rcFilePath, homeDir, "~", 1)
	}
	fmt.Printf("RC file: %s\n", displayPath)
	fmt.Println()

	// Read rc file content
	data, err := os.ReadFile(rcFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("RC file not found: %s\n", displayPath)
			fmt.Println("No changes needed.")
			return nil
		}
		return fmt.Errorf("failed to read RC file: %w", err)
	}
	content := string(data)

	// Step 2: Check for existing installation
	if !installer.HasMarkerBlock(content) {
		fmt.Printf("ptt is not configured in %s\n", displayPath)
		fmt.Println("No changes needed.")
		return nil
	}

	// Step 3: Show what will be removed
	fmt.Printf("The following block will be removed from %s:\n\n", displayPath)

	// Extract and display the marker block
	blockLines := strings.Split(strings.TrimSuffix(installer.MarkerBlock(shellType), "\n"), "\n")
	for _, line := range blockLines {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println()

	fmt.Print("Proceed? [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "n" || response == "no" {
		fmt.Println("\nNo changes made.")
		return nil
	}

	// Step 4: Remove marker block
	fmt.Println()
	backupPath, err := installer.BackupFile(rcFilePath)
	if err != nil {
		return fmt.Errorf("failed to backup RC file: %w", err)
	}
	fmt.Printf("  Backing up %s → %s\n", displayPath, displayPath+".wt-backup")

	newContent := installer.RemoveMarkerBlock(content)

	if err := os.WriteFile(rcFilePath, []byte(newContent), 0644); err != nil {
		// Rollback on failure
		_ = installer.RestoreBackup(backupPath, rcFilePath)
		return fmt.Errorf("failed to write RC file: %w", err)
	}

	// Step 5: Success
	fmt.Println()
	fmt.Println("Done! ptt shell integration has been removed.")
	fmt.Println()
	fmt.Println("To complete uninstallation:")
	fmt.Println("  1. Restart your terminal (or run: source " + displayPath + ")")
	fmt.Println("  2. Run: npm uninstall -g @atarek/ptt")

	return nil
}

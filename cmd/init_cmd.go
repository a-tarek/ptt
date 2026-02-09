package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

const pttconfigTemplate = `# pttconfig — actions to run when creating new worktrees
#
# Actions:
#   copy <path>       Copy file or directory from source worktree
#   symlink <path>    Symlink to source worktree's file or directory
#   run <command>     Run a shell command in the new worktree
#
# Examples:
#
# copy .env.local
# copy .env
# symlink node_modules
# symlink .venv
# symlink target
# run npm install
`

var (
	initName string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create config template",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Must be inside git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Get config root (bare repo root or home worktree)
		repoRoot, err := git.ConfigRoot()
		if err != nil {
			return fmt.Errorf("failed to get repo root: %w", err)
		}

		// Create .pttconfig directory idempotently
		pttconfigDir := filepath.Join(repoRoot, ".pttconfig")
		if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create .pttconfig directory: %w", err)
		}

		// Determine config filename
		var configName string
		if initName != "" {
			configName = initName
		} else {
			configName = "default"
		}

		// Config path inside .pttconfig directory
		configPath := filepath.Join(pttconfigDir, configName)

		// Error if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf(".pttconfig/%s already exists", configName)
		}

		// Write template
		if err := os.WriteFile(configPath, []byte(pttconfigTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create .pttconfig/%s: %w", configName, err)
		}

		// Silent on success
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initName, "config", "", "create named config (.pttconfig/<name>)")
}

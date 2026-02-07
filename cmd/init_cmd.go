package cmd

import (
	"fmt"
	"os"

	"github.com/ahmedelarabyy/wt/internal/git"
	"github.com/spf13/cobra"
)

const wtconfigTemplate = `# .wtconfig — actions to run when creating new worktrees
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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .wtconfig template",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Must be inside git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Config path in current directory
		configPath := cwd + "/.wtconfig"

		// Error if .wtconfig already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf(".wtconfig already exists")
		}

		// Write template
		if err := os.WriteFile(configPath, []byte(wtconfigTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create .wtconfig: %w", err)
		}

		// Silent on success
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

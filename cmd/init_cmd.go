package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-tarek/ptt/internal/git"
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

var (
	initName string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .wtconfig template",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Must be inside git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Get repo root
		repoRoot, err := git.CurrentWorktreeRoot()
		if err != nil {
			return fmt.Errorf("failed to get repo root: %w", err)
		}

		// Determine config filename
		var configFilename string
		if initName != "" {
			configFilename = ".wtconfig-" + initName
		} else {
			configFilename = ".wtconfig"
		}

		// Config path at repo root
		configPath := filepath.Join(repoRoot, configFilename)

		// Error if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("%s already exists", configFilename)
		}

		// Write template
		if err := os.WriteFile(configPath, []byte(wtconfigTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", configFilename, err)
		}

		// Silent on success
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initName, "name", "", "create named config variant (.wtconfig-{name})")
}

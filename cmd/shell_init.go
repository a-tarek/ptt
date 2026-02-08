package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ahmedelarabyy/wt/internal/shell"
	"github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
	Use:    "shell-init",
	Hidden: true,
	Short:  "Output shell wrapper function for eval",
	Long:   "Outputs shell-specific wrapper code. Add to your rc file:\n  eval $(wt shell-init)",
	Args:   cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellType, err := shell.DetectShell()
		if err != nil {
			return err
		}

		// Resolve absolute path to this binary
		binPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to resolve binary path: %w", err)
		}
		binPath, err = filepath.EvalSymlinks(binPath)
		if err != nil {
			return fmt.Errorf("failed to resolve binary symlinks: %w", err)
		}

		wrapper, err := shell.GetWrapper(shellType, binPath)
		if err != nil {
			return err
		}
		fmt.Print(wrapper)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}

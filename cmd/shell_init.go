package cmd

import (
	"fmt"

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
		wrapper, err := shell.GetWrapper(shellType)
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

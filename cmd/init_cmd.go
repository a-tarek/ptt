package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	initcmd "github.com/a-tarek/ptt/internal/init"
)

var autoYes bool

var green = color.New(color.FgGreen)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize repository for ptt",
	Long: `Initialize the current repository for ptt management.

Detects the repository type and performs appropriate setup:
  - Standard clone: restructures into ptt bare layout with nested worktrees
  - Raw bare repo: adopts into ptt layout with .bare/ wrapper
  - ptt-managed repo: creates config or repairs issues

Use -y to skip confirmation prompt.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Detect repo type
		repoType, info, err := initcmd.DetectRepoType()
		if err != nil {
			return err
		}

		// 2. Handle not-git early exit
		if repoType == initcmd.RepoTypeNotGit {
			return fmt.Errorf("Not inside a git repository. Run 'git init' or 'git clone <url>' first.")
		}

		// 3. Handle ptt-bare-from-worktree redirect
		if info.IsBareFromWorktree {
			fmt.Fprintf(os.Stderr, "Detected ptt-managed repo. Redirecting to container root: %s\n", info.ContainerRoot)
			if err := os.Chdir(info.ContainerRoot); err != nil {
				return fmt.Errorf("failed to change directory to container root: %w", err)
			}
			// Re-detect after changing directory
			repoType, info, err = initcmd.DetectRepoType()
			if err != nil {
				return err
			}
		}

		// 4. Validate
		validation, err := initcmd.ValidateForInit(repoType, info)
		if err != nil {
			return err
		}
		if len(validation.Errors) > 0 {
			for _, errMsg := range validation.Errors {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
			}
			return fmt.Errorf("validation failed")
		}

		// 5. Build and show plan
		actions := initcmd.BuildActionPlan(repoType, info, validation)
		initcmd.ShowPlan(repoType, info, validation, actions)

		// 6. Confirm
		confirmed, err := initcmd.AskConfirmation(autoYes)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		// 7. Execute based on repo type
		progress := func(step string) {
			fmt.Fprintf(os.Stderr, "  %s %s\n", green.Sprint("✓"), step)
		}

		switch repoType {
		case initcmd.RepoTypeNormal:
			if err := initcmd.RestructureNormalRepo(info.RepoRoot, info, progress); err != nil {
				return err
			}
		case initcmd.RepoTypeRawBare:
			if err := initcmd.AdoptRawBareRepo(info.RepoRoot, info, progress); err != nil {
				return err
			}
		case initcmd.RepoTypePttBare:
			if err := initcmd.RepairPttRepo(info.ContainerRoot, validation.RepairItems, progress); err != nil {
				return err
			}
		}

		// 8. Show completion output
		fmt.Fprintf(os.Stderr, "\n%s\n", green.Sprint("Done!"))

		if repoType == initcmd.RepoTypeNormal || repoType == initcmd.RepoTypeRawBare {
			fmt.Fprintf(os.Stderr, "\nRun: ptt cd %s\n", info.DefaultBranch)
		}

		if !info.HasRemote {
			fmt.Fprintf(os.Stderr, "\nNo remote configured. To add one:\n")
			fmt.Fprintf(os.Stderr, "  git remote add origin <url>\n")
			fmt.Fprintf(os.Stderr, "  git push -u origin %s\n", info.DefaultBranch)
			fmt.Fprintf(os.Stderr, "  git fetch origin\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "auto-accept confirmation")
}

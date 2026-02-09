package initcmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
	cyan   = color.New(color.FgCyan)
	bold   = color.New(color.Bold)
)

// BuildActionPlan generates a list of actions based on repo type and validation
func BuildActionPlan(repoType RepoType, info *RepoInfo, validation *ValidationResult) []string {
	actions := []string{}

	switch repoType {
	case RepoTypeNotGit:
		actions = append(actions, "Cannot initialize - not in a git repository")

	case RepoTypeNormal:
		if len(validation.Errors) == 0 {
			actions = append(actions, "Clone repository to new bare structure")
			actions = append(actions, "Create .bare/ directory with bare clone")
			actions = append(actions, "Create .git pointer file")
			actions = append(actions, "Configure fetch refspec and reflog")
			actions = append(actions, fmt.Sprintf("Create initial worktree for '%s'", info.DefaultBranch))
			if info.HasRemote {
				actions = append(actions, "Fetch from remote origin")
			}
		}

	case RepoTypeRawBare:
		if len(validation.Errors) == 0 {
			actions = append(actions, "Adopt existing bare repository")
			actions = append(actions, "Create .git pointer file")
			actions = append(actions, "Configure fetch refspec and reflog")
			actions = append(actions, fmt.Sprintf("Create initial worktree for '%s'", info.DefaultBranch))
			if info.HasRemote {
				actions = append(actions, "Fetch from remote origin")
			}
		}

	case RepoTypePttBare:
		if len(validation.RepairItems) > 0 {
			actions = append(actions, "Repair existing ptt repository:")
			for _, item := range validation.RepairItems {
				actions = append(actions, fmt.Sprintf("  - Fix: %s", item))
			}
		} else {
			actions = append(actions, "Repository is already ptt-managed and healthy")
		}
	}

	return actions
}

// ShowPlan displays the plan to the user
func ShowPlan(repoType RepoType, info *RepoInfo, validation *ValidationResult, actions []string) {
	// Header
	bold.Fprintf(os.Stderr, "\n=== Repository Analysis ===\n\n")

	// Repo type
	fmt.Fprintf(os.Stderr, "Type: %s\n", cyan.Sprint(repoType.String()))

	if info.RepoRoot != "" {
		fmt.Fprintf(os.Stderr, "Root: %s\n", info.RepoRoot)
	}
	if info.ContainerRoot != "" {
		fmt.Fprintf(os.Stderr, "Container: %s\n", info.ContainerRoot)
	}
	if info.DefaultBranch != "" {
		fmt.Fprintf(os.Stderr, "Default branch: %s\n", info.DefaultBranch)
	}
	if info.HasRemote {
		fmt.Fprintf(os.Stderr, "Remote: %s\n", info.RemoteURL)
	} else {
		fmt.Fprintf(os.Stderr, "Remote: %s\n", yellow.Sprint("none"))
	}

	// Existing worktrees
	if len(info.ExistingWorktrees) > 0 {
		fmt.Fprintf(os.Stderr, "\nExisting worktrees:\n")
		for _, wt := range info.ExistingWorktrees {
			if wt.Branch != "" {
				fmt.Fprintf(os.Stderr, "  - %s (%s)\n", wt.Path, wt.Branch)
			} else {
				fmt.Fprintf(os.Stderr, "  - %s (detached)\n", wt.Path)
			}
		}
	}

	// Errors
	if len(validation.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		bold.Fprintf(os.Stderr, "Errors:\n")
		for _, err := range validation.Errors {
			fmt.Fprintf(os.Stderr, "  %s %s\n", red.Sprint("✗"), err)
		}
	}

	// Warnings
	if len(validation.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		bold.Fprintf(os.Stderr, "Warnings:\n")
		for _, warn := range validation.Warnings {
			fmt.Fprintf(os.Stderr, "  %s %s\n", yellow.Sprint("⚠"), warn)
		}
	}

	// Repair items
	if len(validation.RepairItems) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		bold.Fprintf(os.Stderr, "Issues detected:\n")
		for _, item := range validation.RepairItems {
			fmt.Fprintf(os.Stderr, "  %s %s\n", yellow.Sprint("⚠"), item)
		}
	}

	// Actions
	if len(actions) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		bold.Fprintf(os.Stderr, "Plan:\n")
		for _, action := range actions {
			if strings.HasPrefix(action, "  -") {
				// Indented repair action
				fmt.Fprintf(os.Stderr, "  %s\n", action)
			} else {
				fmt.Fprintf(os.Stderr, "  • %s\n", action)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
}

// AskConfirmation asks the user to confirm the plan
func AskConfirmation(autoYes bool) (bool, error) {
	if autoYes {
		return true, nil
	}

	fmt.Fprintf(os.Stderr, "Proceed? (yes/no): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false, fmt.Errorf("failed to read input")
	}

	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "yes" || response == "y", nil
}

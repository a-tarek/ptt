package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

// ANSI color codes
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiCyan    = "\033[36m"
	ansiGray    = "\033[90m"
)

var promptCmd = &cobra.Command{
	Use:    "prompt",
	Short:  "Output worktree-aware prompt segment",
	Long:   "Outputs a formatted prompt segment with potato emoji, branch info, and git status.\nAt bare repo root: 🥔 root (gray).\nInside a worktree: 🥔 [branch status].\nOutside a ptt repo: prints nothing (exit 1).",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE:   runPrompt,
}

func init() {
	rootCmd.AddCommand(promptCmd)
}

func runPrompt(cmd *cobra.Command, args []string) error {
	bareRoot, err := git.BareRepoRoot()
	if err != nil {
		return fmt.Errorf("not a ptt repo")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("not a ptt repo")
	}

	// Resolve symlinks for reliable comparison
	cwd, _ = filepath.EvalSymlinks(cwd)
	bareRoot, _ = filepath.EvalSymlinks(bareRoot)

	if cwd == bareRoot {
		fmt.Printf("🥔 %sroot%s", ansiGray, ansiReset)
		return nil
	}

	status, _ := git.Status(cwd)

	branch := status.Branch
	if branch == "" {
		branch = filepath.Base(cwd)
	}

	var parts []string

	if status.Staged > 0 {
		parts = append(parts, fmt.Sprintf("%s+%d%s", ansiGreen, status.Staged, ansiReset))
	}
	if status.Modified > 0 {
		parts = append(parts, fmt.Sprintf("%s~%d%s", ansiYellow, status.Modified, ansiReset))
	}
	if status.Untracked > 0 {
		parts = append(parts, fmt.Sprintf("%s?%d%s", ansiGray, status.Untracked, ansiReset))
	}
	if status.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("%s↑%d%s", ansiGreen, status.Ahead, ansiReset))
	}
	if status.Behind > 0 {
		parts = append(parts, fmt.Sprintf("%s↓%d%s", ansiRed, status.Behind, ansiReset))
	}

	statusStr := ""
	if len(parts) > 0 {
		statusStr = " " + strings.Join(parts, " ")
	}

	fmt.Printf("🥔 %s[%s%s%s%s%s]%s", ansiDim, ansiReset, ansiBold+ansiCyan, branch, ansiReset, statusStr+ansiDim, ansiReset)
	return nil
}

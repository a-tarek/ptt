package initcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepairPttRepo repairs a ptt-managed bare repository based on detected issues.
// This is used when a ptt repo exists but has configuration problems.
//
// Repair items can include:
// - "fix-git-pointer": Rewrite .git pointer file
// - "fix-refspec": Configure fetch refspec
// - "fix-reflog": Enable reflog
// - "create-pttconfig": Create .pttconfig/default if missing
func RepairPttRepo(containerRoot string, repairItems []string, progress ProgressCallback) error {
	// If no repair items, check if we just need to create .pttconfig
	if len(repairItems) == 0 {
		pttconfigDefault := filepath.Join(containerRoot, ".pttconfig", "default")
		if _, err := os.Stat(pttconfigDefault); os.IsNotExist(err) {
			repairItems = []string{"create-pttconfig"}
		} else {
			// Nothing to do
			return nil
		}
	}

	for _, item := range repairItems {
		// Map human-readable repair items to actions
		// Items come from validate.go detectRepairItems()

		if strings.Contains(item, ".git pointer file") {
			progress("Fixing .git pointer file")
			gitPointer := filepath.Join(containerRoot, ".git")
			if err := os.WriteFile(gitPointer, []byte("gitdir: ./.bare\n"), 0644); err != nil {
				return fmt.Errorf("failed to fix .git pointer file: %w", err)
			}

		} else if strings.Contains(item, "fetch refspec") {
			progress("Configuring fetch refspec")
			cmd := exec.Command("git", "-C", containerRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
			if err := cmd.Run(); err != nil {
				// Check if remote exists first
				checkCmd := exec.Command("git", "-C", containerRoot, "remote", "get-url", "origin")
				if checkErr := checkCmd.Run(); checkErr != nil {
					// No remote, skip this repair item
					progress("No remote origin found, skipping refspec configuration")
					continue
				}
				return fmt.Errorf("failed to configure fetch refspec: %w", err)
			}

		} else if strings.Contains(item, "reflog") {
			progress("Enabling reflog")
			cmd := exec.Command("git", "-C", containerRoot, "config", "core.logallrefupdates", "true")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to enable reflog: %w", err)
			}

		} else if strings.Contains(item, ".bare/ directory") {
			// Critical issue - can't repair a missing .bare directory
			return fmt.Errorf("cannot repair: .bare/ directory is missing (repo structure is corrupted)")

		} else if strings.Contains(item, "pttconfig") || item == "create-pttconfig" {
			// Create .pttconfig/default
			progress("Creating .pttconfig")
			pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
			if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
				return fmt.Errorf("failed to create .pttconfig directory: %w", err)
			}
			defaultConfig := filepath.Join(pttconfigDir, "default")
			if err := os.WriteFile(defaultConfig, []byte{}, 0644); err != nil {
				return fmt.Errorf("failed to create .pttconfig/default: %w", err)
			}

		} else {
			// Unknown repair item - log but don't fail
			progress(fmt.Sprintf("Warning: unknown repair item: %s", item))
		}
	}

	return nil
}

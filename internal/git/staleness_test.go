package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpstreamStatus_String(t *testing.T) {
	tests := []struct {
		s    UpstreamStatus
		want string
	}{
		{UpstreamActive, "active"},
		{UpstreamGone, "gone"},
		{UpstreamNone, "no-upstream"},
		{UpstreamUnknown, "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("UpstreamStatus(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestHeadModTime_RegularRepo(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}

	want := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(headPath, want, want); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	got, err := HeadModTime(dir)
	if err != nil {
		t.Fatalf("HeadModTime: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHeadModTime_LinkedWorktree(t *testing.T) {
	dir := t.TempDir()

	// Simulate a linked worktree: .git is a file pointing to gitdir/.
	wtGitdir := filepath.Join(dir, "real-gitdir")
	if err := os.Mkdir(wtGitdir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	headPath := filepath.Join(wtGitdir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/feat\n"), 0644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}

	gitFile := filepath.Join(dir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: "+wtGitdir+"\n"), 0644); err != nil {
		t.Fatalf("write .git pointer: %v", err)
	}

	want := time.Now().Add(-3 * 24 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(headPath, want, want); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	got, err := HeadModTime(dir)
	if err != nil {
		t.Fatalf("HeadModTime: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHeadModTime_Missing(t *testing.T) {
	if _, err := HeadModTime(t.TempDir()); err == nil {
		t.Error("expected error for path without .git")
	}
}

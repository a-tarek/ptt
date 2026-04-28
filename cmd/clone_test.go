package cmd

import "testing"

func TestDeriveCloneName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://github.com/foo/bar.git", "bar"},
		{"https://github.com/foo/bar", "bar"},
		{"git@github.com:foo/bar.git", "bar"},
		{"git@github.com:foo/bar", "bar"},
		{"ssh://git@github.com/foo/bar.git", "bar"},
		{"https://gitlab.com/group/sub/proj.git", "proj"},
	}
	for _, tt := range tests {
		got, err := deriveCloneName(tt.in)
		if err != nil {
			t.Errorf("deriveCloneName(%q): unexpected error %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("deriveCloneName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCloneCmd_Registered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "clone" {
			return
		}
	}
	t.Fatal("clone command is not registered on rootCmd")
}

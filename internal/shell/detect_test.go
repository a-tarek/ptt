package shell

import (
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name        string
		shellEnv    string
		wantShell   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "bash from /bin/bash",
			shellEnv:  "/bin/bash",
			wantShell: "bash",
			wantErr:   false,
		},
		{
			name:      "zsh from /usr/local/bin/zsh",
			shellEnv:  "/usr/local/bin/zsh",
			wantShell: "zsh",
			wantErr:   false,
		},
		{
			name:      "fish from /usr/bin/fish",
			shellEnv:  "/usr/bin/fish",
			wantShell: "fish",
			wantErr:   false,
		},
		{
			name:        "unsupported csh",
			shellEnv:    "/bin/csh",
			wantShell:   "",
			wantErr:     true,
			errContains: "unsupported shell: csh",
		},
		{
			name:        "unset SHELL",
			shellEnv:    "",
			wantShell:   "",
			wantErr:     true,
			errContains: "SHELL environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shellEnv)

			got, err := DetectShell()

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectShell() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("DetectShell() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DetectShell() unexpected error: %v", err)
				return
			}

			if got != tt.wantShell {
				t.Errorf("DetectShell() = %q, want %q", got, tt.wantShell)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}

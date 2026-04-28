package cmd

import (
	"strings"
	"testing"
)

func TestValidateWorktreeName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string // substring expected in error message; "" means no error
	}{
		{"plain hyphenated name", "feat-x", ""},
		{"alphanumeric", "abc123", ""},
		{"reject empty", "", "empty"},
		{"reject slash", "feat/x", "/"},
		{"reject leading hyphen", "-feat", "'-'"},
		{"reject nested slash", "team/sub/feat", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorktreeName(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validateWorktreeName(%q): unexpected error %v", tt.input, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateWorktreeName(%q): expected error containing %q, got nil", tt.input, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("validateWorktreeName(%q): error %q does not contain %q", tt.input, err.Error(), tt.wantErr)
			}
		})
	}
}

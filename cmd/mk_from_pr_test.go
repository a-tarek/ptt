package cmd

import "testing"

func TestParsePRNumber(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"2605", 2605, false},
		{"#2605", 2605, false},
		{"1", 1, false},
		{"0", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		got, err := parsePRNumber(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parsePRNumber(%q): expected error, got %d", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parsePRNumber(%q): unexpected error %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parsePRNumber(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestMkFromPRCmd_Registered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "mk-from-pr" {
			return
		}
	}
	t.Fatal("mk-from-pr command is not registered on rootCmd")
}

package git

import (
	"testing"
)

func TestParseBranches(t *testing.T) {
	tests := []struct {
		name         string
		localOutput  string
		remoteOutput string
		want         []string
	}{
		{
			name:         "local only",
			localOutput:  "main\ndev\nfeature/foo\n",
			remoteOutput: "",
			want:         []string{"main", "dev", "feature/foo"},
		},
		{
			name:         "with remotes deduped",
			localOutput:  "main\ndev\n",
			remoteOutput: "origin/main\norigin/dev\norigin/feature/bar\n",
			want:         []string{"main", "dev", "feature/bar"},
		},
		{
			name:         "HEAD pointer filtered",
			localOutput:  "main\n",
			remoteOutput: "origin/HEAD\norigin/main\norigin/release\n",
			want:         []string{"main", "release"},
		},
		{
			name:         "empty output",
			localOutput:  "",
			remoteOutput: "",
			want:         nil,
		},
		{
			name:         "branches with slashes",
			localOutput:  "feature/auth/login\n",
			remoteOutput: "origin/feature/auth/signup\n",
			want:         []string{"feature/auth/login", "feature/auth/signup"},
		},
		{
			name:         "whitespace trimmed",
			localOutput:  "  main  \n  dev  \n",
			remoteOutput: "  origin/staging  \n",
			want:         []string{"main", "dev", "staging"},
		},
		{
			name:         "remote only no local",
			localOutput:  "",
			remoteOutput: "origin/main\norigin/dev\n",
			want:         []string{"main", "dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBranches(tt.localOutput, tt.remoteOutput)

			if len(got) != len(tt.want) {
				t.Fatalf("parseBranches() returned %d items, want %d\ngot:  %v\nwant: %v",
					len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseBranches()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

package installer

import (
	"strings"
	"testing"
)

func TestHasMarkerBlock(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "empty content",
			content: "",
			want:    false,
		},
		{
			name: "has marker block",
			content: `some config
# >>> wt >>>
eval "$(wt shell-init)"
# <<< wt <<<
more config`,
			want: true,
		},
		{
			name: "no marker block",
			content: `some config
export PATH=$PATH:/usr/local/bin
alias ll='ls -la'`,
			want: false,
		},
		{
			name: "has begin marker only",
			content: `some config
# >>> wt >>>
eval "$(wt shell-init)"`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasMarkerBlock(tt.content); got != tt.want {
				t.Errorf("HasMarkerBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindV1Lines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []int
	}{
		{
			name:    "empty content",
			content: "",
			want:    nil,
		},
		{
			name: "source with full path",
			content: `export PATH=$PATH:/usr/local/bin
source ~/.oh-my-zsh/custom/plugins/wt/wt.zsh
alias ll='ls -la'`,
			want: []int{1},
		},
		{
			name: "dot-source",
			content: `export PATH=$PATH:/usr/local/bin
. ~/wt.zsh
alias ll='ls -la'`,
			want: []int{1},
		},
		{
			name: "commented v1 line",
			content: `export PATH=$PATH:/usr/local/bin
# source ~/.oh-my-zsh/custom/plugins/wt/wt.zsh
alias ll='ls -la'`,
			want: nil,
		},
		{
			name: "multiple v1 entries",
			content: `source /path/to/wt.zsh
export PATH=$PATH:/usr/local/bin
. ~/other/wt.zsh`,
			want: []int{0, 2},
		},
		{
			name: "no v1 entries",
			content: `export PATH=$PATH:/usr/local/bin
alias wt='echo hello'
source ~/.bashrc`,
			want: nil,
		},
		{
			name: "whitespace before source",
			content: `export PATH=$PATH:/usr/local/bin
  source ~/wt.zsh
    . /path/wt.zsh`,
			want: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindV1Lines(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("FindV1Lines() returned %d indices, want %d: got=%v, want=%v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FindV1Lines()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCommentOutLines(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		indices []int
		want    []string
	}{
		{
			name:    "empty",
			lines:   []string{},
			indices: []int{},
			want:    []string{},
		},
		{
			name:    "comment single line",
			lines:   []string{"line1", "line2", "line3"},
			indices: []int{1},
			want:    []string{"line1", "# [wt v2 migration] line2", "line3"},
		},
		{
			name:    "comment multiple lines",
			lines:   []string{"line1", "line2", "line3", "line4"},
			indices: []int{0, 2},
			want:    []string{"# [wt v2 migration] line1", "line2", "# [wt v2 migration] line3", "line4"},
		},
		{
			name:    "no indices",
			lines:   []string{"line1", "line2"},
			indices: []int{},
			want:    []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify original is not mutated
			original := make([]string, len(tt.lines))
			copy(original, tt.lines)

			got := CommentOutLines(tt.lines, tt.indices)

			// Check original unchanged
			for i := range tt.lines {
				if tt.lines[i] != original[i] {
					t.Errorf("CommentOutLines() mutated input at index %d", i)
				}
			}

			// Check result
			if len(got) != len(tt.want) {
				t.Errorf("CommentOutLines() returned %d lines, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CommentOutLines()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestInsertMarkerBlock(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shellType string
		wantHas   bool
	}{
		{
			name:      "empty content bash",
			content:   "",
			shellType: "bash",
			wantHas:   true,
		},
		{
			name:      "content with newline zsh",
			content:   "export PATH=$PATH:/usr/local/bin\n",
			shellType: "zsh",
			wantHas:   true,
		},
		{
			name:      "content without newline fish",
			content:   "export PATH=$PATH:/usr/local/bin",
			shellType: "fish",
			wantHas:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InsertMarkerBlock(tt.content, tt.shellType)

			// Check marker exists
			if HasMarkerBlock(got) != tt.wantHas {
				t.Errorf("InsertMarkerBlock() marker presence = %v, want %v", HasMarkerBlock(got), tt.wantHas)
			}

			// Check original content is preserved
			if !strings.Contains(got, strings.TrimSuffix(tt.content, "\n")) {
				t.Errorf("InsertMarkerBlock() lost original content")
			}

			// Check ends with newline
			if !strings.HasSuffix(got, "\n") {
				t.Errorf("InsertMarkerBlock() result doesn't end with newline")
			}

			// Check has correct eval line for shell
			expectedEval := EvalLine(tt.shellType)
			if !strings.Contains(got, expectedEval) {
				t.Errorf("InsertMarkerBlock() missing expected eval line for %s: %q", tt.shellType, expectedEval)
			}
		})
	}
}

func TestRemoveMarkerBlock(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "no marker block",
			content: "export PATH=$PATH:/usr/local/bin\nalias ll='ls -la'",
			want:    "export PATH=$PATH:/usr/local/bin\nalias ll='ls -la'",
		},
		{
			name: "has marker block in middle",
			content: `export PATH=$PATH:/usr/local/bin
# >>> wt >>>
eval "$(wt shell-init)"
# <<< wt <<<
alias ll='ls -la'`,
			want: `export PATH=$PATH:/usr/local/bin

alias ll='ls -la'`,
		},
		{
			name: "has marker block at end",
			content: `export PATH=$PATH:/usr/local/bin
# >>> wt >>>
eval "$(wt shell-init)"
# <<< wt <<<`,
			want: `export PATH=$PATH:/usr/local/bin
`,
		},
		{
			name: "has marker block at start",
			content: `# >>> wt >>>
eval "$(wt shell-init)"
# <<< wt <<<
export PATH=$PATH:/usr/local/bin`,
			want: `
export PATH=$PATH:/usr/local/bin`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveMarkerBlock(tt.content)
			if got != tt.want {
				t.Errorf("RemoveMarkerBlock() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMarkerBlock(t *testing.T) {
	tests := []struct {
		shellType string
		wantEval  string
	}{
		{"bash", `eval "$(wt shell-init)"`},
		{"zsh", `eval "$(wt shell-init)"`},
		{"fish", `wt shell-init | source`},
	}

	for _, tt := range tests {
		t.Run(tt.shellType, func(t *testing.T) {
			got := MarkerBlock(tt.shellType)

			// Check has markers
			if !strings.Contains(got, MarkerBegin) {
				t.Errorf("MarkerBlock(%s) missing begin marker", tt.shellType)
			}
			if !strings.Contains(got, MarkerEnd) {
				t.Errorf("MarkerBlock(%s) missing end marker", tt.shellType)
			}

			// Check has correct eval line
			if !strings.Contains(got, tt.wantEval) {
				t.Errorf("MarkerBlock(%s) = %q, want to contain %q", tt.shellType, got, tt.wantEval)
			}

			// Check has managed comment
			if !strings.Contains(got, "Managed by wt installer") {
				t.Errorf("MarkerBlock(%s) missing managed comment", tt.shellType)
			}
		})
	}
}

func TestEvalLine(t *testing.T) {
	tests := []struct {
		shellType string
		want      string
	}{
		{"bash", `eval "$(wt shell-init)"`},
		{"zsh", `eval "$(wt shell-init)"`},
		{"fish", `wt shell-init | source`},
	}

	for _, tt := range tests {
		t.Run(tt.shellType, func(t *testing.T) {
			if got := EvalLine(tt.shellType); got != tt.want {
				t.Errorf("EvalLine(%s) = %q, want %q", tt.shellType, got, tt.want)
			}
		})
	}
}

package config

// Action types
const (
	ActionCopy    = "copy"
	ActionSymlink = "symlink"
	ActionRun     = "run"
)

// Action represents a single config action
type Action struct {
	Type string // "copy", "symlink", or "run"
	Path string // file/directory path or command string for "run"
	Line int    // source line number (for error reporting)
}

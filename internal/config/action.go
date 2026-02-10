package config

// Action types
const (
	ActionCopy    = "copy"
	ActionSymlink = "symlink"
	ActionRun     = "run"
	ActionCopyEnv = "copyEnv"
)

// EnvVarRule defines a strategy for resolving an env variable value.
type EnvVarRule struct {
	Strategy string // "ptt_increment" or a shell command
}

// CopyEnvConfig holds configuration for a copyEnv action.
type CopyEnvConfig struct {
	File      string                // source env file path (relative to config root)
	Vars      map[string]EnvVarRule // from YAML config
	Overrides map[string]string     // from --set flags (KEY=VALUE or KEY=ptt_increment)
}

// Action represents a single config action
type Action struct {
	Type    string         // "copy", "symlink", "run", or "copyEnv"
	Path    string         // file/directory path or command string for "run"
	Line    int            // source line number (for error reporting)
	CopyEnv *CopyEnvConfig // non-nil only for ActionCopyEnv
}

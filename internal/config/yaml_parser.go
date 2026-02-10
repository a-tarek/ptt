package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LifecycleConfig holds actions for create and remove lifecycle phases.
type LifecycleConfig struct {
	Create []Action
	Remove []Action
}

type yamlConfig struct {
	Create []yamlAction `yaml:"create"`
	Remove []yamlAction `yaml:"remove"`
}

type yamlAction struct {
	Copy    string             `yaml:"copy"`
	Symlink string             `yaml:"symlink"`
	Run     string             `yaml:"run"`
	CopyEnv *yamlCopyEnvConfig `yaml:"copyEnv"`
}

type yamlCopyEnvConfig struct {
	File string                    `yaml:"file"`
	Vars map[string]yamlEnvVarRule `yaml:"vars"`
}

type yamlEnvVarRule struct {
	Strategy string `yaml:"strategy"`
}

// ParseYAMLFile reads a YAML config file and returns a LifecycleConfig.
func ParseYAMLFile(path string) (*LifecycleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	lc := &LifecycleConfig{}

	for i, a := range yc.Create {
		action, err := convertYAMLAction(a, i+1)
		if err != nil {
			return nil, fmt.Errorf("create[%d]: %w", i, err)
		}
		lc.Create = append(lc.Create, action)
	}

	for i, a := range yc.Remove {
		action, err := convertYAMLAction(a, i+1)
		if err != nil {
			return nil, fmt.Errorf("remove[%d]: %w", i, err)
		}
		lc.Remove = append(lc.Remove, action)
	}

	return lc, nil
}

// IsYAMLConfig returns true if the path has a .yml or .yaml extension.
func IsYAMLConfig(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yml" || ext == ".yaml"
}

func convertYAMLAction(a yamlAction, line int) (Action, error) {
	count := 0
	var action Action
	action.Line = line

	if a.Copy != "" {
		count++
		action.Type = ActionCopy
		action.Path = a.Copy
	}
	if a.Symlink != "" {
		count++
		action.Type = ActionSymlink
		action.Path = a.Symlink
	}
	if a.Run != "" {
		count++
		action.Type = ActionRun
		action.Path = a.Run
	}
	if a.CopyEnv != nil {
		count++
		action.Type = ActionCopyEnv
		action.Path = a.CopyEnv.File
		vars := make(map[string]EnvVarRule, len(a.CopyEnv.Vars))
		for k, v := range a.CopyEnv.Vars {
			vars[k] = EnvVarRule{Strategy: v.Strategy}
		}
		action.CopyEnv = &CopyEnvConfig{
			File: a.CopyEnv.File,
			Vars: vars,
		}
	}

	if count == 0 {
		return Action{}, fmt.Errorf("action has no fields set")
	}
	if count > 1 {
		return Action{}, fmt.Errorf("action has multiple fields set")
	}

	return action, nil
}

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// MasterConfig represents the MasterCortex.yml configuration.
// It allows orchestrating multiple Cortexfiles together.
type MasterConfig struct {
	// Name is an optional name for this master workflow
	Name string `yaml:"name"`

	// Description describes what this master workflow does
	Description string `yaml:"description"`

	// Mode defines execution mode: "parallel" or "sequential" (default: sequential)
	Mode string `yaml:"mode"`

	// MaxParallel limits concurrent Cortexfile executions (0 = unlimited)
	MaxParallel int `yaml:"max_parallel"`

	// StopOnError stops execution on first error (default: true for sequential, false for parallel)
	StopOnError *bool `yaml:"stop_on_error"`

	// Workflows defines the Cortexfiles to run
	Workflows []WorkflowEntry `yaml:"workflows"`

	// Variables defines global variables available to all workflows
	Variables map[string]string `yaml:"variables"`
}

// WorkflowEntry represents a single Cortexfile entry in the master config.
type WorkflowEntry struct {
	// Name is an optional friendly name for this workflow
	Name string `yaml:"name"`

	// Path is the path to the Cortexfile (supports glob patterns)
	Path string `yaml:"path"`

	// Workdir overrides the working directory for this workflow
	Workdir string `yaml:"workdir"`

	// Enabled allows disabling a workflow without removing it (default: true)
	Enabled *bool `yaml:"enabled"`

	// Needs specifies dependencies on other workflows (by name)
	Needs StringList `yaml:"needs"`

	// Variables for this specific workflow (merged with global)
	Variables map[string]string `yaml:"variables"`
}

// MasterCortexFiles are the filenames to search for
var MasterCortexFiles = []string{
	"MasterCortex.yml",
	"MasterCortex.yaml",
	"master-cortex.yml",
	"master-cortex.yaml",
}

// FindMasterCortex searches for a MasterCortex file in the given directory.
func FindMasterCortex(dir string) (string, error) {
	for _, name := range MasterCortexFiles {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no MasterCortex file found in %s (tried: %v)", dir, MasterCortexFiles)
}

// LoadMasterConfig loads a MasterCortex configuration from the given path.
func LoadMasterConfig(path string) (*MasterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read master config: %w", err)
	}

	var config MasterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse master config: %w", err)
	}

	// Apply defaults
	if config.Mode == "" {
		config.Mode = "sequential"
	}

	// Validate mode
	if config.Mode != "sequential" && config.Mode != "parallel" {
		return nil, fmt.Errorf("invalid mode %q: must be 'sequential' or 'parallel'", config.Mode)
	}

	// Set default for StopOnError based on mode
	if config.StopOnError == nil {
		stopOnError := config.Mode == "sequential"
		config.StopOnError = &stopOnError
	}

	// Set defaults for workflow entries
	for i := range config.Workflows {
		if config.Workflows[i].Enabled == nil {
			enabled := true
			config.Workflows[i].Enabled = &enabled
		}
		// Generate name if not provided
		if config.Workflows[i].Name == "" {
			config.Workflows[i].Name = fmt.Sprintf("workflow-%d", i+1)
		}
	}

	return &config, nil
}

// ValidateMasterConfig validates the master configuration.
func ValidateMasterConfig(cfg *MasterConfig) error {
	if len(cfg.Workflows) == 0 {
		return fmt.Errorf("no workflows defined")
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for _, w := range cfg.Workflows {
		if names[w.Name] {
			return fmt.Errorf("duplicate workflow name: %s", w.Name)
		}
		names[w.Name] = true
	}

	// Validate dependencies exist
	for _, w := range cfg.Workflows {
		for _, dep := range w.Needs {
			if !names[dep] {
				return fmt.Errorf("workflow %q depends on unknown workflow %q", w.Name, dep)
			}
		}
	}

	// Check for path
	for _, w := range cfg.Workflows {
		if w.Path == "" {
			return fmt.Errorf("workflow %q has no path specified", w.Name)
		}
	}

	return nil
}

// ResolveWorkflowPaths expands glob patterns in workflow paths and returns resolved entries.
func ResolveWorkflowPaths(cfg *MasterConfig, baseDir string) ([]WorkflowEntry, error) {
	var resolved []WorkflowEntry

	for _, w := range cfg.Workflows {
		if w.Enabled != nil && !*w.Enabled {
			continue
		}

		// Make path absolute relative to baseDir
		path := w.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}

		// Check if it's a glob pattern
		if containsGlob(path) {
			matches, err := filepath.Glob(path)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern %q: %w", path, err)
			}
			for i, match := range matches {
				entry := w
				entry.Path = match
				if len(matches) > 1 {
					entry.Name = fmt.Sprintf("%s-%d", w.Name, i+1)
				}
				resolved = append(resolved, entry)
			}
		} else {
			entry := w
			entry.Path = path
			resolved = append(resolved, entry)
		}
	}

	return resolved, nil
}

// containsGlob checks if a path contains glob characters
func containsGlob(s string) bool {
	for _, c := range s {
		if c == '*' || c == '?' || c == '[' {
			return true
		}
	}
	return false
}

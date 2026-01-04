package config

import (
	"strings"
	"testing"
)

// Helper to check if any error contains a substring
func errorsContain(errs *ConfigErrors, substring string) bool {
	if errs == nil {
		return false
	}
	for _, err := range errs.Errors {
		if strings.Contains(err.Error(), substring) {
			return true
		}
	}
	return false
}

// TestValidate_EmptyConfig tests validation of empty configurations.
func TestValidate_EmptyConfig(t *testing.T) {
	tests := []struct {
		name            string
		config          *AgentflowConfig
		wantErrCount    int
		wantErrContains []string
	}{
		{
			name:            "completely empty config",
			config:          &AgentflowConfig{},
			wantErrCount:    2,
			wantErrContains: []string{"no agents defined", "no tasks defined"},
		},
		{
			name: "no agents",
			config: &AgentflowConfig{
				Tasks: map[string]TaskConfig{
					"task1": {Agent: "agent1", Prompt: "test"},
				},
			},
			wantErrCount:    2, // "no agents defined" + "references undefined agent"
			wantErrContains: []string{"no agents defined"},
		},
		{
			name: "no tasks",
			config: &AgentflowConfig{
				Agents: map[string]AgentConfig{
					"agent1": {Tool: "claude-code"},
				},
			},
			wantErrCount:    1,
			wantErrContains: []string{"no tasks defined"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			valErr, ok := err.(*ConfigErrors)
			if !ok {
				t.Fatalf("expected *ConfigErrors, got %T", err)
			}

			if len(valErr.Errors) != tt.wantErrCount {
				t.Errorf("expected %d errors, got %d: %v", tt.wantErrCount, len(valErr.Errors), valErr.Errors)
			}

			for _, expectedMsg := range tt.wantErrContains {
				if !errorsContain(valErr, expectedMsg) {
					t.Errorf("expected error message containing %q, got errors: %v", expectedMsg, valErr.Error())
				}
			}
		})
	}
}

// TestValidate_AgentValidation tests agent-level validation.
func TestValidate_AgentValidation(t *testing.T) {
	tests := []struct {
		name            string
		agents          map[string]AgentConfig
		tasks           map[string]TaskConfig
		wantErrContains []string
	}{
		{
			name: "missing tool",
			agents: map[string]AgentConfig{
				"agent1": {Tool: ""},
			},
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test"},
			},
			wantErrContains: []string{`agent "agent1": tool is required`},
		},
		{
			name: "unsupported tool",
			agents: map[string]AgentConfig{
				"agent1": {Tool: "invalid-tool"},
			},
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test"},
			},
			wantErrContains: []string{`unsupported tool "invalid-tool"`},
		},
		{
			name: "valid supported tools",
			agents: map[string]AgentConfig{
				"agent1": {Tool: "claude-code"},
				"agent2": {Tool: "opencode"},
			},
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test1"},
				"task2": {Agent: "agent2", Prompt: "test2"},
			},
			wantErrContains: nil, // No errors expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AgentflowConfig{
				Agents: tt.agents,
				Tasks:  tt.tasks,
			}

			err := Validate(config)
			if tt.wantErrContains == nil {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			valErr, ok := err.(*ConfigErrors)
			if !ok {
				t.Fatalf("expected *ConfigErrors, got %T", err)
			}

			for _, expectedMsg := range tt.wantErrContains {
				if !errorsContain(valErr, expectedMsg) {
					t.Errorf("expected error containing %q, got errors: %v", expectedMsg, valErr.Error())
				}
			}
		})
	}
}

// TestValidate_TaskValidation tests task-level validation.
func TestValidate_TaskValidation(t *testing.T) {
	validAgent := map[string]AgentConfig{
		"agent1": {Tool: "claude-code"},
	}

	tests := []struct {
		name            string
		tasks           map[string]TaskConfig
		wantErrContains []string
	}{
		{
			name: "missing agent reference",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "", Prompt: "test"},
			},
			wantErrContains: []string{`task "task1": agent is required`},
		},
		{
			name: "undefined agent reference",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "nonexistent", Prompt: "test"},
			},
			wantErrContains: []string{`references undefined agent "nonexistent"`},
		},
		{
			name: "missing both prompt and prompt_file",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1"},
			},
			wantErrContains: []string{`task "task1" has no prompt defined`},
		},
		{
			name: "both prompt and prompt_file specified",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test", PromptFile: "test.txt"},
			},
			wantErrContains: []string{`cannot have both 'prompt' and 'prompt_file'`},
		},
		{
			name: "undefined dependency",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test", Needs: []string{"nonexistent"}},
			},
			wantErrContains: []string{`depends on undefined task "nonexistent"`},
		},
		{
			name: "self-dependency",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test", Needs: []string{"task1"}},
			},
			wantErrContains: []string{`cannot depend on itself`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AgentflowConfig{
				Agents: validAgent,
				Tasks:  tt.tasks,
			}

			err := Validate(config)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			valErr, ok := err.(*ConfigErrors)
			if !ok {
				t.Fatalf("expected *ConfigErrors, got %T", err)
			}

			for _, expectedMsg := range tt.wantErrContains {
				if !errorsContain(valErr, expectedMsg) {
					t.Errorf("expected error containing %q, got errors: %v", expectedMsg, valErr.Error())
				}
			}
		})
	}
}

// TestValidate_CircularDependencies tests cycle detection.
func TestValidate_CircularDependencies(t *testing.T) {
	validAgent := map[string]AgentConfig{
		"agent1": {Tool: "claude-code"},
	}

	tests := []struct {
		name            string
		tasks           map[string]TaskConfig
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "simple two-task cycle",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test1", Needs: []string{"task2"}},
				"task2": {Agent: "agent1", Prompt: "test2", Needs: []string{"task1"}},
			},
			wantErr:         true,
			wantErrContains: "circular dependency detected",
		},
		{
			name: "three-task cycle",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test1", Needs: []string{"task2"}},
				"task2": {Agent: "agent1", Prompt: "test2", Needs: []string{"task3"}},
				"task3": {Agent: "agent1", Prompt: "test3", Needs: []string{"task1"}},
			},
			wantErr:         true,
			wantErrContains: "circular dependency detected",
		},
		{
			name: "self-cycle already caught by task validation",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test1", Needs: []string{"task1"}},
			},
			wantErr:         true,
			wantErrContains: "cannot depend on itself",
		},
		{
			name: "valid DAG - no cycles",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test1"},
				"task2": {Agent: "agent1", Prompt: "test2", Needs: []string{"task1"}},
				"task3": {Agent: "agent1", Prompt: "test3", Needs: []string{"task1", "task2"}},
			},
			wantErr: false,
		},
		{
			name: "diamond dependency - valid",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "test1"},
				"task2": {Agent: "agent1", Prompt: "test2", Needs: []string{"task1"}},
				"task3": {Agent: "agent1", Prompt: "test3", Needs: []string{"task1"}},
				"task4": {Agent: "agent1", Prompt: "test4", Needs: []string{"task2", "task3"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AgentflowConfig{
				Agents: validAgent,
				Tasks:  tt.tasks,
			}

			err := Validate(config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidate_TemplateVariables tests template variable validation.
func TestValidate_TemplateVariables(t *testing.T) {
	validAgent := map[string]AgentConfig{
		"agent1": {Tool: "claude-code"},
	}

	tests := []struct {
		name            string
		tasks           map[string]TaskConfig
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "valid template reference",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "first task"},
				"task2": {Agent: "agent1", Prompt: "Use output: {{outputs.task1}}", Needs: []string{"task1"}},
			},
			wantErr: false,
		},
		{
			name: "template references undefined task",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "Use output: {{outputs.nonexistent}}"},
			},
			wantErr:         true,
			wantErrContains: `template references undefined task "nonexistent"`,
		},
		{
			name: "template references task not in needs",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "first"},
				"task2": {Agent: "agent1", Prompt: "Use output: {{outputs.task1}}"},
			},
			wantErr:         true,
			wantErrContains: `template references "task1" which is not in 'needs'`,
		},
		{
			name: "multiple valid template references",
			tasks: map[string]TaskConfig{
				"task1": {Agent: "agent1", Prompt: "first"},
				"task2": {Agent: "agent1", Prompt: "second"},
				"task3": {
					Agent:  "agent1",
					Prompt: "Combine: {{outputs.task1}} and {{outputs.task2}}",
					Needs:  []string{"task1", "task2"},
				},
			},
			wantErr: false,
		},
		{
			name: "template with hyphens and underscores",
			tasks: map[string]TaskConfig{
				"task-1_test": {Agent: "agent1", Prompt: "first"},
				"task2": {
					Agent:  "agent1",
					Prompt: "Use: {{outputs.task-1_test}}",
					Needs:  []string{"task-1_test"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AgentflowConfig{
				Agents: validAgent,
				Tasks:  tt.tasks,
			}

			err := Validate(config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidateTemplateVarsStructured tests the template variable validation function directly.
func TestValidateTemplateVarsStructured(t *testing.T) {
	tests := []struct {
		name            string
		taskName        string
		prompt          string
		needs           []string
		tasks           map[string]TaskConfig
		wantErrCount    int
		wantErrContains string
	}{
		{
			name:         "no template vars",
			taskName:     "task1",
			prompt:       "plain prompt",
			needs:        []string{},
			tasks:        map[string]TaskConfig{},
			wantErrCount: 0,
		},
		{
			name:     "valid single template var",
			taskName: "task2",
			prompt:   "Use {{outputs.task1}}",
			needs:    []string{"task1"},
			tasks: map[string]TaskConfig{
				"task1": {Prompt: "test"},
			},
			wantErrCount: 0,
		},
		{
			name:            "template references undefined task",
			taskName:        "task2",
			prompt:          "Use {{outputs.undefined}}",
			needs:           []string{},
			tasks:           map[string]TaskConfig{},
			wantErrCount:    1,
			wantErrContains: "template references undefined task",
		},
		{
			name:     "template references task not in needs",
			taskName: "task2",
			prompt:   "Use {{outputs.task1}}",
			needs:    []string{},
			tasks: map[string]TaskConfig{
				"task1": {Prompt: "test"},
			},
			wantErrCount:    1,
			wantErrContains: "which is not in 'needs'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateTemplateVarsStructured("test.yml", tt.taskName, tt.prompt, tt.needs, tt.tasks)
			if len(errors) != tt.wantErrCount {
				t.Errorf("expected %d errors, got %d: %v", tt.wantErrCount, len(errors), errors)
			}
			if tt.wantErrContains != "" {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Error(), tt.wantErrContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got errors: %v", tt.wantErrContains, errors)
				}
			}
		})
	}
}

// TestDetectCycles tests the cycle detection function directly.
func TestDetectCycles(t *testing.T) {
	tests := []struct {
		name    string
		tasks   map[string]TaskConfig
		wantErr bool
	}{
		{
			name:    "no tasks",
			tasks:   map[string]TaskConfig{},
			wantErr: false,
		},
		{
			name: "single task no dependencies",
			tasks: map[string]TaskConfig{
				"task1": {Prompt: "test"},
			},
			wantErr: false,
		},
		{
			name: "linear dependency chain",
			tasks: map[string]TaskConfig{
				"task1": {Prompt: "test1"},
				"task2": {Prompt: "test2", Needs: []string{"task1"}},
				"task3": {Prompt: "test3", Needs: []string{"task2"}},
			},
			wantErr: false,
		},
		{
			name: "simple cycle",
			tasks: map[string]TaskConfig{
				"task1": {Prompt: "test1", Needs: []string{"task2"}},
				"task2": {Prompt: "test2", Needs: []string{"task1"}},
			},
			wantErr: true,
		},
		{
			name: "long cycle",
			tasks: map[string]TaskConfig{
				"task1": {Prompt: "test1", Needs: []string{"task2"}},
				"task2": {Prompt: "test2", Needs: []string{"task3"}},
				"task3": {Prompt: "test3", Needs: []string{"task4"}},
				"task4": {Prompt: "test4", Needs: []string{"task1"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := detectCycles(tt.tasks)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectCycles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigErrors_Error tests the ConfigErrors error message formatting.
func TestConfigErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   []*ConfigError
		contains string
	}{
		{
			name: "single error",
			errors: []*ConfigError{
				{Message: "error 1"},
			},
			contains: "error 1",
		},
		{
			name: "multiple errors",
			errors: []*ConfigError{
				{Message: "error 1"},
				{Message: "error 2"},
				{Message: "error 3"},
			},
			contains: "3 configuration errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := &ConfigErrors{Errors: tt.errors}
			got := ce.Error()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("ConfigErrors.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

// TestIsSupportedTool tests the tool support check function.
func TestIsSupportedTool(t *testing.T) {
	tests := []struct {
		name string
		tool string
		want bool
	}{
		{"claude-code supported", "claude-code", true},
		{"opencode supported", "opencode", true},
		{"unsupported tool", "invalid-tool", false},
		{"empty string", "", false},
		{"case sensitive", "Claude-Code", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedTool(tt.tool); got != tt.want {
				t.Errorf("IsSupportedTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

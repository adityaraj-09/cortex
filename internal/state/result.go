// Package state handles run results and persistence.
package state

import (
	"time"
)

// TokenUsage represents token usage for a task.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
	CacheRead    int `json:"cache_read_tokens,omitempty"`
	CacheWrite   int `json:"cache_write_tokens,omitempty"`
}

// TaskResult represents the result of executing a single task.
type TaskResult struct {
	TaskName   string     `json:"task_name"`
	Agent      string     `json:"agent"`
	Tool       string     `json:"tool"`
	Model      string     `json:"model,omitempty"`
	Prompt     string     `json:"prompt"`
	Stdout     string     `json:"stdout"`
	Stderr     string     `json:"stderr,omitempty"`
	Success    bool       `json:"success"`
	ExitCode   int        `json:"exit_code"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    time.Time  `json:"end_time"`
	Duration   string     `json:"duration"` // Human-readable duration
	TokenUsage TokenUsage `json:"token_usage,omitempty"`
}

// RunResult represents the complete result of an agentflow run.
type RunResult struct {
	RunID      string       `json:"run_id"`
	StartTime  time.Time    `json:"start_time"`
	EndTime    time.Time    `json:"end_time"`
	Success    bool         `json:"success"`
	Tasks      []TaskResult `json:"tasks"`
	TokenUsage TokenUsage   `json:"token_usage,omitempty"` // Aggregate token usage
}

// CalculateTotalTokens calculates aggregate token usage from all tasks.
func (r *RunResult) CalculateTotalTokens() {
	r.TokenUsage = TokenUsage{}
	for _, task := range r.Tasks {
		r.TokenUsage.InputTokens += task.TokenUsage.InputTokens
		r.TokenUsage.OutputTokens += task.TokenUsage.OutputTokens
		r.TokenUsage.TotalTokens += task.TokenUsage.TotalTokens
		r.TokenUsage.CacheRead += task.TokenUsage.CacheRead
		r.TokenUsage.CacheWrite += task.TokenUsage.CacheWrite
	}
}

// NewTaskResult creates a new TaskResult with timing started.
func NewTaskResult(taskName, agent, tool, model, prompt string) *TaskResult {
	return &TaskResult{
		TaskName:  taskName,
		Agent:     agent,
		Tool:      tool,
		Model:     model,
		Prompt:    prompt,
		StartTime: time.Now(),
	}
}

// Complete marks the task as completed with the given result.
func (r *TaskResult) Complete(stdout, stderr string, exitCode int, success bool) {
	r.Stdout = stdout
	r.Stderr = stderr
	r.ExitCode = exitCode
	r.Success = success
	r.EndTime = time.Now()
	r.Duration = r.EndTime.Sub(r.StartTime).Round(time.Millisecond * 100).String()
}

// SetTokenUsage sets the token usage for the task.
func (r *TaskResult) SetTokenUsage(input, output, cacheRead, cacheWrite int) {
	r.TokenUsage = TokenUsage{
		InputTokens:  input,
		OutputTokens: output,
		TotalTokens:  input + output,
		CacheRead:    cacheRead,
		CacheWrite:   cacheWrite,
	}
}

// Package observability provides logging and monitoring capabilities for Cortex.
package observability

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// LogFormat specifies the output format for logs
type LogFormat string

const (
	FormatText LogFormat = "text"
	FormatJSON LogFormat = "json"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Task    string    `json:"task,omitempty"`
	Event   string    `json:"event,omitempty"`
	Data    any       `json:"data,omitempty"`
}

// Logger provides structured logging capabilities
type Logger struct {
	level   LogLevel
	format  LogFormat
	output  io.Writer
	mu      sync.Mutex
	enabled bool
}

// LoggerConfig holds configuration for creating a Logger
type LoggerConfig struct {
	Level   LogLevel
	Format  LogFormat
	Output  io.Writer
	Enabled bool
}

// NewLogger creates a new Logger with the specified configuration
func NewLogger(cfg LoggerConfig) *Logger {
	output := cfg.Output
	if output == nil {
		output = os.Stderr
	}

	return &Logger{
		level:   cfg.Level,
		format:  cfg.Format,
		output:  output,
		enabled: cfg.Enabled,
	}
}

// DefaultLogger returns a logger with default settings
func DefaultLogger() *Logger {
	return NewLogger(LoggerConfig{
		Level:   LevelInfo,
		Format:  FormatText,
		Output:  os.Stderr,
		Enabled: false, // Disabled by default
	})
}

// SetEnabled enables or disables logging
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	l.enabled = enabled
	l.mu.Unlock()
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// SetFormat sets the output format
func (l *Logger) SetFormat(format LogFormat) {
	l.mu.Lock()
	l.format = format
	l.mu.Unlock()
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

// log writes a log entry at the specified level
func (l *Logger) log(level LogLevel, msg string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled || level < l.level {
		return
	}

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level.String(),
		Message: msg,
	}

	// Apply fields
	for _, f := range fields {
		f.Apply(&entry)
	}

	// Format and write
	var output string
	if l.format == FormatJSON {
		data, err := json.Marshal(entry)
		if err != nil {
			output = fmt.Sprintf(`{"error":"failed to marshal log entry: %s"}`, err)
		} else {
			output = string(data)
		}
	} else {
		output = l.formatText(entry)
	}

	fmt.Fprintln(l.output, output)
}

// formatText formats a log entry as human-readable text
func (l *Logger) formatText(entry LogEntry) string {
	var sb strings.Builder

	// Timestamp
	sb.WriteString(entry.Time.Format("15:04:05"))
	sb.WriteString(" ")

	// Level with color-like prefix
	switch entry.Level {
	case "debug":
		sb.WriteString("[DBG]")
	case "info":
		sb.WriteString("[INF]")
	case "warn":
		sb.WriteString("[WRN]")
	case "error":
		sb.WriteString("[ERR]")
	}
	sb.WriteString(" ")

	// Message
	sb.WriteString(entry.Message)

	// Task
	if entry.Task != "" {
		sb.WriteString(fmt.Sprintf(" task=%s", entry.Task))
	}

	// Event
	if entry.Event != "" {
		sb.WriteString(fmt.Sprintf(" event=%s", entry.Event))
	}

	// Data
	if entry.Data != nil {
		if data, err := json.Marshal(entry.Data); err == nil {
			sb.WriteString(fmt.Sprintf(" data=%s", string(data)))
		}
	}

	return sb.String()
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// Field represents a log field that can be added to an entry
type Field func(*LogEntry)

// Apply applies the field to a log entry
func (f Field) Apply(entry *LogEntry) {
	f(entry)
}

// WithTask adds a task name to the log entry
func WithTask(name string) Field {
	return func(entry *LogEntry) {
		entry.Task = name
	}
}

// WithEvent adds an event type to the log entry
func WithEvent(event string) Field {
	return func(entry *LogEntry) {
		entry.Event = event
	}
}

// WithData adds arbitrary data to the log entry
func WithData(data any) Field {
	return func(entry *LogEntry) {
		entry.Data = data
	}
}

// Event types for structured logging
const (
	EventRunStart     = "run_start"
	EventRunComplete  = "run_complete"
	EventTaskStart    = "task_start"
	EventTaskComplete = "task_complete"
	EventTaskFailed   = "task_failed"
	EventWebhookSent  = "webhook_sent"
)

// TaskData represents task-related data for logging
type TaskData struct {
	Duration     string `json:"duration,omitempty"`
	ExitCode     int    `json:"exit_code,omitempty"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
	TotalTokens  int    `json:"total_tokens,omitempty"`
	Tool         string `json:"tool,omitempty"`
	Model        string `json:"model,omitempty"`
}

// RunData represents run-related data for logging
type RunData struct {
	RunID      string `json:"run_id,omitempty"`
	Project    string `json:"project,omitempty"`
	TaskCount  int    `json:"task_count,omitempty"`
	Duration   string `json:"duration,omitempty"`
	Success    bool   `json:"success"`
	ConfigFile string `json:"config_file,omitempty"`
}

// Global logger instance (can be replaced)
var globalLogger = DefaultLogger()

// SetGlobalLogger replaces the global logger
func SetGlobalLogger(l *Logger) {
	globalLogger = l
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Log convenience functions using global logger

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

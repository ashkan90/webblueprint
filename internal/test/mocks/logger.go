package mocks

import (
	"fmt"
	"strings"
)

// LogEntry represents a log entry
type LogEntry struct {
	Level  string
	Msg    string
	Fields map[string]interface{}
}

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	entries []LogEntry
	options map[string]interface{}
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		entries: make([]LogEntry, 0),
		options: make(map[string]interface{}),
	}
}

// Opts sets options for the logger
func (m *MockLogger) Opts(options map[string]interface{}) {
	m.options = options
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.entries = append(m.entries, LogEntry{
		Level:  "DEBUG",
		Msg:    msg,
		Fields: fields,
	})
}

// Info logs an info message
func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.entries = append(m.entries, LogEntry{
		Level:  "INFO",
		Msg:    msg,
		Fields: fields,
	})
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.entries = append(m.entries, LogEntry{
		Level:  "WARN",
		Msg:    msg,
		Fields: fields,
	})
}

// Error logs an error message
func (m *MockLogger) Error(msg string, fields map[string]interface{}) {
	m.entries = append(m.entries, LogEntry{
		Level:  "ERROR",
		Msg:    msg,
		Fields: fields,
	})
}

// GetEntries returns all log entries
func (m *MockLogger) GetEntries() []LogEntry {
	return m.entries
}

// GetLastEntry returns the most recent log entry
func (m *MockLogger) GetLastEntry() (LogEntry, bool) {
	if len(m.entries) == 0 {
		return LogEntry{}, false
	}
	return m.entries[len(m.entries)-1], true
}

// GetEntriesWithLevel returns all log entries with the specified level
func (m *MockLogger) GetEntriesWithLevel(level string) []LogEntry {
	level = strings.ToUpper(level)
	var entries []LogEntry
	for _, entry := range m.entries {
		if entry.Level == level {
			entries = append(entries, entry)
		}
	}
	return entries
}

// ContainsMessage checks if any log entry contains the specified message
func (m *MockLogger) ContainsMessage(msg string) bool {
	for _, entry := range m.entries {
		if strings.Contains(entry.Msg, msg) {
			return true
		}
	}
	return false
}

// String returns a string representation of all log entries
func (m *MockLogger) String() string {
	var builder strings.Builder
	for i, entry := range m.entries {
		builder.WriteString(fmt.Sprintf("[%s] %s", entry.Level, entry.Msg))
		if entry.Fields != nil && len(entry.Fields) > 0 {
			builder.WriteString(" - Fields: ")
			for k, v := range entry.Fields {
				builder.WriteString(fmt.Sprintf("%s=%v ", k, v))
			}
		}
		if i < len(m.entries)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

// Clear clears all log entries
func (m *MockLogger) Clear() {
	m.entries = make([]LogEntry, 0)
}

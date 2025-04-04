package event

import "fmt"

// Logger interface for event system logging
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Opts(options map[string]interface{})
}

// defaultLogger is a simple default logger implementation
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, fields map[string]interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

func (l *defaultLogger) Info(msg string, fields map[string]interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *defaultLogger) Warn(msg string, fields map[string]interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, fields)
}

func (l *defaultLogger) Error(msg string, fields map[string]interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *defaultLogger) Opts(options map[string]interface{}) {
	// No-op for default logger
}

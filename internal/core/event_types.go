package core

// SystemEventType defines types of system events
type SystemEventType string

const (
	EventTypeInitialize SystemEventType = "system.initialize"
	EventTypeShutdown   SystemEventType = "system.shutdown"
	EventTypeError      SystemEventType = "system.error"
	EventTypeTimer      SystemEventType = "system.timer"
)

package bperrors

import (
	"fmt"
	"time"
)

// ErrorType defines categories of errors that can occur in the system
type ErrorType string

const (
	ErrorTypeExecution  ErrorType = "execution"  // Errors during blueprint execution
	ErrorTypeConnection ErrorType = "connection" // Errors related to node connections
	ErrorTypeValidation ErrorType = "validation" // Errors validating blueprint structure
	ErrorTypePermission ErrorType = "permission" // Errors related to permissions
	ErrorTypeDatabase   ErrorType = "database"   // Errors interacting with the database
	ErrorTypeNetwork    ErrorType = "network"    // Network-related errors
	ErrorTypePlugin     ErrorType = "plugin"     // Plugin-related errors
	ErrorTypeSystem     ErrorType = "system"     // System-level errors
	ErrorTypeUnknown    ErrorType = "unknown"    // Unclassified errors
)

// ErrorSeverity defines the severity level of errors
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical" // System-breaking errors
	SeverityHigh     ErrorSeverity = "high"     // Errors that prevent operation but not system-breaking
	SeverityMedium   ErrorSeverity = "medium"   // Errors that affect functionality but allow continued operation
	SeverityLow      ErrorSeverity = "low"      // Errors that are minor and don't significantly affect operation
	SeverityInfo     ErrorSeverity = "info"     // Informational errors
)

// BlueprintErrorCode standardizes error codes across the system
type BlueprintErrorCode string

const (
	// Execution errors
	ErrNodeExecutionFailed   BlueprintErrorCode = "E001"
	ErrNodeNotFound          BlueprintErrorCode = "E002"
	ErrNodeTypeNotRegistered BlueprintErrorCode = "E003"
	ErrExecutionTimeout      BlueprintErrorCode = "E004"
	ErrExecutionCancelled    BlueprintErrorCode = "E005"
	ErrNoEntryPoints         BlueprintErrorCode = "E006"

	// Connection errors
	ErrInvalidConnection    BlueprintErrorCode = "C001"
	ErrCircularDependency   BlueprintErrorCode = "C002"
	ErrMissingRequiredInput BlueprintErrorCode = "C003"
	ErrTypeMismatch         BlueprintErrorCode = "C004"
	ErrNodeDisconnected     BlueprintErrorCode = "C005"

	// Validation errors
	ErrInvalidBlueprintStructure BlueprintErrorCode = "V001"
	ErrInvalidNodeConfiguration  BlueprintErrorCode = "V002"
	ErrMissingProperty           BlueprintErrorCode = "V003"
	ErrInvalidPropertyValue      BlueprintErrorCode = "V004"

	// Database errors
	ErrDatabaseConnection       BlueprintErrorCode = "D001"
	ErrBlueprintNotFound        BlueprintErrorCode = "D002"
	ErrBlueprintVersionNotFound BlueprintErrorCode = "D003"
	ErrDatabaseQuery            BlueprintErrorCode = "D004"

	// System errors
	ErrInternalServerError BlueprintErrorCode = "S001"
	ErrResourceExhausted   BlueprintErrorCode = "S002"
	ErrSystemUnavailable   BlueprintErrorCode = "S003"

	// Other error codes
	ErrUnknown BlueprintErrorCode = "U001"
)

// RecoveryStrategy defines possible error recovery strategies
type RecoveryStrategy string

const (
	RecoveryRetry              RecoveryStrategy = "retry"             // Retry the operation
	RecoverySkipNode           RecoveryStrategy = "skip_node"         // Skip the problematic node
	RecoveryUseDefaultValue    RecoveryStrategy = "use_default_value" // Use a default value for the operation
	RecoveryManualIntervention RecoveryStrategy = "manual"            // Require manual intervention
	RecoveryNone               RecoveryStrategy = "none"              // No recovery possible
)

// BlueprintError represents a structured error with metadata for better diagnostics
type BlueprintError struct {
	Type            ErrorType              `json:"type"`
	Code            BlueprintErrorCode     `json:"code"`
	Message         string                 `json:"message"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Severity        ErrorSeverity          `json:"severity"`
	Recoverable     bool                   `json:"recoverable"`
	RecoveryOptions []RecoveryStrategy     `json:"recoveryOptions,omitempty"`
	NodeID          string                 `json:"nodeId,omitempty"`
	PinID           string                 `json:"pinId,omitempty"`
	BlueprintID     string                 `json:"blueprintId,omitempty"`
	ExecutionID     string                 `json:"executionId,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	OriginalError   error                  `json:"-"`
	StackTrace      []string               `json:"stackTrace,omitempty"`
}

// Error implements the error interface
func (e *BlueprintError) Error() string {
	if e.NodeID != "" && e.PinID != "" {
		return fmt.Sprintf("[%s-%s] %s: %s (Node: %s, Pin: %s)",
			e.Type, e.Code, e.Severity, e.Message, e.NodeID, e.PinID)
	} else if e.NodeID != "" {
		return fmt.Sprintf("[%s-%s] %s: %s (Node: %s)",
			e.Type, e.Code, e.Severity, e.Message, e.NodeID)
	}
	return fmt.Sprintf("[%s-%s] %s: %s", e.Type, e.Code, e.Severity, e.Message)
}

// Unwrap returns the original error
func (e *BlueprintError) Unwrap() error {
	return e.OriginalError
}

// WithDetails adds details to the error
func (e *BlueprintError) WithDetails(details map[string]interface{}) *BlueprintError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithNodeInfo adds node information to the error
func (e *BlueprintError) WithNodeInfo(nodeID, pinID string) *BlueprintError {
	e.NodeID = nodeID
	if pinID != "" {
		e.PinID = pinID
	}
	return e
}

// WithBlueprintInfo adds blueprint information to the error
func (e *BlueprintError) WithBlueprintInfo(blueprintID, executionID string) *BlueprintError {
	e.BlueprintID = blueprintID
	if executionID != "" {
		e.ExecutionID = executionID
	}
	return e
}

// WithRecoveryOptions sets recovery options for the error
func (e *BlueprintError) WithRecoveryOptions(options ...RecoveryStrategy) *BlueprintError {
	e.RecoveryOptions = options
	e.Recoverable = len(options) > 0
	return e
}

// New creates a new BlueprintError
func New(errType ErrorType, code BlueprintErrorCode, message string, severity ErrorSeverity) *BlueprintError {
	return &BlueprintError{
		Type:        errType,
		Code:        code,
		Message:     message,
		Severity:    severity,
		Timestamp:   time.Now(),
		Recoverable: false,
		Details:     make(map[string]interface{}),
	}
}

// Wrap wraps an error in a BlueprintError
func Wrap(err error, errType ErrorType, code BlueprintErrorCode, message string, severity ErrorSeverity) *BlueprintError {
	// If it's already a BlueprintError, just update fields that are provided
	if bpErr, ok := err.(*BlueprintError); ok {
		if errType != "" {
			bpErr.Type = errType
		}
		if code != "" {
			bpErr.Code = code
		}
		if message != "" {
			bpErr.Message = message
		}
		if severity != "" {
			bpErr.Severity = severity
		}
		return bpErr
	}

	return &BlueprintError{
		Type:          errType,
		Code:          code,
		Message:       message,
		Severity:      severity,
		Timestamp:     time.Now(),
		OriginalError: err,
		Details:       make(map[string]interface{}),
	}
}

type ValidationResult struct {
	Valid      bool                `json:"valid"`
	Errors     []*BlueprintError   `json:"errors,omitempty"`
	Warnings   []*BlueprintError   `json:"warnings,omitempty"`
	NodeIssues map[string][]string `json:"nodeIssues,omitempty"`
}

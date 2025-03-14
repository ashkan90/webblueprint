package common

import "time"

// ExecutionResult represents the basic result of blueprint execution
type ExecutionResult struct {
	Success        bool                              `json:"success"`
	ExecutionID    string                            `json:"executionId"`
	StartTime      time.Time                         `json:"startTime"`
	EndTime        time.Time                         `json:"endTime"`
	Error          error                             `json:"error,omitempty"`
	NodeResults    map[string]map[string]interface{} `json:"nodeResults,omitempty"` // NodeID -> PinID -> Value
	ErrorAnalysis  map[string]interface{}            `json:"errorAnalysis,omitempty"`
	PartialSuccess bool                              `json:"partialSuccess"`
}

// ValidationResult represents the result of a blueprint validation
type ValidationResult struct {
	Valid      bool                `json:"valid"`
	Errors     []error             `json:"errors,omitempty"`
	Warnings   []error             `json:"warnings,omitempty"`
	NodeIssues map[string][]string `json:"nodeIssues,omitempty"`
}

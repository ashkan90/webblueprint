package engine

import (
	"time"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/common"
)

// ExtendedExecutionResult represents the enhanced result of blueprint execution with error details
type ExtendedExecutionResult struct {
	// Basic result data
	Success        bool                              `json:"success"`
	PartialSuccess bool                              `json:"partialSuccess,omitempty"`
	ExecutionID    string                            `json:"executionId"`
	StartTime      time.Time                         `json:"startTime"`
	EndTime        time.Time                         `json:"endTime"`
	Error          error                             `json:"error,omitempty"`
	NodeResults    map[string]map[string]interface{} `json:"nodeResults,omitempty"` // NodeID -> PinID -> Value

	// Error analysis and diagnostics
	ErrorAnalysis     map[string]interface{}   `json:"errorAnalysis,omitempty"`
	RecoveryAttempts  []map[string]interface{} `json:"recoveryAttempts,omitempty"`
	ValidationResults *common.ValidationResult `json:"validationResults,omitempty"`
	FailedNodes       []string                 `json:"failedNodes,omitempty"`
	SuccessfulNodes   []string                 `json:"successfulNodes,omitempty"`
}

// ConvertToExtendedResult converts a basic ExecutionResult to an ExtendedExecutionResult
func ConvertToExtendedResult(basic common.ExecutionResult) ExtendedExecutionResult {
	extended := ExtendedExecutionResult{
		Success:     basic.Success,
		ExecutionID: basic.ExecutionID,
		StartTime:   basic.StartTime,
		EndTime:     basic.EndTime,
		NodeResults: basic.NodeResults,
	}

	// Handle error conversion
	if basic.Error != nil {
		if bpErr, ok := basic.Error.(*bperrors.BlueprintError); ok {
			extended.Error = bpErr
		} else {
			// Wrap standard error as BlueprintError
			extended.Error = bperrors.Wrap(
				basic.Error,
				bperrors.ErrorTypeExecution,
				bperrors.ErrUnknown,
				basic.Error.Error(),
				bperrors.SeverityHigh,
			)
		}
	}

	return extended
}

// ToBasicResult converts an ExtendedExecutionResult to a standard ExecutionResult
func (e *ExtendedExecutionResult) ToBasicResult() common.ExecutionResult {
	return common.ExecutionResult{
		Success:     e.Success,
		ExecutionID: e.ExecutionID,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		Error:       e.Error,
		NodeResults: e.NodeResults,
	}
}

// AddErrorAnalysis adds error analysis data to the result
func (e *ExtendedExecutionResult) AddErrorAnalysis(analysis map[string]interface{}) {
	e.ErrorAnalysis = analysis

	// Update PartialSuccess flag
	if !e.Success && analysis != nil {
		if errorCount, ok := analysis["totalErrors"].(int); ok && errorCount > 0 {
			// If we have recoverable errors but execution completed
			if recoverableCount, ok := analysis["recoverableErrors"].(int); ok {
				if recoverableCount > 0 && recoverableCount == errorCount {
					e.PartialSuccess = true
				}
			}
		}
	}
}

// AddRecoveryAttempt adds a recovery attempt to the result
func (e *ExtendedExecutionResult) AddRecoveryAttempt(attempt map[string]interface{}) {
	if e.RecoveryAttempts == nil {
		e.RecoveryAttempts = make([]map[string]interface{}, 0)
	}
	e.RecoveryAttempts = append(e.RecoveryAttempts, attempt)
}

// AddValidationResults adds validation results to the execution result
func (e *ExtendedExecutionResult) AddValidationResults(validation *common.ValidationResult) {
	e.ValidationResults = validation
}

// AddNodeStatus adds a node to either the successful or failed nodes list
func (e *ExtendedExecutionResult) AddNodeStatus(nodeID string, successful bool) {
	if successful {
		if e.SuccessfulNodes == nil {
			e.SuccessfulNodes = make([]string, 0)
		}
		e.SuccessfulNodes = append(e.SuccessfulNodes, nodeID)
	} else {
		if e.FailedNodes == nil {
			e.FailedNodes = make([]string, 0)
		}
		e.FailedNodes = append(e.FailedNodes, nodeID)
	}
}

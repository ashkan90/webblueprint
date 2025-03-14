package bperrors

import (
	"errors"
	"webblueprint/internal/common"
)

// ExtendedExecutionInfo extends ExecutionResult with error analysis
type ExtendedExecutionInfo struct {
	// Base execution result
	BaseResult *common.ExecutionResult

	// Error analysis and diagnostics
	ErrorAnalysis     map[string]interface{}
	RecoveryAttempts  []RecoveryContext
	ValidationResults *common.ValidationResult
	PartialSuccess    bool
	FailedNodes       []string
	SuccessfulNodes   []string
}

// NewExtendedExecutionInfo creates a new extended execution info
func NewExtendedExecutionInfo(baseResult *common.ExecutionResult) *ExtendedExecutionInfo {
	return &ExtendedExecutionInfo{
		BaseResult:       baseResult,
		PartialSuccess:   false,
		RecoveryAttempts: make([]RecoveryContext, 0),
		FailedNodes:      make([]string, 0),
		SuccessfulNodes:  make([]string, 0),
	}
}

// AddErrorAnalysis adds error analysis information
func (e *ExtendedExecutionInfo) AddErrorAnalysis(analysis map[string]interface{}) {
	e.ErrorAnalysis = analysis

	// Update PartialSuccess flag if appropriate
	if !e.BaseResult.Success && analysis != nil {
		if totalErrors, ok := analysis["totalErrors"].(int); ok && totalErrors > 0 {
			// If we have recoverable errors but execution completed
			if recoverableCount, ok := analysis["recoverableErrors"].(int); ok {
				if recoverableCount > 0 && recoverableCount == totalErrors {
					e.PartialSuccess = true
				}
			}
		}
	}
}

// AddRecoveryAttempt adds information about a recovery attempt
func (e *ExtendedExecutionInfo) AddRecoveryAttempt(context RecoveryContext) {
	e.RecoveryAttempts = append(e.RecoveryAttempts, context)
}

// AddValidationResults adds blueprint validation results
func (e *ExtendedExecutionInfo) AddValidationResults(validation *common.ValidationResult) {
	e.ValidationResults = validation
}

// AddNodeStatus updates the node status lists
func (e *ExtendedExecutionInfo) AddNodeStatus(nodeID string, successful bool) {
	if successful {
		e.SuccessfulNodes = append(e.SuccessfulNodes, nodeID)
	} else {
		e.FailedNodes = append(e.FailedNodes, nodeID)
	}
}

// ToMap converts the extended info to a map for JSON serialization
func (e *ExtendedExecutionInfo) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"success":        e.BaseResult.Success,
		"executionId":    e.BaseResult.ExecutionID,
		"startTime":      e.BaseResult.StartTime,
		"endTime":        e.BaseResult.EndTime,
		"partialSuccess": e.PartialSuccess,
	}

	// Add error if present
	if e.BaseResult.Error != nil {
		var bpErr *BlueprintError
		if errors.As(e.BaseResult.Error, &bpErr) {
			result["error"] = bpErr
		}
	}

	// Add results if present
	if e.BaseResult.NodeResults != nil {
		result["nodeResults"] = e.BaseResult.NodeResults
	}

	// Add error analysis if present
	if e.ErrorAnalysis != nil {
		result["errorAnalysis"] = e.ErrorAnalysis
	}

	// Add recovery attempts if present
	if len(e.RecoveryAttempts) > 0 {
		attempts := make([]map[string]interface{}, 0, len(e.RecoveryAttempts))
		for _, attempt := range e.RecoveryAttempts {
			attemptMap := map[string]interface{}{
				"strategy":   string(attempt.Strategy),
				"successful": attempt.Successful,
				"timestamp":  attempt.RecoveryTimestamp,
				"details":    attempt.Details,
			}

			if attempt.Error != nil {
				attemptMap["errorCode"] = string(attempt.Error.Code)
				attemptMap["nodeId"] = attempt.Error.NodeID
			}

			attempts = append(attempts, attemptMap)
		}
		result["recoveryAttempts"] = attempts
	}

	// Add validation results if present
	if e.ValidationResults != nil {
		result["validationResults"] = e.ValidationResults
	}

	// Add node status lists if present
	if len(e.FailedNodes) > 0 {
		result["failedNodes"] = e.FailedNodes
	}

	if len(e.SuccessfulNodes) > 0 {
		result["successfulNodes"] = e.SuccessfulNodes
	}

	return result
}

// ExecutionInfoStore manages extended execution information
type ExecutionInfoStore struct {
	executionInfos map[string]*ExtendedExecutionInfo
}

// NewExecutionInfoStore creates a new execution info store
func NewExecutionInfoStore() *ExecutionInfoStore {
	return &ExecutionInfoStore{
		executionInfos: make(map[string]*ExtendedExecutionInfo),
	}
}

// StoreExecutionInfo stores extended info for an execution
func (s *ExecutionInfoStore) StoreExecutionInfo(executionID string, info *ExtendedExecutionInfo) {
	s.executionInfos[executionID] = info
}

// GetExecutionInfo retrieves extended info for an execution
func (s *ExecutionInfoStore) GetExecutionInfo(executionID string) (*ExtendedExecutionInfo, bool) {
	info, exists := s.executionInfos[executionID]
	return info, exists
}

// RemoveExecutionInfo removes extended info for an execution
func (s *ExecutionInfoStore) RemoveExecutionInfo(executionID string) {
	delete(s.executionInfos, executionID)
}

// GetAllExecutionIDs gets all execution IDs with extended info
func (s *ExecutionInfoStore) GetAllExecutionIDs() []string {
	ids := make([]string, 0, len(s.executionInfos))
	for id := range s.executionInfos {
		ids = append(ids, id)
	}
	return ids
}

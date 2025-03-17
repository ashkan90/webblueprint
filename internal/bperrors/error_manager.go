package bperrors

import (
	"fmt"
	"sync"
	"time"
)

// ErrorManager handles error recording, analysis, and recovery strategies
type ErrorManager struct {
	errors             map[string][]*BlueprintError // Maps executionID to errors
	errorHandlers      map[ErrorType][]ErrorHandler
	recoveryStrategies map[BlueprintErrorCode][]RecoveryStrategy
	mutex              sync.RWMutex
}

// ErrorHandler is a function that handles specific error types
type ErrorHandler func(err *BlueprintError) error

// NewErrorManager creates a new error manager
func NewErrorManager() *ErrorManager {
	manager := &ErrorManager{
		errors:             make(map[string][]*BlueprintError),
		errorHandlers:      make(map[ErrorType][]ErrorHandler),
		recoveryStrategies: make(map[BlueprintErrorCode][]RecoveryStrategy),
	}

	// Register default recovery strategies
	manager.registerDefaultRecoveryStrategies()

	return manager
}

// registerDefaultRecoveryStrategies sets up default recovery options for various error codes
func (em *ErrorManager) registerDefaultRecoveryStrategies() {
	// Execution errors
	em.RegisterRecoveryStrategy(ErrNodeExecutionFailed, RecoveryRetry, RecoverySkipNode)
	em.RegisterRecoveryStrategy(ErrNodeNotFound, RecoverySkipNode)
	em.RegisterRecoveryStrategy(ErrNodeTypeNotRegistered, RecoverySkipNode)
	em.RegisterRecoveryStrategy(ErrExecutionTimeout, RecoveryRetry)

	// Connection errors
	em.RegisterRecoveryStrategy(ErrMissingRequiredInput, RecoveryUseDefaultValue)
	em.RegisterRecoveryStrategy(ErrTypeMismatch, RecoveryUseDefaultValue)

	// Database errors
	em.RegisterRecoveryStrategy(ErrDatabaseConnection, RecoveryRetry)
}

// RecordError records an error for a given execution
func (em *ErrorManager) RecordError(executionID string, err *BlueprintError) {
	if em.mutex.TryLock() {
		defer em.mutex.Unlock()
	}

	// If missing recovery options, add them from our registry
	if len(err.RecoveryOptions) == 0 {
		if strategies, ok := em.recoveryStrategies[err.Code]; ok {
			err.RecoveryOptions = strategies
			err.Recoverable = len(strategies) > 0
		}
	}

	// Add the error to our collection
	if _, ok := em.errors[executionID]; !ok {
		em.errors[executionID] = make([]*BlueprintError, 0)
	}
	em.errors[executionID] = append(em.errors[executionID], err)

	// Call relevant error handlers
	if handlers, ok := em.errorHandlers[err.Type]; ok {
		for _, handler := range handlers {
			// Don't propagate handler errors, just log them
			if handlerErr := handler(err); handlerErr != nil {
				fmt.Printf("Error handler failed: %s\n", handlerErr.Error())
			}
		}
	}
}

// GetErrors returns all errors for an execution
func (em *ErrorManager) GetErrors(executionID string) []*BlueprintError {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if errors, ok := em.errors[executionID]; ok {
		result := make([]*BlueprintError, len(errors))
		copy(result, errors)
		return result
	}
	return []*BlueprintError{}
}

// GetNodeErrors returns errors for a specific node in an execution
func (em *ErrorManager) GetNodeErrors(executionID, nodeID string) []*BlueprintError {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	errors := em.errors[executionID]
	result := make([]*BlueprintError, 0)

	for _, err := range errors {
		if err.NodeID == nodeID {
			result = append(result, err)
		}
	}

	return result
}

// ClearErrors removes all errors for an execution
func (em *ErrorManager) ClearErrors(executionID string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	delete(em.errors, executionID)
}

// RegisterErrorHandler adds a handler for a specific error type
func (em *ErrorManager) RegisterErrorHandler(errType ErrorType, handler ErrorHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, ok := em.errorHandlers[errType]; !ok {
		em.errorHandlers[errType] = make([]ErrorHandler, 0)
	}
	em.errorHandlers[errType] = append(em.errorHandlers[errType], handler)
}

// RegisterRecoveryStrategy registers possible recovery strategies for an error code
func (em *ErrorManager) RegisterRecoveryStrategy(code BlueprintErrorCode, strategies ...RecoveryStrategy) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.recoveryStrategies[code] = strategies
}

// GetRecoveryStrategies gets possible recovery strategies for an error
func (em *ErrorManager) GetRecoveryStrategies(err *BlueprintError) []RecoveryStrategy {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if strategies, ok := em.recoveryStrategies[err.Code]; ok {
		result := make([]RecoveryStrategy, len(strategies))
		copy(result, strategies)
		return result
	}
	return []RecoveryStrategy{RecoveryNone}
}

// AttemptRecovery tries to recover from an error using the specified strategy
func (em *ErrorManager) AttemptRecovery(err *BlueprintError, strategy RecoveryStrategy) (bool, map[string]interface{}) {
	// Check if the error is recoverable
	if !err.Recoverable {
		return false, nil
	}

	// Check if the strategy is valid for this error
	validStrategy := false
	for _, s := range err.RecoveryOptions {
		if s == strategy {
			validStrategy = true
			break
		}
	}

	if !validStrategy {
		return false, nil
	}

	// Attempt recovery based on the strategy
	switch strategy {
	case RecoveryUseDefaultValue:
		// Return a default value based on the error context
		return true, map[string]interface{}{
			"recoveryType": "default_value",
			"timestamp":    time.Now(),
		}

	case RecoveryRetry:
		// Signal that a retry should be attempted
		return true, map[string]interface{}{
			"recoveryType": "retry",
			"maxRetries":   3,
			"timestamp":    time.Now(),
		}

	case RecoverySkipNode:
		// Signal that the node should be skipped
		return true, map[string]interface{}{
			"recoveryType": "skip_node",
			"nodeId":       err.NodeID,
			"timestamp":    time.Now(),
		}

	default:
		return false, nil
	}
}

// AnalyzeErrors analyzes errors for patterns and provides insights
func (em *ErrorManager) AnalyzeErrors(executionID string) map[string]interface{} {
	errors := em.GetErrors(executionID)
	if len(errors) == 0 {
		return map[string]interface{}{
			"totalErrors": 0,
		}
	}

	// Count errors by type and severity
	typeCount := make(map[string]int)
	severityCount := make(map[string]int)
	nodeErrors := make(map[string]int)
	recoverableCount := 0

	for _, err := range errors {
		typeCount[string(err.Type)]++
		severityCount[string(err.Severity)]++
		if err.NodeID != "" {
			nodeErrors[err.NodeID]++
		}
		if err.Recoverable {
			recoverableCount++
		}
	}

	// Find most problematic nodes
	type nodeErrorCount struct {
		NodeID string
		Count  int
	}
	nodeErrorList := make([]nodeErrorCount, 0, len(nodeErrors))
	for nodeID, count := range nodeErrors {
		nodeErrorList = append(nodeErrorList, nodeErrorCount{NodeID: nodeID, Count: count})
	}

	// Sort by count (simple bubble sort)
	for i := 0; i < len(nodeErrorList)-1; i++ {
		for j := 0; j < len(nodeErrorList)-i-1; j++ {
			if nodeErrorList[j].Count < nodeErrorList[j+1].Count {
				nodeErrorList[j], nodeErrorList[j+1] = nodeErrorList[j+1], nodeErrorList[j]
			}
		}
	}

	// Take top 5 problematic nodes
	topNodes := make([]map[string]interface{}, 0, 5)
	for i := 0; i < len(nodeErrorList) && i < 5; i++ {
		topNodes = append(topNodes, map[string]interface{}{
			"nodeId": nodeErrorList[i].NodeID,
			"count":  nodeErrorList[i].Count,
		})
	}

	// Most common error codes
	codeCount := make(map[string]int)
	for _, err := range errors {
		codeCount[string(err.Code)]++
	}

	// Build error analysis
	analysis := map[string]interface{}{
		"totalErrors":       len(errors),
		"recoverableErrors": recoverableCount,
		"typeBreakdown":     typeCount,
		"severityBreakdown": severityCount,
		"topProblemNodes":   topNodes,
		"mostCommonCodes":   codeCount,
		"timestamp":         time.Now(),
	}

	return analysis
}

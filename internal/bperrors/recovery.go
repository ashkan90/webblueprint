package bperrors

import (
	"fmt"
	"sync"
	"time"
	"webblueprint/internal/types"
)

// DefaultValueProvider defines functions to provide default values when recovery needs them
type DefaultValueProvider func(pinType *types.PinType) types.Value

// RecoveryContext holds information about an attempted recovery
type RecoveryContext struct {
	Error             *BlueprintError
	Strategy          RecoveryStrategy
	Successful        bool
	RecoveryTimestamp time.Time
	RetryCount        int
	DefaultValues     map[string]types.Value
	Details           map[string]interface{}
}

// RecoveryManager handles error recovery attempts
type RecoveryManager struct {
	errorManager     *ErrorManager
	defaultProviders map[string]DefaultValueProvider         // Use string key instead of pointer for better map access
	recoveryAttempts map[string]map[string][]RecoveryContext // ExecutionID -> NodeID -> recovery attempts
	mutex            sync.RWMutex
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(errorManager *ErrorManager) *RecoveryManager {
	rm := &RecoveryManager{
		errorManager:     errorManager,
		defaultProviders: make(map[string]DefaultValueProvider),
		recoveryAttempts: make(map[string]map[string][]RecoveryContext),
		mutex:            sync.RWMutex{},
	}

	// Register default value providers
	rm.registerDefaultValueProviders()

	return rm
}

// registerDefaultValueProviders sets up functions to provide default values for different types
func (rm *RecoveryManager) registerDefaultValueProviders() {
	// String default provider
	rm.RegisterDefaultValueProvider("string", func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.String, "")
	})

	// Number default provider
	rm.RegisterDefaultValueProvider("number", func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Number, 0)
	})

	// Boolean default provider
	rm.RegisterDefaultValueProvider("boolean", func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Boolean, false)
	})

	// Array default provider
	rm.RegisterDefaultValueProvider("array", func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Array, []interface{}{})
	})

	// Object default provider
	rm.RegisterDefaultValueProvider("object", func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Object, map[string]interface{}{})
	})

	// Any default provider
	rm.RegisterDefaultValueProvider("any", func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Any, nil)
	})
}

// RegisterDefaultValueProvider registers a function to provide default values for a pin type
func (rm *RecoveryManager) RegisterDefaultValueProvider(typeName string, provider DefaultValueProvider) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.defaultProviders[typeName] = provider
}

// RecoverFromError attempts to recover from an error using an appropriate strategy
func (rm *RecoveryManager) RecoverFromError(executionID string, err *BlueprintError) (bool, map[string]interface{}) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Get available recovery strategies for this error
	strategies := rm.errorManager.GetRecoveryStrategies(err)
	if len(strategies) == 0 || strategies[0] == RecoveryNone {
		return false, nil
	}

	// Choose the best recovery strategy
	// (in real implementation, this could use more sophisticated logic)
	chosenStrategy := strategies[0]

	// Check if we've already tried this strategy too many times
	if rm.tooManyAttempts(executionID, err.NodeID, chosenStrategy) {
		return false, map[string]interface{}{
			"reason":      "too_many_attempts",
			"maxAttempts": 3,
		}
	}

	// Attempt recovery
	success, details := rm.errorManager.AttemptRecovery(err, chosenStrategy)

	// Record the attempt
	rc := RecoveryContext{
		Error:             err,
		Strategy:          chosenStrategy,
		Successful:        success,
		RecoveryTimestamp: time.Now(),
		Details:           details,
	}

	// Initialize map structure if needed
	if _, ok := rm.recoveryAttempts[executionID]; !ok {
		rm.recoveryAttempts[executionID] = make(map[string][]RecoveryContext)
	}
	if _, ok := rm.recoveryAttempts[executionID][err.NodeID]; !ok {
		rm.recoveryAttempts[executionID][err.NodeID] = make([]RecoveryContext, 0)
	}

	// Add to recovery attempts
	rm.recoveryAttempts[executionID][err.NodeID] = append(
		rm.recoveryAttempts[executionID][err.NodeID],
		rc,
	)

	return success, details
}

// GetDefaultValue provides a default value for a pin type (used during recovery)
func (rm *RecoveryManager) GetDefaultValue(pinType *types.PinType) (types.Value, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// Get type name as string key
	typeName := string(pinType.Name)

	if provider, ok := rm.defaultProviders[typeName]; ok {
		return provider(pinType), nil
	}

	// If no specific provider, try to use the Any provider
	if provider, ok := rm.defaultProviders["any"]; ok {
		return provider(pinType), nil
	}

	return types.Value{}, fmt.Errorf("no default value provider for pin type %v", pinType)
}

// GetRecoveryAttempts gets recovery attempts for a node in an execution
func (rm *RecoveryManager) GetRecoveryAttempts(executionID, nodeID string) []RecoveryContext {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if attempts, ok := rm.recoveryAttempts[executionID]; ok {
		if nodeAttempts, ok := attempts[nodeID]; ok {
			return nodeAttempts
		}
	}
	return []RecoveryContext{}
}

// ClearRecoveryAttempts clears all recovery attempts for an execution
func (rm *RecoveryManager) ClearRecoveryAttempts(executionID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.recoveryAttempts, executionID)
}

// CountRecoveryAttempts counts how many times a specific recovery strategy has been attempted
func (rm *RecoveryManager) CountRecoveryAttempts(executionID, nodeID string, strategy RecoveryStrategy) int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	count := 0
	attempts := rm.GetRecoveryAttempts(executionID, nodeID)

	for _, attempt := range attempts {
		if attempt.Strategy == strategy {
			count++
		}
	}

	return count
}

// tooManyAttempts checks if a recovery strategy has been tried too many times
func (rm *RecoveryManager) tooManyAttempts(executionID, nodeID string, strategy RecoveryStrategy) bool {
	count := rm.CountRecoveryAttempts(executionID, nodeID, strategy)

	// Limit each strategy to 3 attempts
	// In a real implementation, different strategies might have different limits
	return count >= 3
}

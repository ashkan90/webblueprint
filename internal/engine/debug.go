package engine

import (
	"sync"
	"time"
)

// DebugManager stores and manages debug information during execution
type DebugManager struct {
	// Maps: executionID -> nodeID -> data
	debugData map[string]map[string]map[string]interface{}

	// Maps: executionID -> nodeID -> pinID -> value
	outputValues map[string]map[string]map[string]interface{}

	mutex sync.RWMutex
}

// NewDebugManager creates a new debug manager
func NewDebugManager() *DebugManager {
	return &DebugManager{
		debugData:    make(map[string]map[string]map[string]interface{}),
		outputValues: make(map[string]map[string]map[string]interface{}),
	}
}

// StoreNodeDebugData stores debug data for a node
func (dm *DebugManager) StoreNodeDebugData(executionID, nodeID string, data map[string]interface{}) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Initialize maps if needed
	if _, exists := dm.debugData[executionID]; !exists {
		dm.debugData[executionID] = make(map[string]map[string]interface{})
	}

	if _, exists := dm.debugData[executionID][nodeID]; !exists {
		dm.debugData[executionID][nodeID] = make(map[string]interface{})
	}

	// Store data with timestamp
	for key, value := range data {
		dm.debugData[executionID][nodeID][key] = map[string]interface{}{
			"value":     value,
			"timestamp": time.Now(),
		}
	}
}

// StoreNodeOutputValue stores an output value for a node
func (dm *DebugManager) StoreNodeOutputValue(executionID, nodeID, pinID string, value interface{}) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Initialize maps if needed
	if _, exists := dm.outputValues[executionID]; !exists {
		dm.outputValues[executionID] = make(map[string]map[string]interface{})
	}

	if _, exists := dm.outputValues[executionID][nodeID]; !exists {
		dm.outputValues[executionID][nodeID] = make(map[string]interface{})
	}

	// Store the value
	dm.outputValues[executionID][nodeID][pinID] = value
}

// GetNodeDebugData retrieves debug data for a node
func (dm *DebugManager) GetNodeDebugData(executionID, nodeID string) (map[string]interface{}, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if execData, exists := dm.debugData[executionID]; exists {
		if nodeData, exists := execData[nodeID]; exists {
			return nodeData, true
		}
	}

	return nil, false
}

// GetNodeOutputValue retrieves an output value for a node
func (dm *DebugManager) GetNodeOutputValue(executionID, nodeID, pinID string) (interface{}, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if execData, exists := dm.outputValues[executionID]; exists {
		if nodeData, exists := execData[nodeID]; exists {
			if value, exists := nodeData[pinID]; exists {
				return value, true
			}
		}
	}

	return nil, false
}

// GetAllNodeOutputValues retrieves all output values for a node
func (dm *DebugManager) GetAllNodeOutputValues(executionID, nodeID string) (map[string]interface{}, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if execData, exists := dm.outputValues[executionID]; exists {
		if nodeData, exists := execData[nodeID]; exists {
			return nodeData, true
		}
	}

	return nil, false
}

// GetExecutionDebugData retrieves all debug data for an execution
func (dm *DebugManager) GetExecutionDebugData(executionID string) map[string]map[string]interface{} {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if execData, exists := dm.debugData[executionID]; exists {
		// Create a copy to avoid race conditions
		result := make(map[string]map[string]interface{})
		for nodeID, nodeData := range execData {
			result[nodeID] = make(map[string]interface{})
			for key, value := range nodeData {
				result[nodeID][key] = value
			}
		}
		return result
	}

	return nil
}

// GetExecutionOutputValues retrieves all output values for an execution
func (dm *DebugManager) GetExecutionOutputValues(executionID string) map[string]map[string]interface{} {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if execData, exists := dm.outputValues[executionID]; exists {
		// Create a copy to avoid race conditions
		result := make(map[string]map[string]interface{})
		for nodeID, nodeData := range execData {
			result[nodeID] = make(map[string]interface{})
			for key, value := range nodeData {
				result[nodeID][key] = value
			}
		}
		return result
	}

	return nil
}

// ClearExecutionData removes all data for an execution
func (dm *DebugManager) ClearExecutionData(executionID string) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	delete(dm.debugData, executionID)
	delete(dm.outputValues, executionID)
}

// ClearAllData removes all debug data
func (dm *DebugManager) ClearAllData() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.debugData = make(map[string]map[string]map[string]interface{})
	dm.outputValues = make(map[string]map[string]map[string]interface{})
}

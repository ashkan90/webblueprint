package engine

import (
	"fmt"
	"sync"
	"time"
)

// DebugManager stores and manages debug information during execution
type DebugManager struct {
	// Maps: executionID -> nodeID -> data
	debugData map[string]map[string]map[string]interface{}

	// Maps: executionID -> nodeID -> pinID -> value
	outputValues map[string]map[string]map[string]interface{}

	// Maps: executionID -> data
	executionData map[string]map[string]interface{}

	mutex sync.RWMutex
}

// NewDebugManager creates a new debug manager
func NewDebugManager() *DebugManager {
	return &DebugManager{
		debugData:     make(map[string]map[string]map[string]interface{}),
		outputValues:  make(map[string]map[string]map[string]interface{}),
		executionData: make(map[string]map[string]interface{}),
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

// StoreExecutionDebugData stores debug data for an entire execution
func (dm *DebugManager) StoreExecutionDebugData(executionID string, data map[string]interface{}) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Initialize map if needed
	if _, exists := dm.executionData[executionID]; !exists {
		dm.executionData[executionID] = make(map[string]interface{})
	}

	// Store data with timestamp
	for key, value := range data {
		dm.executionData[executionID][key] = map[string]interface{}{
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

	// Log for debugging
	fmt.Printf("[DEBUG] Stored output value for %s.%s: %v (type: %T)\n",
		nodeID, pinID, value, value)

	// Dump the entire outputValues map structure for debugging
	fmt.Printf("[DEBUG] Current outputValues structure for execution %s:\n", executionID)
	for n, pins := range dm.outputValues[executionID] {
		fmt.Printf("  Node %s:\n", n)
		for p, v := range pins {
			fmt.Printf("    Pin %s: %v (type: %T)\n", p, v, v)
		}
	}
}

// ClearCachedPinTypes clears any cached type information for a pin
func (dm *DebugManager) ClearCachedPinTypes(executionID, nodeID, pinID string) {
	// This is a placeholder for future type caching optimization
	// Currently, we don't cache pin types, but this method ensures we have a hook
	// for clearing type information if we implement caching in the future

	// Ensure the value is fresh by re-fetching it if needed
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	// In a more advanced implementation, we might clear a type cache here
	// For now, this is just a placeholder
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

// GetExecutionDebugData retrieves debug data for an execution
func (dm *DebugManager) GetExecutionDebugData(executionID string) map[string]interface{} {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if execData, exists := dm.executionData[executionID]; exists {
		// Create a copy to avoid race conditions
		result := make(map[string]interface{})
		for key, value := range execData {
			result[key] = value
		}
		return result
	}

	return nil
}

// GetNodeOutputValue retrieves an output value for a node
func (dm *DebugManager) GetNodeOutputValue(executionID, nodeID, pinID string) (interface{}, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	fmt.Printf("[DEBUG] Attempting to get output value for %s.%s in execution %s\n",
		nodeID, pinID, executionID)

	// Dump the entire outputValues map structure for debugging
	fmt.Printf("[DEBUG] Current outputValues structure for execution %s:\n", executionID)
	if execData, exists := dm.outputValues[executionID]; exists {
		for n, pins := range execData {
			fmt.Printf("  Node %s:\n", n)
			for p, v := range pins {
				fmt.Printf("    Pin %s: %v (type: %T)\n", p, v, v)
			}
		}
	} else {
		fmt.Printf("  No data found for execution %s\n", executionID)
	}

	if execData, exists := dm.outputValues[executionID]; exists {
		if nodeData, exists := execData[nodeID]; exists {
			if value, exists := nodeData[pinID]; exists {
				fmt.Printf("[DEBUG] Retrieved output value for %s.%s: %v (type: %T)\n",
					nodeID, pinID, value, value)
				return value, true
			} else {
				fmt.Printf("[DEBUG] No value found for pin %s in node %s\n", pinID, nodeID)
			}
		} else {
			fmt.Printf("[DEBUG] No data found for node %s in execution %s\n", nodeID, executionID)
		}
	} else {
		fmt.Printf("[DEBUG] No data found for execution %s\n", executionID)
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

// GetAllNodeDebugData retrieves all debug data for all nodes in an execution
func (dm *DebugManager) GetAllNodeDebugData(executionID string) map[string]map[string]interface{} {
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
	delete(dm.executionData, executionID)
}

// ClearAllData removes all debug data
func (dm *DebugManager) ClearAllData() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.debugData = make(map[string]map[string]map[string]interface{})
	dm.outputValues = make(map[string]map[string]map[string]interface{})
	dm.executionData = make(map[string]map[string]interface{})
}

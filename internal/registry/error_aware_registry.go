package registry

import (
	errors "webblueprint/internal/bperrors"
	"webblueprint/internal/node"
)

// ErrorAwareNodeRegistry extends the standard node registry with error handling capabilities
type ErrorAwareNodeRegistry struct {
	*GlobalNodeRegistry
	errorManager    *errors.ErrorManager
	recoveryManager *errors.RecoveryManager
}

// NewErrorAwareNodeRegistry creates a new error-aware node registry
func NewErrorAwareNodeRegistry(registry *GlobalNodeRegistry, errorManager *errors.ErrorManager, recoveryManager *errors.RecoveryManager) *ErrorAwareNodeRegistry {
	return &ErrorAwareNodeRegistry{
		GlobalNodeRegistry: registry,
		errorManager:       errorManager,
		recoveryManager:    recoveryManager,
	}
}

// RegisterErrorAwareNode registers a node with error handling capabilities
func (r *ErrorAwareNodeRegistry) RegisterErrorAwareNode(nodeID string, factory node.NodeFactory) {
	// Register the node as usual
	r.GlobalNodeRegistry.RegisterNodeType(nodeID, factory)

	// You could add special handling for error-aware nodes here
	// For example, logging, validation, or adding to a special category
}

// RegisterBuiltInErrorAwareNodes registers all built-in error-aware nodes
func (r *ErrorAwareNodeRegistry) RegisterBuiltInErrorAwareNodes() {
	// Register all error-aware nodes here
	// This would be a centralized place to register all error-aware node implementations

	// Example:
	// r.RegisterErrorAwareNode("http-request-with-recovery", web.NewHTTPRequestNodeWithRecovery)
	// r.RegisterErrorAwareNode("database-query-with-recovery", database.NewDatabaseQueryNodeWithRecovery)
	// r.RegisterErrorAwareNode("file-operation-with-recovery", storage.NewFileOperationNodeWithRecovery)
}

// GetErrorManager returns the error manager
func (r *ErrorAwareNodeRegistry) GetErrorManager() *errors.ErrorManager {
	return r.errorManager
}

// GetRecoveryManager returns the recovery manager
func (r *ErrorAwareNodeRegistry) GetRecoveryManager() *errors.RecoveryManager {
	return r.recoveryManager
}

// IsErrorAwareNode checks if a node type has error handling capabilities
func (r *ErrorAwareNodeRegistry) IsErrorAwareNode(nodeType string) bool {
	// In a real implementation, you might maintain a list of error-aware node types
	// or check for a specific interface implementation

	// For this example, we'll just check for nodes with "recovery" in their name
	return contains(nodeType, "recovery") || contains(nodeType, "error-handling")
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package engineext

import (
	"webblueprint/internal/node"
)

// Helper interface to check if a context wraps another (assuming decorators implement this)
type contextWrapper interface {
	Unwrap() node.ExecutionContext
}

// GetExtendedContext attempts to unwrap decorators from a given ExecutionContext
// to find and return the underlying context that implements ExtendedExecutionContext.
// Returns nil if not found after unwrapping.
func GetExtendedContext(ctx node.ExecutionContext) node.ExtendedExecutionContext {
	if ctx == nil {
		return nil
	}

	currentCtx := ctx
	// Loop to unwrap decorators (limit depth to prevent infinite loops)
	for i := 0; i < 10; i++ {
		// Check if the current context implements the target interface
		if extCtx, ok := currentCtx.(node.ExtendedExecutionContext); ok {
			return extCtx // Found it
		}

		// Try to unwrap using the contextWrapper interface
		if wrapper, ok := currentCtx.(contextWrapper); ok {
			currentCtx = wrapper.Unwrap()
			if currentCtx == nil {
				return nil // Stop if unwrapping leads to nil
			}
		} else {
			// Cannot unwrap further
			return nil
		}
	}
	// Exceeded unwrap depth or couldn't find the interface
	return nil
}

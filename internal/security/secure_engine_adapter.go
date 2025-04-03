package security

import (
	"errors"
	"fmt"
	"sync"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// SecureEngineAdapter adapts the execution engine to use security features
type SecureEngineAdapter struct {
	securityManager  *SecurityManager
	contextCache     map[string]*SandboxedExecutionContext
	userSessionCache map[string]map[string]interface{} // userID -> sessionData
	mutex            sync.RWMutex
}

// NewSecureEngineAdapter creates a new secure engine adapter
func NewSecureEngineAdapter() *SecureEngineAdapter {
	return &SecureEngineAdapter{
		securityManager:  NewSecurityManager(),
		contextCache:     make(map[string]*SandboxedExecutionContext),
		userSessionCache: make(map[string]map[string]interface{}),
	}
}

// WrapExecutionContext wraps an execution context with security features
func (adapter *SecureEngineAdapter) WrapExecutionContext(
	baseCtx node.ExecutionContext,
	userID, blueprintID, executionID string,
) node.ExecutionContext {
	// Get sandbox options for this user and blueprint
	options := adapter.securityManager.GetSandboxOptions(userID, blueprintID)

	// Create sandboxed context
	ctx := NewSandboxedExecutionContext(
		baseCtx,
		userID,
		blueprintID,
		executionID,
		options,
	)

	// Cache the context for later reference
	adapter.mutex.Lock()
	cacheKey := fmt.Sprintf("%s:%s:%s", userID, blueprintID, executionID)
	adapter.contextCache[cacheKey] = ctx
	adapter.mutex.Unlock()

	return ctx
}

// GetCachedContext retrieves a cached execution context
func (adapter *SecureEngineAdapter) GetCachedContext(userID, blueprintID, executionID string) (*SandboxedExecutionContext, bool) {
	adapter.mutex.RLock()
	defer adapter.mutex.RUnlock()

	cacheKey := fmt.Sprintf("%s:%s:%s", userID, blueprintID, executionID)
	ctx, exists := adapter.contextCache[cacheKey]
	return ctx, exists
}

// RemoveCachedContext removes a context from the cache and cleans up resources
func (adapter *SecureEngineAdapter) RemoveCachedContext(userID, blueprintID, executionID string) {
	adapter.mutex.Lock()
	defer adapter.mutex.Unlock()

	cacheKey := fmt.Sprintf("%s:%s:%s", userID, blueprintID, executionID)
	if ctx, exists := adapter.contextCache[cacheKey]; exists {
		ctx.Cleanup()
		delete(adapter.contextCache, cacheKey)
	}
}

// CheckSecurityPreExecution performs comprehensive security checks before executing a blueprint
func (adapter *SecureEngineAdapter) CheckSecurityPreExecution(
	bp *blueprint.Blueprint,
	userID string,
) error {
	// Get user permission level
	permLevel := adapter.securityManager.GetUserPermission(userID)

	// Get blueprint permission level (overrides user if set)
	if bpPerm := adapter.securityManager.GetBlueprintPermission(bp.ID); bpPerm != "" {
		permLevel = bpPerm
	}

	// Create security checker with this permission level
	checker := NewSecurityChecker(permLevel)

	// Check if rate limited
	rateLimitKey := RateLimitKey(fmt.Sprintf("%s:%s", userID, bp.ID))
	if err := adapter.securityManager.CheckRateLimit(RateLimitTypeUserAndBlueprint, rateLimitKey); err != nil {
		return fmt.Errorf("rate limit exceeded: %v", err)
	}

	// Check for potentially dangerous nodes
	for _, nodeConfig := range bp.Nodes {
		if !checker.IsNodeAllowed(nodeConfig.Type, permLevel) {
			return fmt.Errorf("blueprint contains restricted node type: %s", nodeConfig.Type)
		}

		// Additional checks for specific node types
		if nodeConfig.Type == "http-request" || nodeConfig.Type == "websocket" {
			// Check for URL properties
			for _, prop := range nodeConfig.Properties {
				if prop.Name == "url" || prop.Name == "endpoint" {
					if url, ok := prop.Value.(string); ok {
						if !checker.IsNetworkRequestAllowed(url, "GET", permLevel) {
							return fmt.Errorf("blueprint contains disallowed network access to: %s", url)
						}
					}
				}
			}
		} else if nodeConfig.Type == "file-read" || nodeConfig.Type == "file-write" {
			// Check for file path properties
			for _, prop := range nodeConfig.Properties {
				if prop.Name == "path" || prop.Name == "filePath" {
					if path, ok := prop.Value.(string); ok {
						if !checker.IsFilePathAllowed(path) {
							return fmt.Errorf("blueprint contains disallowed file access to: %s", path)
						}
					}
				}
			}
		} else if nodeConfig.Type == "execute-command" {
			// Check for command properties
			for _, prop := range nodeConfig.Properties {
				if prop.Name == "command" {
					if cmd, ok := prop.Value.(string); ok {
						if !checker.IsCommandExecutionAllowed(cmd, permLevel) {
							return fmt.Errorf("blueprint contains disallowed command execution: %s", cmd)
						}
					}
				}
			}
		}
	}

	return nil
}

// GetSecurityManager returns the security manager
func (adapter *SecureEngineAdapter) GetSecurityManager() *SecurityManager {
	return adapter.securityManager
}

// CheckBlueprintVariables validates blueprint variables for potential security issues
func (adapter *SecureEngineAdapter) CheckBlueprintVariables(
	bp *blueprint.Blueprint,
	variables map[string]types.Value,
) error {
	for name, value := range variables {
		// Check the variable name for patterns that might indicate sensitive data
		checker := NewSecurityChecker(PermissionLevelStandard)
		if !checker.IsDataAccessAllowed(name, PermissionLevelStandard) {
			return fmt.Errorf("suspicious variable name detected: %s", name)
		}

		// Check variable content if it's a string
		if value.Type == types.PinTypes.String {
			if strVal, ok := value.RawValue.(string); ok {
				if valid, reason := checker.ValidateContent(strVal); !valid {
					return errors.New(reason)
				}
			}
		}
	}

	return nil
}

// ClearUserSession removes all session data for a user
func (adapter *SecureEngineAdapter) ClearUserSession(userID string) {
	adapter.mutex.Lock()
	defer adapter.mutex.Unlock()

	delete(adapter.userSessionCache, userID)

	// Also clear any cached contexts for this user
	for cacheKey := range adapter.contextCache {
		if len(cacheKey) > len(userID) && cacheKey[:len(userID)] == userID && cacheKey[len(userID)] == ':' {
			if ctx := adapter.contextCache[cacheKey]; ctx != nil {
				ctx.Cleanup()
			}
			delete(adapter.contextCache, cacheKey)
		}
	}
}

// SetUserSessionData stores session data for a user
func (adapter *SecureEngineAdapter) SetUserSessionData(userID string, key string, value interface{}) {
	adapter.mutex.Lock()
	defer adapter.mutex.Unlock()

	if _, exists := adapter.userSessionCache[userID]; !exists {
		adapter.userSessionCache[userID] = make(map[string]interface{})
	}

	adapter.userSessionCache[userID][key] = value
}

// GetUserSessionData retrieves session data for a user
func (adapter *SecureEngineAdapter) GetUserSessionData(userID string, key string) (interface{}, bool) {
	adapter.mutex.RLock()
	defer adapter.mutex.RUnlock()

	if sessionData, exists := adapter.userSessionCache[userID]; exists {
		value, exists := sessionData[key]
		return value, exists
	}

	return nil, false
}

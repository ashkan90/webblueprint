package security

import (
	"fmt"
	"sync"
	"time"
)

// SecurityManager coordinates security features and provides a central interface for security operations
type SecurityManager struct {
	rateLimiter          *RateLimiter
	resourceProfiles     map[string]ResourceLimits
	defaultOptions       SandboxOptions
	userPermissions      map[string]PermissionLevel
	blueprintPermissions map[string]PermissionLevel
	mutex                sync.RWMutex
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		rateLimiter: NewRateLimiter(),
		resourceProfiles: map[string]ResourceLimits{
			string(ResourceProfileLow):     GetResourceLimits(ResourceProfileLow),
			string(ResourceProfileMedium):  GetResourceLimits(ResourceProfileMedium),
			string(ResourceProfileHigh):    GetResourceLimits(ResourceProfileHigh),
			string(ResourceProfileUnlimit): GetResourceLimits(ResourceProfileUnlimit),
		},
		defaultOptions:       DefaultSandboxOptions(),
		userPermissions:      make(map[string]PermissionLevel),
		blueprintPermissions: make(map[string]PermissionLevel),
	}
}

// GetSandboxOptions returns sandbox options for a specific user and blueprint
func (sm *SecurityManager) GetSandboxOptions(userID, blueprintID string) SandboxOptions {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Start with default options
	options := sm.defaultOptions

	// Apply user-specific permission level if defined
	if permLevel, exists := sm.userPermissions[userID]; exists {
		options.PermissionLevel = permLevel
	}

	// Apply blueprint-specific permission level if defined (takes precedence)
	if permLevel, exists := sm.blueprintPermissions[blueprintID]; exists {
		options.PermissionLevel = permLevel
	}

	// Always use the global rate limiter
	options.RateLimiter = sm.rateLimiter

	return options
}

// SetUserPermission sets the permission level for a user
func (sm *SecurityManager) SetUserPermission(userID string, permLevel PermissionLevel) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.userPermissions[userID] = permLevel
}

// SetBlueprintPermission sets the permission level for a blueprint
func (sm *SecurityManager) SetBlueprintPermission(blueprintID string, permLevel PermissionLevel) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.blueprintPermissions[blueprintID] = permLevel
}

// GetUserPermission gets the permission level for a user
func (sm *SecurityManager) GetUserPermission(userID string) PermissionLevel {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if permLevel, exists := sm.userPermissions[userID]; exists {
		return permLevel
	}

	return sm.defaultOptions.PermissionLevel
}

// GetBlueprintPermission gets the permission level for a blueprint
func (sm *SecurityManager) GetBlueprintPermission(blueprintID string) PermissionLevel {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if permLevel, exists := sm.blueprintPermissions[blueprintID]; exists {
		return permLevel
	}

	return sm.defaultOptions.PermissionLevel
}

// SetDefaultSandboxOptions updates the default sandbox options
func (sm *SecurityManager) SetDefaultSandboxOptions(options SandboxOptions) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.defaultOptions = options
}

// SetResourceProfile updates or adds a custom resource profile
func (sm *SecurityManager) SetResourceProfile(name string, limits ResourceLimits) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.resourceProfiles[name] = limits
}

// GetResourceProfile retrieves a resource profile by name
func (sm *SecurityManager) GetResourceProfile(name string) (ResourceLimits, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	limits, exists := sm.resourceProfiles[name]
	return limits, exists
}

// CheckRateLimit checks if a request is allowed based on rate limits
func (sm *SecurityManager) CheckRateLimit(limitType RateLimitType, key RateLimitKey) error {
	return sm.rateLimiter.CheckRateLimit(limitType, key)
}

// SetRateLimit updates a rate limit
func (sm *SecurityManager) SetRateLimit(limitType RateLimitType, count int, duration time.Duration) {
	sm.rateLimiter.SetRateLimit(limitType, count, duration)
}

// ResetRateLimit resets rate limiting for a specific key
func (sm *SecurityManager) ResetRateLimit(limitType RateLimitType, key RateLimitKey) {
	sm.rateLimiter.Reset(limitType, key)
}

// CheckSecurity performs a comprehensive security check
func (sm *SecurityManager) CheckSecurity(
	userID, blueprintID string,
	nodeType, operation string,
	targetResource string,
) error {
	// Get permissions
	options := sm.GetSandboxOptions(userID, blueprintID)
	permLevel := options.PermissionLevel

	// Create security checker
	checker := NewSecurityChecker(permLevel)

	// Check node permissions
	if !checker.IsNodeAllowed(nodeType, permLevel) {
		return fmt.Errorf("security violation: node type %s not allowed at permission level %s",
			nodeType, permLevel)
	}

	// Check based on operation type
	switch operation {
	case "network":
		if options.BlockNetwork {
			return fmt.Errorf("network access is blocked")
		}

		// Parse targetResource as URL
		if !checker.IsNetworkRequestAllowed(targetResource, "GET", permLevel) {
			return fmt.Errorf("network request to %s not allowed", targetResource)
		}

		// Check rate limit
		rateLimitKey := RateLimitKey(fmt.Sprintf("%s:network", userID))
		if err := sm.rateLimiter.CheckRateLimit(RateLimitTypeUser, rateLimitKey); err != nil {
			return fmt.Errorf("rate limit exceeded: %v", err)
		}

	case "file":
		if options.BlockFileSystem {
			return fmt.Errorf("file system access is blocked")
		}

		if !checker.IsFilePathAllowed(targetResource) {
			return fmt.Errorf("file access to %s not allowed", targetResource)
		}

		// Check rate limit
		rateLimitKey := RateLimitKey(fmt.Sprintf("%s:file", userID))
		if err := sm.rateLimiter.CheckRateLimit(RateLimitTypeUser, rateLimitKey); err != nil {
			return fmt.Errorf("rate limit exceeded: %v", err)
		}

	case "command":
		if options.BlockCommandExec {
			return fmt.Errorf("command execution is blocked")
		}

		if !checker.IsCommandExecutionAllowed(targetResource, permLevel) {
			return fmt.Errorf("command execution of %s not allowed", targetResource)
		}

		// Check rate limit
		rateLimitKey := RateLimitKey(fmt.Sprintf("%s:command", userID))
		if err := sm.rateLimiter.CheckRateLimit(RateLimitTypeUser, rateLimitKey); err != nil {
			return fmt.Errorf("rate limit exceeded: %v", err)
		}

	case "data":
		if !checker.IsDataAccessAllowed(targetResource, permLevel) {
			return fmt.Errorf("data access to %s not allowed", targetResource)
		}
	}

	return nil
}

// GetRateLimiter returns the rate limiter
func (sm *SecurityManager) GetRateLimiter() *RateLimiter {
	return sm.rateLimiter
}

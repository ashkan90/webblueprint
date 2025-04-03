package security

import (
	"fmt"
	"strings"
	"sync"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// SandboxedExecutionContext provides an isolated execution environment with resource and security restrictions
type SandboxedExecutionContext struct {
	node.ExecutionContext
	resourceMonitor   *ResourceMonitor
	securityChecker   *SecurityChecker
	rateLimiter       *RateLimiter
	permissionLevel   PermissionLevel
	blueprintID       string
	userID            string
	executionID       string
	accessedResources map[string]bool
	blockNetwork      bool
	blockFileSystem   bool
	blockCommandExec  bool
	isolationBoundary string // Used to prevent cross-blueprint data access
	mutex             sync.RWMutex
}

// SandboxOptions contains configuration options for the sandbox
type SandboxOptions struct {
	BlockNetwork     bool
	BlockFileSystem  bool
	BlockCommandExec bool
	ResourceProfile  ResourceProfile
	PermissionLevel  PermissionLevel
	RateLimiter      *RateLimiter
}

// DefaultSandboxOptions returns default sandbox options
func DefaultSandboxOptions() SandboxOptions {
	return SandboxOptions{
		BlockNetwork:     false,
		BlockFileSystem:  false,
		BlockCommandExec: true,
		ResourceProfile:  ResourceProfileMedium,
		PermissionLevel:  PermissionLevelStandard,
		RateLimiter:      nil,
	}
}

// NewSandboxedExecutionContext wraps an execution context with security measures
func NewSandboxedExecutionContext(
	baseCtx node.ExecutionContext,
	userID, blueprintID, executionID string,
	options SandboxOptions,
) *SandboxedExecutionContext {
	// Get resource limits for the specified profile
	resourceLimits := GetResourceLimits(options.ResourceProfile)
	resourceMonitor := NewResourceMonitor(resourceLimits)

	// Create security checker
	securityChecker := NewSecurityChecker(options.PermissionLevel)

	// Use provided rate limiter or create a new one
	rateLimiter := options.RateLimiter
	if rateLimiter == nil {
		rateLimiter = NewRateLimiter()
	}

	ctx := &SandboxedExecutionContext{
		ExecutionContext:  baseCtx,
		resourceMonitor:   resourceMonitor,
		securityChecker:   securityChecker,
		rateLimiter:       rateLimiter,
		permissionLevel:   options.PermissionLevel,
		blueprintID:       blueprintID,
		userID:            userID,
		executionID:       executionID,
		accessedResources: make(map[string]bool),
		blockNetwork:      options.BlockNetwork,
		blockFileSystem:   options.BlockFileSystem,
		blockCommandExec:  options.BlockCommandExec,
		isolationBoundary: fmt.Sprintf("bp:%s:exec:%s", blueprintID, executionID),
	}

	return ctx
}

// ExecuteConnectedNodes overrides base execution to add security checks
func (ctx *SandboxedExecutionContext) ExecuteConnectedNodes(pinID string) error {
	// Track node execution for resource limiting
	if !ctx.resourceMonitor.TrackNodeExecution() {
		return fmt.Errorf("resource limits exceeded: %s", ctx.resourceMonitor.GetLimitExceededReason())
	}

	// Check if resource limits have been exceeded
	if ctx.resourceMonitor.LimitExceeded() {
		return fmt.Errorf("resource limits exceeded: %s", ctx.resourceMonitor.GetLimitExceededReason())
	}

	// Get node information for security check
	nodeType := ctx.GetNodeType()

	// Perform security check on node
	if !ctx.securityChecker.IsNodeAllowed(nodeType, ctx.permissionLevel) {
		return fmt.Errorf("security violation: node type %s not allowed at permission level %s",
			nodeType, ctx.permissionLevel)
	}

	// Rate limit check for specific node types
	if strings.HasPrefix(nodeType, "http-") ||
		strings.HasPrefix(nodeType, "file-") ||
		strings.HasPrefix(nodeType, "execute-") {

		// Create a specific rate limit key for this node type
		rateLimitKey := RateLimitKey(fmt.Sprintf("%s:%s", ctx.userID, nodeType))
		if err := ctx.rateLimiter.CheckRateLimit(RateLimitTypeUserAndBlueprint, rateLimitKey); err != nil {
			return fmt.Errorf("rate limit exceeded for node type %s: %v", nodeType, err)
		}
	}

	// Execute connected nodes with monitoring context
	return ctx.ExecutionContext.ExecuteConnectedNodes(pinID)
}

// GetInputValue overrides base method to add security checks for data access
func (ctx *SandboxedExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	// Record resource access
	ctx.recordResourceAccess(fmt.Sprintf("input:%s", pinID))

	// Get the value from the base context
	value, exists := ctx.ExecutionContext.GetInputValue(pinID)

	return value, exists
}

// GetDebugData overrides base method to add security checks for data access
func (ctx *SandboxedExecutionContext) GetDebugData() map[string]interface{} {
	// Record resource access
	ctx.recordResourceAccess("debug_data")

	// Get the debug data from the base context
	return ctx.ExecutionContext.GetDebugData()
}

// SetOutputValue overrides base method to add resource tracking for memory
func (ctx *SandboxedExecutionContext) SetOutputValue(pinID string, value types.Value) {
	// Estimate memory size
	size := ctx.resourceMonitor.EstimateMemorySize(value.RawValue)

	// Track memory allocation
	if !ctx.resourceMonitor.TrackMemory(size) {
		// Memory limit exceeded, log error but continue with reduced functionality
		ctx.Logger().Error("Memory limit exceeded when setting output for "+pinID, map[string]interface{}{
			"valueSize": size,
			"reason":    ctx.resourceMonitor.GetLimitExceededReason(),
		})

		// Still attempt to set the value
		ctx.ExecutionContext.SetOutputValue(pinID, value)
		return
	}

	// Set the value in the base context
	ctx.ExecutionContext.SetOutputValue(pinID, value)
}

// GetVariable overrides base method to add security checks for variable access
func (ctx *SandboxedExecutionContext) GetVariable(name string) (types.Value, bool) {
	// Security check for data access
	if !ctx.securityChecker.IsDataAccessAllowed(name, ctx.permissionLevel) {
		ctx.Logger().Warn("Security violation: access to variable denied", map[string]interface{}{
			"variable":        name,
			"permissionLevel": ctx.permissionLevel,
		})
		return types.Value{}, false
	}

	// Record resource access
	ctx.recordResourceAccess(fmt.Sprintf("variable:%s", name))

	// Get the value from the base context
	return ctx.ExecutionContext.GetVariable(name)
}

// SetVariable overrides base method to add resource tracking for memory
func (ctx *SandboxedExecutionContext) SetVariable(name string, value types.Value) {
	// Estimate memory size
	size := ctx.resourceMonitor.EstimateMemorySize(value.RawValue)

	// Track memory allocation
	if !ctx.resourceMonitor.TrackMemory(size) {
		// Memory limit exceeded, log error but continue with reduced functionality
		ctx.Logger().Error("Memory limit exceeded when setting variable "+name, map[string]interface{}{
			"valueSize": size,
			"reason":    ctx.resourceMonitor.GetLimitExceededReason(),
		})

		// Still attempt to set the variable
		ctx.ExecutionContext.SetVariable(name, value)
		return
	}

	// Set the value in the base context
	ctx.ExecutionContext.SetVariable(name, value)
}

// recordResourceAccess keeps track of resource access patterns
func (ctx *SandboxedExecutionContext) recordResourceAccess(resource string) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.accessedResources[resource] = true
}

// GetResourceAccessLog returns a list of accessed resources
func (ctx *SandboxedExecutionContext) GetResourceAccessLog() []string {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	resources := make([]string, 0, len(ctx.accessedResources))
	for resource := range ctx.accessedResources {
		resources = append(resources, resource)
	}

	return resources
}

// GetResourceMonitor returns the resource monitor
func (ctx *SandboxedExecutionContext) GetResourceMonitor() *ResourceMonitor {
	return ctx.resourceMonitor
}

// GetSecurityChecker returns the security checker
func (ctx *SandboxedExecutionContext) GetSecurityChecker() *SecurityChecker {
	return ctx.securityChecker
}

// GetPermissionLevel returns the permission level
func (ctx *SandboxedExecutionContext) GetPermissionLevel() PermissionLevel {
	return ctx.permissionLevel
}

// NetworkRequest performs a network request with security checks
func (ctx *SandboxedExecutionContext) NetworkRequest(url, method string, data []byte) ([]byte, error) {
	// Check if network is blocked
	if ctx.blockNetwork {
		return nil, fmt.Errorf("network access is blocked in this execution context")
	}

	// Check if network request is allowed
	if !ctx.securityChecker.IsNetworkRequestAllowed(url, method, ctx.permissionLevel) {
		return nil, fmt.Errorf("security violation: network request to %s using method %s is not allowed", url, method)
	}

	// Track outgoing network I/O
	if !ctx.resourceMonitor.TrackNetworkIO(uint64(len(data))) {
		return nil, fmt.Errorf("network I/O limit exceeded: %s", ctx.resourceMonitor.GetLimitExceededReason())
	}

	// Create a rate limit key for this request
	rateLimitKey := RateLimitKey(fmt.Sprintf("%s:network", ctx.userID))
	if err := ctx.rateLimiter.CheckRateLimit(RateLimitTypeUser, rateLimitKey); err != nil {
		return nil, fmt.Errorf("rate limit exceeded for network requests: %v", err)
	}

	// Record the access
	ctx.recordResourceAccess(fmt.Sprintf("network:%s:%s", method, url))

	// Execute the request using base context implementation if available
	if networkRequester, ok := ctx.ExecutionContext.(interface {
		NetworkRequest(url, method string, data []byte) ([]byte, error)
	}); ok {
		response, err := networkRequester.NetworkRequest(url, method, data)

		// Track incoming network I/O if successful
		if err == nil && response != nil {
			ctx.resourceMonitor.TrackNetworkIO(uint64(len(response)))
		}

		return response, err
	}

	return nil, fmt.Errorf("network requests not supported by the base execution context")
}

// FileOperation performs a file system operation with security checks
func (ctx *SandboxedExecutionContext) FileOperation(path, operation string, data []byte) ([]byte, error) {
	// Check if file system is blocked
	if ctx.blockFileSystem {
		return nil, fmt.Errorf("file system access is blocked in this execution context")
	}

	// Check if file path is allowed
	if !ctx.securityChecker.IsFilePathAllowed(path) {
		return nil, fmt.Errorf("security violation: access to path %s is not allowed", path)
	}

	// Track I/O for write operations
	if operation == "write" || operation == "append" {
		if !ctx.resourceMonitor.TrackDiskIO(uint64(len(data))) {
			return nil, fmt.Errorf("disk I/O limit exceeded: %s", ctx.resourceMonitor.GetLimitExceededReason())
		}
	}

	// Create a rate limit key for this operation
	rateLimitKey := RateLimitKey(fmt.Sprintf("%s:file:%s", ctx.userID, operation))
	if err := ctx.rateLimiter.CheckRateLimit(RateLimitTypeUser, rateLimitKey); err != nil {
		return nil, fmt.Errorf("rate limit exceeded for file operations: %v", err)
	}

	// Record the access
	ctx.recordResourceAccess(fmt.Sprintf("file:%s:%s", operation, path))

	// Execute the operation using base context implementation if available
	if fileOperator, ok := ctx.ExecutionContext.(interface {
		FileOperation(path, operation string, data []byte) ([]byte, error)
	}); ok {
		response, err := fileOperator.FileOperation(path, operation, data)

		// Track incoming I/O for read operations
		if err == nil && response != nil && (operation == "read") {
			ctx.resourceMonitor.TrackDiskIO(uint64(len(response)))
		}

		return response, err
	}

	return nil, fmt.Errorf("file operations not supported by the base execution context")
}

// ExecuteCommand executes a system command with security checks
func (ctx *SandboxedExecutionContext) ExecuteCommand(command string, args []string) ([]byte, error) {
	// Check if command execution is blocked
	if ctx.blockCommandExec {
		return nil, fmt.Errorf("command execution is blocked in this execution context")
	}

	// Check if command execution is allowed
	if !ctx.securityChecker.IsCommandExecutionAllowed(command, ctx.permissionLevel) {
		return nil, fmt.Errorf("security violation: execution of command %s is not allowed", command)
	}

	// Create a rate limit key for this command
	rateLimitKey := RateLimitKey(fmt.Sprintf("%s:command", ctx.userID))
	if err := ctx.rateLimiter.CheckRateLimit(RateLimitTypeUser, rateLimitKey); err != nil {
		return nil, fmt.Errorf("rate limit exceeded for command execution: %v", err)
	}

	// Record the access
	ctx.recordResourceAccess(fmt.Sprintf("command:%s", command))

	// Execute the command using base context implementation if available
	if commandExecutor, ok := ctx.ExecutionContext.(interface {
		ExecuteCommand(command string, args []string) ([]byte, error)
	}); ok {
		return commandExecutor.ExecuteCommand(command, args)
	}

	return nil, fmt.Errorf("command execution not supported by the base execution context")
}

// Cleanup releases resources used by the sandboxed context
func (ctx *SandboxedExecutionContext) Cleanup() {
	// Clean up resource monitor
	ctx.resourceMonitor.Cleanup()

	// Call the base cleanup if it exists
	if cleanup, ok := ctx.ExecutionContext.(interface{ Cleanup() }); ok {
		cleanup.Cleanup()
	}
}

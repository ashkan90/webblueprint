# WebBlueprint Security Framework

This security framework provides a robust implementation for Blueprint İzolasyonu ve Güvenlik (Blueprint Isolation and Security) as defined in the WEB-16 task. It aims to ensure that blueprints run within controlled environments with appropriate resource limits, security restrictions, and isolation mechanisms.

## Components

### 1. Resource Limiting System

The resource limiting system (`resource_limits.go`) tracks and enforces limits on resource usage during blueprint execution:

- CPU usage monitoring
- Memory allocation tracking
- Execution time limits
- Disk I/O usage restrictions
- Network I/O usage restrictions
- Node execution count limits

Resource profiles are provided for different levels of access:
- Low: For simple blueprints
- Medium: For standard blueprints (default)
- High: For complex blueprints
- Unlimited: For trusted blueprints

### 2. Security Checker

The security checker (`security_checker.go`) validates potentially dangerous operations:

- Validates node types against permission levels
- Checks network domains for security concerns
- Verifies file paths for allowed access
- Validates commands for potential security issues
- Checks data access for sensitive information
- Content validation for dangerous patterns

Permission levels include:
- Restricted: Most restricted mode
- Standard: Default security level
- Trusted: Elevated privileges for trusted blueprints
- Admin: Full administrative access

### 3. Rate Limiter

The rate limiter (`rate_limiter.go`) enforces limits on execution frequency to prevent abuse:

- Per-user rate limiting
- Per-blueprint execution frequency limits
- API call throttling
- Gradual backoff for repeated executions
- Different limit types (user, blueprint, API)

### 4. Sandboxed Execution Context

The sandboxed execution context (`sandbox_context.go`) provides an isolated environment:

- Context isolation for execution
- Resource tracking during execution
- Security checks for operations
- Rate limiting integration
- Configurable restrictions (network, filesystem, commands)

### 5. Security Manager

The security manager (`security_manager.go`) coordinates security features:

- Central security configuration
- Permission management
- Resource profile management
- Comprehensive security checks
- Rate limit management

## Integration

The `integration.go` file shows how to integrate the security framework with the existing execution engine. It provides:

- Adapter for the execution engine
- Security wrapper for blueprint execution
- Helpers for creating secured execution contexts

## Usage

```go
// Create security manager
securityManager := security.NewSecurityManager()

// Set permissions
securityManager.SetUserPermission("user123", security.PermissionLevelStandard)
securityManager.SetBlueprintPermission("blueprint456", security.PermissionLevelTrusted)

// Create secure engine adapter
adapter := security.NewSecureEngineAdapter()

// Wrap execution context with security
secureContext := adapter.WrapExecutionContext(
    baseCtx, 
    "user123", 
    "blueprint456", 
    "execution789", 
)

// Execute blueprint with security
secureExecuteFn := security.SecureExecutionWrapper(adapter, executionEngine)
result, err := secureExecuteFn(blueprint, executionID, userID, initialData)
```

## Testing

A comprehensive test suite is included in `security_test.go` to verify the behavior of all security components.

Run the tests with:
```
go test ./internal/security
```

## Implementation Details

This implementation fulfills the requirements specified in the WEB-16 task by providing:

1. Resource limiting mechanisms that prevent excessive consumption of CPU, memory, and execution time
2. Security checks that block potentially dangerous operations
3. Rate limiting to prevent system abuse
4. Sandbox execution providing proper isolation between blueprints
5. Permission system enforcing access control for operations

The implementation is designed to be minimally invasive to the existing codebase while providing robust security features.

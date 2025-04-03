package security

import (
	"fmt"
	"testing"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// Test resource limits
func TestResourceLimits(t *testing.T) {
	// Create resource monitor with low limits
	limits := ResourceLimits{
		MaxCPUTime:        1 * time.Second,
		MaxMemoryBytes:    1024 * 1024, // 1MB
		MaxExecutionTime:  2 * time.Second,
		MaxDiskIOBytes:    1024,
		MaxNetworkIOBytes: 1024,
		MaxNodeExecutions: 10,
	}
	monitor := NewResourceMonitor(limits)

	// Test node execution tracking
	for i := 0; i < 10; i++ {
		if !monitor.TrackNodeExecution() {
			t.Errorf("Node execution tracking failed on iteration %d", i)
		}
	}

	// This should exceed the limit
	if monitor.TrackNodeExecution() {
		t.Errorf("Node execution tracking should have failed after exceeding limit")
	}

	if !monitor.LimitExceeded() {
		t.Errorf("LimitExceeded() should return true after exceeding node execution limit")
	}

	// Clean up
	monitor.Cleanup()
}

// Test security checker
func TestSecurityChecker(t *testing.T) {
	// Create security checker with standard permission
	checker := NewSecurityChecker(PermissionLevelStandard)

	// Test node permissions
	if !checker.IsNodeAllowed("constant-number", PermissionLevelStandard) {
		t.Errorf("Basic node type should be allowed at standard permission level")
	}

	if checker.IsNodeAllowed("execute-command", PermissionLevelStandard) {
		t.Errorf("Command execution should not be allowed at standard permission level")
	}

	// Test network request permissions
	if !checker.IsNetworkRequestAllowed("https://example.com", "GET", PermissionLevelStandard) {
		t.Errorf("Standard HTTP GET should be allowed at standard permission level")
	}

	if checker.IsNetworkRequestAllowed("https://example.com", "DELETE", PermissionLevelStandard) {
		t.Errorf("DELETE method should not be allowed at standard permission level")
	}

	// Test local network restrictions
	if checker.IsDomainAllowed("localhost") {
		t.Errorf("localhost should be blocked")
	}

	if checker.IsDomainAllowed("192.168.1.1") {
		t.Errorf("Local IP should be blocked")
	}

	// Test file path permissions
	if checker.IsFilePathAllowed("/etc/passwd") {
		t.Errorf("System files should be blocked")
	}

	// Test content validation
	valid, _ := checker.ValidateContent("Hello world")
	if !valid {
		t.Errorf("Simple text should be valid")
	}

	valid, reason := checker.ValidateContent("<script>alert('xss')</script>")
	if valid {
		t.Errorf("Script tag should be flagged as invalid")
	} else {
		fmt.Printf("Content validation reason: %s\n", reason)
	}
}

// Test rate limiter
func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter()

	// Configure a test limit
	limiter.SetRateLimit(RateLimitTypeUser, 5, 10*time.Second)

	// Test successful operations
	key := RateLimitKey("test-user")
	for i := 0; i < 5; i++ {
		if !limiter.Allow(RateLimitTypeUser, key) {
			t.Errorf("Rate limiter should allow operation %d", i)
		}
	}

	// This should exceed the limit
	if limiter.Allow(RateLimitTypeUser, key) {
		t.Errorf("Rate limiter should deny operation after exceeding limit")
	}

	// Test error handling
	err := limiter.CheckRateLimit(RateLimitTypeUser, key)
	if err == nil {
		t.Errorf("CheckRateLimit should return error when limit exceeded")
	}

	// Verify it's a rate limit error
	if !IsRateLimitError(err) {
		t.Errorf("IsRateLimitError should return true for rate limit errors")
	}

	// Reset the limit
	limiter.Reset(RateLimitTypeUser, key)

	// Should work again
	if !limiter.Allow(RateLimitTypeUser, key) {
		t.Errorf("Rate limiter should allow operation after reset")
	}
}

// Test security manager
func TestSecurityManager(t *testing.T) {
	manager := NewSecurityManager()

	// Test permission management
	manager.SetUserPermission("user1", PermissionLevelRestricted)
	manager.SetBlueprintPermission("blueprint1", PermissionLevelTrusted)

	if manager.GetUserPermission("user1") != PermissionLevelRestricted {
		t.Errorf("User permission not set correctly")
	}

	if manager.GetBlueprintPermission("blueprint1") != PermissionLevelTrusted {
		t.Errorf("Blueprint permission not set correctly")
	}

	// Test sandbox options
	options := manager.GetSandboxOptions("user1", "blueprint1")

	// Blueprint permission should override user permission
	if options.PermissionLevel != PermissionLevelTrusted {
		t.Errorf("Blueprint permission should take precedence over user permission")
	}

	// Test comprehensive security check
	err := manager.CheckSecurity("user1", "blueprint1", "http-request", "network", "https://example.com")
	if err != nil {
		t.Errorf("Security check failed: %v", err)
	}

	err = manager.CheckSecurity("user1", "blueprint1", "execute-command", "command", "rm -rf /")
	if err == nil {
		t.Errorf("Dangerous command should be blocked")
	} else {
		fmt.Printf("Command security check error: %v\n", err)
	}
}

// Test sandboxed execution context
func TestSandboxedExecutionContext(t *testing.T) {
	// Create a mock base context
	baseCtx := &TestMockExecutionContext{
		nodeID:      "node1",
		nodeType:    "test-node",
		blueprintID: "bp1",
		executionID: "exec1",
		inputs:      make(map[string]types.Value),
		outputs:     make(map[string]types.Value),
		variables:   make(map[string]types.Value),
		logger:      &mockLogger{},
	}

	// Add some test values
	baseCtx.inputs["input1"] = types.NewValue(types.PinTypes.String, "test input")
	baseCtx.variables["var1"] = types.NewValue(types.PinTypes.Number, 42)

	// Create sandbox options
	options := SandboxOptions{
		BlockNetwork:     false,
		BlockFileSystem:  true,
		BlockCommandExec: true,
		ResourceProfile:  ResourceProfileLow,
		PermissionLevel:  PermissionLevelStandard,
		RateLimiter:      NewRateLimiter(),
	}

	// Create sandboxed context
	ctx := NewSandboxedExecutionContext(
		baseCtx,
		"user1",
		"bp1",
		"exec1",
		options,
	)

	// Test input access
	if val, exists := ctx.GetInputValue("input1"); !exists || val.RawValue != "test input" {
		t.Errorf("Input value retrieval failed")
	}

	// Test variable access
	if val, exists := ctx.GetVariable("var1"); !exists || val.RawValue.(int) != 42 {
		t.Errorf("Variable retrieval failed")
	}

	// Test setting outputs
	ctx.SetOutputValue("output1", types.NewValue(types.PinTypes.String, "test output"))

	// Check outputs through base context
	if val, exists := baseCtx.outputs["output1"]; !exists || val.RawValue != "test output" {
		t.Errorf("Output value setting failed")
	}

	// Test file operation (should be blocked)
	_, err := ctx.FileOperation("/tmp/test.txt", "read", nil)
	if err == nil {
		t.Errorf("File operation should be blocked")
	} else {
		fmt.Printf("File operation error: %v\n", err)
	}

	// Clean up
	ctx.Cleanup()
}

// Mock logger for testing
type mockLogger struct{}

func (l *mockLogger) Debug(msg string, fields map[string]interface{}) {}
func (l *mockLogger) Info(msg string, fields map[string]interface{})  {}
func (l *mockLogger) Warn(msg string, fields map[string]interface{})  {}
func (l *mockLogger) Error(msg string, fields map[string]interface{}) {}
func (l *mockLogger) Opts(opts map[string]interface{})                {}

// Mock execution context for testing
type TestMockExecutionContext struct {
	nodeID      string
	nodeType    string
	blueprintID string
	executionID string
	inputs      map[string]types.Value
	outputs     map[string]types.Value
	variables   map[string]types.Value
	logger      node.Logger
	hooks       *node.ExecutionHooks
}

func (m *TestMockExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := m.inputs[pinID]
	return value, exists
}

func (m *TestMockExecutionContext) SetOutputValue(pinID string, value types.Value) {
	m.outputs[pinID] = value
}

func (m *TestMockExecutionContext) GetActivatedOutputFlows() []string {
	return []string{}
}

func (m *TestMockExecutionContext) ActivateOutputFlow(pinID string) error {
	return nil
}

func (m *TestMockExecutionContext) ExecuteConnectedNodes(pinID string) error {
	return nil
}

func (m *TestMockExecutionContext) GetVariable(name string) (types.Value, bool) {
	value, exists := m.variables[name]
	return value, exists
}

func (m *TestMockExecutionContext) SetVariable(name string, value types.Value) {
	m.variables[name] = value
}

func (m *TestMockExecutionContext) Logger() node.Logger {
	return m.logger
}

func (m *TestMockExecutionContext) GetNodeID() string {
	return m.nodeID
}

func (m *TestMockExecutionContext) GetNodeType() string {
	return m.nodeType
}

func (m *TestMockExecutionContext) GetBlueprintID() string {
	return m.blueprintID
}

func (m *TestMockExecutionContext) GetExecutionID() string {
	return m.executionID
}

func (m *TestMockExecutionContext) GetDebugData() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *TestMockExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// Do nothing
}

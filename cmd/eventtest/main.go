package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	"webblueprint/internal/core" // Import core
	"webblueprint/internal/event"
	"webblueprint/internal/test/mocks"
	"webblueprint/internal/types"
)

// Simple logger for testing
type testLogger struct{}

func (l *testLogger) Debug(msg string, fields map[string]interface{}) {
	log.Printf("[DEBUG] %s %v\n", msg, fields)
}

func (l *testLogger) Info(msg string, fields map[string]interface{}) {
	log.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *testLogger) Warn(msg string, fields map[string]interface{}) {
	log.Printf("[WARN] %s %v\n", msg, fields)
}

func (l *testLogger) Error(msg string, fields map[string]interface{}) {
	log.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *testLogger) Opts(options map[string]interface{}) {
	// No-op for test logger
}

// Basic execution context for testing
func newTestExecutionContext() *mocks.MockExecutionContext {
	logger := &testLogger{}
	mock := mocks.NewMockExecutionContext("test-node-1", "test-node", logger)
	return mock
}

// Test event handler - No longer needed as handler logic is internal to EventManager
/*
type testEventHandler struct {
	id      string
	handler func(ctx event.EventHandlerContext) error
}

func (h *testEventHandler) HandleEvent(evt event.EventDispatchRequest) error {
	ctx := event.EventHandlerContext{
		EventID:     evt.EventID,
		Parameters:  evt.Parameters,
		SourceID:    evt.SourceID,
		BlueprintID: evt.BlueprintID,
		ExecutionID: evt.ExecutionID,
		HandlerID:   h.id,
		Timestamp:   evt.Timestamp,
	}
	return h.handler(ctx)
}

func (h *testEventHandler) GetHandlerID() string {
	return h.id
}
*/

// mockEngineController is a simple mock for testing
type mockEngineController struct {
	triggeredNodes map[string]core.EventHandlerContext
	mutex          sync.Mutex
}

func newMockEngineController() *mockEngineController {
	return &mockEngineController{
		triggeredNodes: make(map[string]core.EventHandlerContext),
	}
}

func (m *mockEngineController) TriggerNodeExecution(blueprintID string, nodeID string, triggerContext core.EventHandlerContext) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	key := fmt.Sprintf("%s:%s", blueprintID, nodeID)
	m.triggeredNodes[key] = triggerContext
	fmt.Printf("MockEngineController: Triggered Node %s (Blueprint: %s) for Event: %s\n", nodeID, blueprintID, triggerContext.EventID)
	// Simulate successful trigger
	// In a real test, you might check parameters here or simulate errors
	fmt.Printf("  Parameters:\n")
	for name, value := range triggerContext.Parameters {
		fmt.Printf("    %s: %v\n", name, value.RawValue)
	}
	return nil
}

// Helper to check if a node was triggered (for test assertions, though not used in this main func)
func (m *mockEngineController) WasNodeTriggered(blueprintID string, nodeID string) (core.EventHandlerContext, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	key := fmt.Sprintf("%s:%s", blueprintID, nodeID)
	ctx, exists := m.triggeredNodes[key]
	return ctx, exists
}

func main() {
	fmt.Println("===== WebBlueprint Event System Test =====")

	// Create mock engine and event manager
	mockEngine := newMockEngineController()
	eventManager := event.NewEventManager(mockEngine)

	// Define event and handler details
	eventID := "test.event.button.clicked"
	handlerNodeID := "test-handler-node-1" // ID of the node that *would* handle this
	testBlueprintID := "test-blueprint-main"

	// Create an EventBinding to link the event to the handler node
	binding := event.EventBinding{
		ID:          fmt.Sprintf("binding-%s-%s", eventID, handlerNodeID), // Unique ID for the binding
		EventID:     eventID,
		HandlerID:   handlerNodeID,       // ID of the node that should be triggered
		HandlerType: "test-handler-node", // Type of the node being triggered
		BlueprintID: testBlueprintID,     // Blueprint containing the handler node
		Priority:    0,
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	// Register the binding (this also registers the internal handler func in EventManager)
	err := eventManager.BindEvent(binding)
	if err != nil {
		fmt.Printf("Error binding event: %v\n", err)
		return
	}
	fmt.Println("\n> Registered binding for event:", eventID, "to handler:", handlerNodeID)

	// Create a context provider
	// Pass both the event manager and the engine controller
	contextProvider := event.NewContextProvider(eventManager, mockEngine)

	// Create a basic execution context
	baseCtx := newTestExecutionContext()
	// baseCtx.SetBlueprintID(testBlueprintID) // Set blueprint ID on mock context - Method doesn't exist on mock

	// Create an event-aware context
	eventCtx := contextProvider.CreateEventAwareContext(baseCtx, false, nil)

	// Dispatch an event through the context
	fmt.Println("\n> Dispatching event via context...")

	params := map[string]types.Value{
		"buttonId": types.NewValue(types.PinTypes.String, "submit-button"),
		"x":        types.NewValue(types.PinTypes.Number, 120.5),
		"y":        types.NewValue(types.PinTypes.Number, 250.0),
	}

	// Try to cast to the ExecutionContextWithEvents interface
	if eventAwareCtx, ok := eventCtx.(event.ExecutionContextWithEvents); ok {
		err = eventAwareCtx.DispatchEvent(eventID, params)
		if err != nil {
			fmt.Printf("Error dispatching event via context: %v\n", err)
		} else {
			fmt.Println("> Event dispatched successfully via context")
		}
	} else {
		fmt.Println("ERROR: Failed to cast to EventAwareContext")
	}

	// Try direct dispatch using the EventManager
	fmt.Println("\n> Testing direct event dispatch via EventManager...")
	request := event.EventDispatchRequest{
		EventID:     eventID,
		Parameters:  params,
		SourceID:    "test-direct-dispatch",
		BlueprintID: testBlueprintID,
		ExecutionID: "test-execution-direct",
		Timestamp:   time.Now(),
	}

	errors := eventManager.DispatchEvent(request)
	if len(errors) > 0 {
		fmt.Printf("Errors during direct dispatch: %v\n", errors)
	} else {
		fmt.Println("> Direct dispatch successful via EventManager")
	}

	// Removed the adapter test section as it relied on removed/deprecated methods

	fmt.Println("\n===== Event System Test Complete =====")
}

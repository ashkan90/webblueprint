package main

import (
	"log"
	"net/http"
	"webblueprint/internal/api"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/nodes/web"
	"webblueprint/internal/registry"
)

// setupErrorHandling initializes the error handling system
func setupErrorHandling(baseEngine *engine.ExecutionEngine, wsManager *api.WebSocketManager) *engine.ErrorAwareEngine {
	// Create error-aware engine
	errorAwareEngine := engine.NewErrorAwareEngine(baseEngine)

	// Get error manager from the engine
	errorManager := errorAwareEngine.GetErrorManager()

	// Create WebSocket handler for error notifications
	errorWsHandler := api.NewErrorWebSocketHandler(wsManager)
	errorWsHandler.RegisterWithErrorManager(errorManager)

	// Register error handlers for logging
	errorManager.RegisterErrorHandler(bperrors.ErrorTypeExecution, func(err *bperrors.BlueprintError) error {
		log.Printf("[ERROR] Execution error: %s (Node: %s, Code: %s)",
			err.Message, err.NodeID, err.Code)
		return nil
	})

	return errorAwareEngine
}

// setupErrorAPI sets up API endpoints for error handling
func setupErrorAPI(router *http.ServeMux, errorAwareEngine *engine.ErrorAwareEngine, wsManager *api.WebSocketManager) {
	// Create API handler
	errorAPI := api.NewErrorHandlingAPI(
		errorAwareEngine.GetErrorManager(),
		errorAwareEngine.GetRecoveryManager(),
		errorAwareEngine.InfoStore,
	)

	// Register API routes
	router.HandleFunc("/api/errors/list", errorAPI.HandleListErrors)
	router.HandleFunc("/api/errors/analysis", errorAPI.HandleGetErrorAnalysis)
	router.HandleFunc("/api/errors/info", errorAPI.HandleGetExecutionInfo)
	router.HandleFunc("/api/errors/recover", errorAPI.HandleAttemptRecovery)
	router.HandleFunc("/api/errors/clear", errorAPI.HandleClearErrors)

	// Setup test API for development
	testErrorAPI := &api.TestErrorHandler{
		ErrorManager:    errorAwareEngine.GetErrorManager(),
		RecoveryManager: errorAwareEngine.GetRecoveryManager(),
		WSManager:       wsManager,
	}

	router.HandleFunc("/api/test/generate-error", testErrorAPI.HandleGenerateTestError)
	router.HandleFunc("/api/test/generate-scenario", testErrorAPI.HandleGenerateErrorScenario)
}

// registerErrorAwareNodes registers nodes with error handling capabilities
func registerErrorAwareNodes(registry *registry.GlobalNodeRegistry) {
	registry.RegisterNodeType("http-request-with-recovery", web.NewHTTPRequestWithRecoveryNode)
}

// Example of how to integrate with the main application
/*
func main() {
	// Create router
	router := http.NewServeMux()

	// Setup base components
	baseEngine := engine.NewExecutionEngine()
	wsManager := api.NewWebSocketManager()

	// Setup error handling
	errorAwareEngine := setupErrorHandling(baseEngine, wsManager)

	// Setup error API
	setupErrorAPI(router, errorAwareEngine, wsManager)

	// Register error-aware nodes
	registerErrorAwareNodes(errorAwareEngine.BaseEngine)

	// Start server
	log.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", router)
}
*/

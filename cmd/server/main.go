package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"webblueprint/internal/api"
	"webblueprint/internal/engine"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/nodes/web"
)

func main() {
	// Initialize WebSocket manager
	wsManager := api.NewWebSocketManager()

	wsLogger := api.NewWebSocketLogger(wsManager)

	// Initialize debug manager
	debugManager := engine.NewDebugManager()

	// Initialize execution engine
	executionEngine := engine.NewExecutionEngine(wsLogger, debugManager)

	// Initialize API server
	apiServer := api.NewAPIServer(executionEngine, wsManager, debugManager)

	// Register execution event listener
	executionEventListener := api.NewExecutionEventListener(wsManager)
	executionEngine.AddExecutionListener(executionEventListener)

	// Register node types
	// Logic nodes
	apiServer.RegisterNodeType("if-condition", logic.NewIfConditionNode)

	// Web nodes
	apiServer.RegisterNodeType("http-request", web.NewHTTPRequestNode)

	// Utility nodes
	apiServer.RegisterNodeType("print", utility.NewPrintNode)

	// Set up routes
	router := apiServer.SetupRoutes()

	// Get server port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089" // Default port
	}

	// Determine frontend path
	frontendPath := "./web/dist"
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		// Try to find the frontend in other locations
		candidates := []string{
			"./web/dist",
			"../web/dist",
			"../../web/dist",
		}

		for _, path := range candidates {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				frontendPath = path
				break
			}
		}
	}

	// Get absolute path for frontend
	absPath, err := filepath.Abs(frontendPath)
	if err == nil {
		log.Printf("Serving frontend from: %s", absPath)
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("WebBlueprint server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"webblueprint/internal/api"
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/nodes/math"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/nodes/web"
	"webblueprint/internal/registry"
)

func main() {
	// Define command line flags
	useActorSystem := flag.Bool("actor", true, "Use the actor system for execution (default: false)")
	port := flag.String("port", "", "Server port (default: 8089 or $PORT env var)")
	flag.Parse()

	// Initialize WebSocket manager
	wsManager := api.NewWebSocketManager()

	wsLogger := api.NewWebSocketLogger(wsManager)

	// Initialize debug manager
	debugManager := engine.NewDebugManager()

	// Initialize execution engine
	executionEngine := engine.NewExecutionEngine(wsLogger, debugManager)

	// Set execution mode if actor system is enabled
	if *useActorSystem {
		log.Println("Using Actor System for execution")
		executionEngine.SetExecutionMode(engine.ModeActor)
	} else {
		log.Println("Using Standard Engine for execution")
	}

	// Initialize API server
	apiServer := api.NewAPIServer(executionEngine, wsManager, debugManager)

	// Register execution event listener
	executionEventListener := api.NewExecutionEventListener(wsManager)
	executionEngine.AddExecutionListener(executionEventListener)

	// Get global node registry
	globalRegistry := registry.GetInstance()

	// Register node types
	// Logic nodes
	registerNodeType(apiServer, globalRegistry, "if-condition", logic.NewIfConditionNode)
	registerNodeType(apiServer, globalRegistry, "loop", logic.NewLoopNode)
	registerNodeType(apiServer, globalRegistry, "sequence", logic.NewSequenceNode)
	registerNodeType(apiServer, globalRegistry, "branch", logic.NewBranchNode)

	// Web nodes
	registerNodeType(apiServer, globalRegistry, "http-request", web.NewHTTPRequestNode)
	registerNodeType(apiServer, globalRegistry, "dom-element", web.NewDOMElementNode)
	registerNodeType(apiServer, globalRegistry, "dom-event", web.NewDOMEventNode)
	registerNodeType(apiServer, globalRegistry, "storage", web.NewStorageNode)

	// Data nodes
	registerNodeType(apiServer, globalRegistry, "constant-string", data.NewStringConstantNode)
	registerNodeType(apiServer, globalRegistry, "constant-number", data.NewNumberConstantNode)
	registerNodeType(apiServer, globalRegistry, "constant-boolean", data.NewBooleanConstantNode)
	registerNodeType(apiServer, globalRegistry, "variable-get", data.NewVariableGetNode)
	registerNodeType(apiServer, globalRegistry, "variable-set", data.NewVariableSetNode)
	registerNodeType(apiServer, globalRegistry, "json-processor", data.NewJSONNode)
	registerNodeType(apiServer, globalRegistry, "array-operations", data.NewArrayNode)
	registerNodeType(apiServer, globalRegistry, "object-operations", data.NewObjectNode)
	registerNodeType(apiServer, globalRegistry, "type-conversion", data.NewTypeConversionNode)

	// Math nodes
	registerNodeType(apiServer, globalRegistry, "math-add", math.NewAddNode)
	registerNodeType(apiServer, globalRegistry, "math-subtract", math.NewSubtractNode)
	registerNodeType(apiServer, globalRegistry, "math-multiply", math.NewMultiplyNode)
	registerNodeType(apiServer, globalRegistry, "math-divide", math.NewDivideNode)

	// Utility nodes
	registerNodeType(apiServer, globalRegistry, "print", utility.NewPrintNode)
	registerNodeType(apiServer, globalRegistry, "timer", utility.NewTimerNode)

	// Set up routes
	router := apiServer.SetupRoutes()

	// Get server port from flag, environment, or use default
	serverPort := *port
	if serverPort == "" {
		serverPort = os.Getenv("PORT")
		if serverPort == "" {
			serverPort = "8089" // Default port
		}
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
	addr := fmt.Sprintf(":%s", serverPort)
	log.Printf("WebBlueprint server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

// Helper function to register a node type with both the API server and global registry
func registerNodeType(apiServer *api.APIServer, globalRegistry *registry.GlobalNodeRegistry, typeID string, factory func() node.Node) {
	apiServer.RegisterNodeType(typeID, factory)
}

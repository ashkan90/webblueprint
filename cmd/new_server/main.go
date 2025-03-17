package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webblueprint/internal/api"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/nodes"
	"webblueprint/internal/nodes/math"
	"webblueprint/internal/nodes/web"
	"webblueprint/internal/registry"
	"webblueprint/pkg/db"
	"webblueprint/pkg/repository"

	"github.com/gorilla/mux"
)

func setupErrorHandling(baseEngine *engine.ExecutionEngine, wsManager *api.WebSocketManager) *engine.ErrorAwareEngine {
	// Create error-aware engine
	errorAwareEngine := engine.NewErrorAwareEngine(baseEngine)

	// Get error manager from the engine
	errorManager := errorAwareEngine.GetErrorManager()
	recoveryManager := errorAwareEngine.GetRecoveryManager()

	// Create WebSocket logger
	wsLogger := &wsLogger{}

	// Register error handlers with WebSocket manager
	wsManager.RegisterErrorHandlers(errorManager, wsLogger)

	// Create error notification handler
	wsLoggerClient := api.NewWebSocketLogger(wsManager)
	notificationHandler := api.NewErrorNotificationHandler(wsManager, wsLoggerClient, errorManager, recoveryManager)

	// Register notification handler with error manager
	for _, errType := range []bperrors.ErrorType{
		bperrors.ErrorTypeExecution,
		bperrors.ErrorTypeConnection,
		bperrors.ErrorTypeValidation,
		bperrors.ErrorTypeSystem,
	} {
		// Convert method to correct signature
		handler := func(err *bperrors.BlueprintError) error {
			notificationHandler.HandleExecutionError(err.ExecutionID, err)
			return nil
		}
		errorManager.RegisterErrorHandler(errType, handler)
	}

	return errorAwareEngine
}

// Setup routes for error management API
func setupErrorAPI(router *mux.Router, errorAwareEngine *engine.ErrorAwareEngine, wsManager *api.WebSocketManager) {
	errorManager := errorAwareEngine.GetErrorManager()
	recoveryManager := errorAwareEngine.GetRecoveryManager()

	// Create recovery handler
	recoveryHandler := api.NewErrorRecoveryHandler(wsManager, errorManager, recoveryManager)
	recoveryHandler.RegisterRoutes(router)

	// Create test error handler
	testHandler := api.NewTestErrorHandler(errorManager, recoveryManager, wsManager)
	testHandler.RegisterRoutes(router)
}

// Custom logger that implements the Logger interface
type wsLogger struct{}

func (l *wsLogger) Debug(msg string, fields map[string]interface{}) {
	slog.Debug(msg, l.toFields(fields))
}

func (l *wsLogger) Info(msg string, fields map[string]interface{}) {
	slog.Info(msg, l.toFields(fields)...)
}

func (l *wsLogger) Warn(msg string, fields map[string]interface{}) {
	slog.Warn(msg, l.toFields(fields))
}

func (l *wsLogger) Error(msg string, fields map[string]interface{}) {
	slog.Error(msg, l.toFields(fields))
}

func (l *wsLogger) Opts(options map[string]interface{}) {
	// Set logger options
}

func (l *wsLogger) toFields(fields map[string]interface{}) []any {
	_fields := make([]any, 0, len(fields))
	for k, field := range fields {
		if k == "" {
			continue
		}
		_fields = append(_fields, slog.Any(k, field))
	}

	return _fields
}

func main() {
	// Define command line flags
	useActorSystem := flag.Bool("actor", true, "Use actor system for execution (default: true)")
	useDatabase := flag.Bool("db", true, "Use database storage (default: true)")
	port := flag.String("port", "8089", "Server port (default: 8089 or $PORT environment variable)")
	flag.Parse()

	// Set up based on flags
	slog.Info("Using Actor System for execution")
	log.Println()

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to catch SIGINT and SIGTERM signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start WebSocket manager
	wsManager := api.NewWebSocketManager()

	// Create debug manager
	debugManager := engine.NewDebugManager()

	// Start execution engine
	var apiServer *api.APIServerWithDB
	var apiRouter *mux.Router

	var logger = &wsLogger{}

	// Register the global node registry
	registry.Make(nodes.Core)
	globalRegistry := registry.GetInstance()

	var executionEngine = engine.NewExecutionEngine(logger, debugManager)

	if *useActorSystem {
		slog.Info("Using actor system for execution")
		executionEngine.SetExecutionMode(engine.ModeActor)
		// TODO: Set actor system mode if needed
	} else {
		slog.Info("Using direct execution")
		executionEngine.SetExecutionMode(engine.ModeStandard)
		// TODO: Set direct execution mode if needed
	}

	// Setup error handling
	errorAwareEngine := setupErrorHandling(executionEngine, wsManager)

	// Set up database if requested
	if *useDatabase {
		slog.Info("Using database storage")

		// Skip database setup for now to fix build
		// TEMP STUB: direct database implementation
		log.Println("Skipping database connection for build fixing")
		connectionManager, repoFactory, err := db.Setup(context.Background())
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer connectionManager.Close()

		// Create API server with database
		apiServer = api.NewAPIServerWithDB(executionEngine, wsManager, debugManager, repoFactory, errorAwareEngine)

		// Register execution event listener
		executionEventListener := api.NewExecutionEventListener(wsManager)
		executionEngine.AddExecutionListener(executionEventListener)

		// Register node types
		registerCoreNodesWithServer(apiServer, globalRegistry)
		registerCoreSafeNodesWithServer(apiServer, globalRegistry)

		// Load blueprints from database
		if err := loadBlueprintsFromDB(ctx, repoFactory, executionEngine); err != nil {
			log.Printf("Warning: Failed to load blueprints from database: %v", err)
		}

		apiRouter = apiServer.SetupRoutes()

		// Set up error handling API endpoints
		setupErrorAPI(apiRouter, errorAwareEngine, wsManager)
	} else {
		log.Println("Using in-memory storage (no database)")
		// Create server without database
		// You'd need to implement this part based on your requirements
	}

	// Get server port from flag, environment, or default
	serverPort := *port
	if serverPort == "" {
		serverPort = os.Getenv("PORT")
		if serverPort == "" {
			serverPort = "8089" // Default port
		}
	}

	// Create the HTTP server
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: apiRouter,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-signals
	log.Println("Shutting down server...")

	// Create a deadline context for server shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}

// Helper function to register a node type
func registerNodeType(server *api.APIServerWithDB, registry *registry.GlobalNodeRegistry, typeName string, factory func() node.Node) {
	registry.RegisterNodeType(typeName, factory)
	server.RegisterNodeType(typeName, factory)
}

// Helper function to register all nodes with the server
func registerCoreSafeNodesWithServer(server *api.APIServerWithDB, registry *registry.GlobalNodeRegistry) {
	// Register any types needed for the error handling demo
	registerNodeType(server, registry, "http-request-with-recovery", web.NewHTTPRequestWithRecoveryNode)
	registerNodeType(server, registry, "safe-divide", math.NewSafeDivideNode)
}

func registerCoreNodesWithServer(server *api.APIServerWithDB, registry *registry.GlobalNodeRegistry) {
	for s, factory := range registry.GetAllNodeFactories() {
		registerNodeType(server, registry, s, factory)
	}
}

// Load blueprints from database
func loadBlueprintsFromDB(ctx context.Context, repoFactory repository.RepositoryFactory, executionEngine *engine.ExecutionEngine) error {
	// Get repositories
	blueprintRepo := repoFactory.GetBlueprintRepository()

	// Get all blueprint IDs
	query := `SELECT id FROM blueprints`

	rows, err := db.GetConnectionManager().GetDB().QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var blueprintIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		blueprintIDs = append(blueprintIDs, id)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Load each blueprint
	for _, id := range blueprintIDs {
		// Get model from database
		blueprintModel, err := blueprintRepo.GetByID(ctx, id)
		if err != nil {
			log.Printf("Could not get blueprint %s: %v", id, err)
			continue
		}

		// Convert to package blueprint
		bp, err := blueprintRepo.ToPkgBlueprint(
			blueprintModel,
			blueprintModel.CurrentVersion,
		)
		if err != nil {
			log.Printf("Could not convert blueprint %s: %v", id, err)
			continue
		}

		// Load to execution engine
		if err := executionEngine.LoadBlueprint(bp); err != nil {
			log.Printf("Could not load blueprint %s to execution engine: %v", id, err)
			continue
		}

		log.Printf("Blueprint loaded: %s (%s)", bp.Name, id)
	}

	return nil
}

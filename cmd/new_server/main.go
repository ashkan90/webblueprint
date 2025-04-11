package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"webblueprint/internal/api"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/engineext"
	"webblueprint/internal/event"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/registry"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/db"
	"webblueprint/pkg/repository"

	"github.com/gorilla/mux"
)

func main() {
	// Define command line flags
	port := flag.String("port", "8089", "Server port (default: 8089 or $PORT environment variable)")
	headlessEnabled := flag.Bool("headless", false, "Enable headless mode")
	path := flag.String("path", "./", "Path to the json file for blueprint")
	bpId := flag.String("blueprintId", "", "Blueprint Id (required)")
	flag.Parse()

	if headlessEnabled != nil && *headlessEnabled {
		headless(bpId, path)
		return
	}

	// Set log level
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))

	// Channel to catch SIGINT and SIGTERM signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Determine port
	serverPort := *port
	if envPort := os.Getenv("PORT"); envPort != "" {
		serverPort = envPort
	}

	ctx := context.Background()
	// Create router
	router := mux.NewRouter()

	registry.Make()

	//registerNodes()

	setupAPI(ctx, router)

	// Set up basic API endpoints
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{\"status\":\"ok\",\"timestamp\":\"%s\"}", time.Now().Format(time.RFC3339))
	})

	// Serve static files
	staticDir := "./web/dist"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "./dist"
	}

	// Print info about where static files will be served from
	slog.Info("Serving static files", slog.String("dir", staticDir))

	// Set up static file server
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server", slog.String("port", serverPort))
		if err := server.ListenAndServe(); err != nil && !strings.Contains(err.Error(), "Server closed") {
			slog.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-signals
	slog.Info("Shutdown signal received")

	// Create a timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 15*time.Second)
	defer shutdownCancel()

	registry.GetInstance().Close()

	// Shutdown gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown complete")
}

func setupAPI(ctx context.Context, router *mux.Router) {
	connManager, repoFactory, dbErr := db.Setup(ctx) // Capture connManager
	if dbErr != nil {
		slog.Error("Failed to setup database", slog.String("error", dbErr.Error()))
		// Decide if the application should exit or continue without DB functionality
		// For now, let's return, preventing API setup without DB.
		return
	}
	dbConn := connManager.GetDB() // Get the *sql.DB connection

	wsManager := api.NewWebSocketManager()
	logger := api.NewWebSocketLogger(wsManager)
	debugManager := engine.NewDebugManager()
	errorManager := bperrors.NewErrorManager()
	recoveryManager := bperrors.NewRecoveryManager(errorManager)

	flowEngine := engine.NewExecutionEngine(logger, debugManager)
	// Create the concrete EventManager, passing the engine as the controller
	eventManager := event.NewEventManager(flowEngine)
	// Get the core interface adapter directly from the concrete manager
	eventManagerCoreAdapter := eventManager.AsEventManagerInterface()
	// eventManagerAdapter := event.NewEventManagerAdapter(eventManager) // Remove usage of the separate adapter

	flowEngine.SetExecutionMode(engine.ModeActor)

	contextManager := engineext.NewContextManager(
		errorManager,
		recoveryManager,
		eventManagerCoreAdapter, // Pass the core interface adapter
		repoFactory,             // Pass the repoFactory
	)
	// engineext.InitializeExtensions expects the concrete manager
	contextExtension := engineext.InitializeExtensions(
		flowEngine,
		contextManager,
		errorManager,
		recoveryManager,
		eventManager, // Pass the concrete *event.EventManager
	)

	flowEngine.SetExtensions(contextExtension)

	server := api.NewAPIServerWithDB(
		flowEngine,
		wsManager,
		debugManager,
		repoFactory,
		dbConn,
		contextExtension,
		logger,
	)

	server.InitiateCoreNodes()
	server.SetupRoutes(router)
	go server.ListenRuntimeNodes()
}

type headlessData struct {
	engine      *engine.ExecutionEngine
	connManager *db.ConnectionManager
	repoFactory repository.RepositoryFactory
	logger      *api.WebSocketLogger
}

func setupHeadless(ctx context.Context) *headlessData {
	connManager, repoFactory, dbErr := db.Setup(ctx) // Capture connManager
	if dbErr != nil {
		slog.Error("Failed to setup database", slog.String("error", dbErr.Error()))
		// Decide if the application should exit or continue without DB functionality
		// For now, let's return, preventing API setup without DB.
		return nil
	}

	wsManager := api.NewWebSocketManager()
	logger := api.NewWebSocketLogger(wsManager)
	debugManager := engine.NewDebugManager()
	errorManager := bperrors.NewErrorManager()
	recoveryManager := bperrors.NewRecoveryManager(errorManager)

	flowEngine := engine.NewExecutionEngine(logger, debugManager)
	// Create the concrete EventManager, passing the engine as the controller
	eventManager := event.NewEventManager(flowEngine)
	// Get the core interface adapter directly from the concrete manager
	eventManagerCoreAdapter := eventManager.AsEventManagerInterface()
	// eventManagerAdapter := event.NewEventManagerAdapter(eventManager) // Remove usage of the separate adapter

	flowEngine.SetExecutionMode(engine.ModeActor)

	contextManager := engineext.NewContextManager(
		errorManager,
		recoveryManager,
		eventManagerCoreAdapter, // Pass the core interface adapter
		repoFactory,             // Pass the repoFactory
	)
	// engineext.InitializeExtensions expects the concrete manager
	contextExtension := engineext.InitializeExtensions(
		flowEngine,
		contextManager,
		errorManager,
		recoveryManager,
		eventManager, // Pass the concrete *event.EventManager
	)

	flowEngine.SetExtensions(contextExtension)

	dbConn := connManager.GetDB()

	server := api.NewAPIServerWithDB(
		flowEngine,
		wsManager,
		debugManager,
		repoFactory,
		dbConn,
		contextExtension,
		logger,
	)

	server.InitiateCoreNodes()

	return &headlessData{
		engine:      flowEngine,
		connManager: connManager,
		repoFactory: repoFactory,
		logger:      logger,
	}
}

func headless(bpId, path *string) {
	log.Println("Cleaning log file...")
	if err := os.Remove("./log.out"); err != nil {
		log.Println("Failed to remove old log file")
	}

	log.Println("Cleaning up done")

	if bpId == nil {
		slog.Error("Headless bpId is required")
		return
	}

	var bp *blueprint.Blueprint

	registry.Make()

	setupData := setupHeadless(context.Background())

	bpModel, err := setupData.repoFactory.GetBlueprintRepository().GetByID(context.Background(), *bpId)
	if err != nil {
		slog.Error("Failed to get Blueprint", slog.String("id", bp.ID))
		return
	}

	bp, _ = setupData.repoFactory.GetBlueprintRepository().ToPkgBlueprint(bpModel, bpModel.CurrentVersion)
	executionId := uuid.New().String()

	for _, variable := range bp.Variables {
		registry.GetInstance().RegisterNodeTypeRuntime(fmt.Sprintf("variable-get-%s", variable.Name), data.NewVariableGetDefinedNode(variable.Name, variable.Type, variable.Type))
		registry.GetInstance().RegisterNodeTypeRuntime(fmt.Sprintf("variable-set-%s", variable.Name), data.NewVariableSetDefinedNode(variable.Name, variable.Type, variable.Value))
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(bp *blueprint.Blueprint, wg *sync.WaitGroup) {
		defer wg.Done()
		_, err := setupData.engine.Execute(bp, executionId, make(map[string]types.Value))
		if err != nil {
			slog.Error("Failed to execute blueprint",
				slog.String("id", bp.ID),
				slog.Any("error", err.Error()),
			)
			return
		}
	}(bp, &wg)

	wg.Wait()

	if path != nil {
		contents, err := json.Marshal(bp)
		if err != nil {
			slog.Error("Failed to marshal blueprint",
				slog.String("id", bp.ID),
				slog.Any("error", err.Error()))
			return
		}

		err = os.WriteFile(*path, contents, 0644)
		if err != nil {
			slog.Error("Failed to take snapshot of blueprint",
				slog.String("id", bp.ID),
				slog.Any("error", err.Error()))
			return
		}

		slog.Info("Successfully take snapshot of blueprint",
			slog.String("id", bp.ID))
	}

	registry.GetInstance().Close()

	setupData.logger.Close()
}

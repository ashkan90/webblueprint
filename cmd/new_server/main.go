package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"webblueprint/internal/api"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/engineext"
	"webblueprint/internal/event"
	"webblueprint/pkg/db"
)

func main() {
	// Define command line flags
	port := flag.String("port", "8089", "Server port (default: 8089 or $PORT environment variable)")
	flag.Parse()

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

	// Shutdown gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown complete")
}

func setupAPI(ctx context.Context, router *mux.Router) {
	_, repoFactory, dbErr := db.Setup(ctx)
	if dbErr != nil {
		return
	}

	wsManager := api.NewWebSocketManager()
	logger := api.NewWebSocketLogger(wsManager)
	debugManager := engine.NewDebugManager()
	errorManager := bperrors.NewErrorManager()
	recoveryManager := bperrors.NewRecoveryManager(errorManager)

	flowEngine := engine.NewExecutionEngine(logger, debugManager)
	eventManager := event.NewEventManager(flowEngine)
	eventManagerAdapter := event.NewEventManagerAdapter(eventManager)
	contextManager := engineext.NewContextManager(
		errorManager,
		recoveryManager,
		eventManagerAdapter,
	)
	contextExtension := engineext.InitializeExtensions(
		flowEngine,
		contextManager,
		errorManager,
		recoveryManager,
		eventManagerAdapter,
	)

	server := api.NewAPIServerWithDB(
		flowEngine,
		wsManager,
		debugManager,
		repoFactory,
		contextExtension,
	)

	server.SetupRoutes(router)
}

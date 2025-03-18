package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/registry"
	"webblueprint/internal/types"
	"webblueprint/pkg/repository"
	"webblueprint/pkg/service"
)

// APIServer handles HTTP API requests with repository integration
type APIServerWithDB struct {
	executionEngine  *engine.ExecutionEngine
	errorAwareEngine *engine.ErrorAwareEngine
	nodeRegistry     map[string]node.NodeFactory
	wsManager        *WebSocketManager
	debugManager     *engine.DebugManager
	errorManager     *bperrors.ErrorManager
	recoveryManager  *bperrors.RecoveryManager
	repoFactory      repository.RepositoryFactory
	blueprintService *service.BlueprintService
	userService      *service.UserService
	workspaceService *service.WorkspaceService
	executionService *service.ExecutionService
	rw               *sync.RWMutex
}

// NewAPIServerWithDB creates a new API server with database integration
func NewAPIServerWithDB(
	executionEngine *engine.ExecutionEngine,
	wsManager *WebSocketManager,
	debugManager *engine.DebugManager,
	repoFactory repository.RepositoryFactory,
	errorAwareEngine *engine.ErrorAwareEngine, // Add errorAwareEngine parameter
) *APIServerWithDB {
	// Create the blueprint service
	blueprintService := service.NewBlueprintService(
		repoFactory.GetBlueprintRepository(),
		repoFactory.GetWorkspaceRepository(),
		repoFactory.GetAssetRepository(),
		repoFactory.GetExecutionRepository(),
	)

	userService := service.NewUserService(repoFactory.GetUserRepository())
	executionService := service.NewExecutionService(
		repoFactory.GetExecutionRepository(),
		repoFactory.GetBlueprintRepository(),
		executionEngine,
	)

	// Pass the blueprint repository to the workspace service
	workspaceService := service.NewWorkspaceService(
		repoFactory.GetWorkspaceRepository(),
		repoFactory.GetUserRepository(),
		repoFactory.GetAssetRepository(),
		repoFactory.GetBlueprintRepository(),
	)

	// Get error handling components from the errorAwareEngine
	errorManager := errorAwareEngine.GetErrorManager()
	recoveryManager := errorAwareEngine.GetRecoveryManager()

	// Register WebSocket handlers with error manager
	wsManager.RegisterErrorHandlers(errorManager, nil)

	return &APIServerWithDB{
		executionEngine:  executionEngine,
		errorAwareEngine: errorAwareEngine,
		nodeRegistry:     make(map[string]node.NodeFactory),
		wsManager:        wsManager,
		debugManager:     debugManager,
		errorManager:     errorManager,
		recoveryManager:  recoveryManager,
		repoFactory:      repoFactory,
		blueprintService: blueprintService,
		userService:      userService,
		workspaceService: workspaceService,
		executionService: executionService,
		rw:               &sync.RWMutex{},
	}
}

// RegisterNodeType registers a node type with both the execution engine and API server
func (s *APIServerWithDB) RegisterNodeType(typeID string, factory node.NodeFactory) {
	// This method stays the same as the original API server
	// Store locally
	s.nodeRegistry[typeID] = factory

	// Register with execution engine
	s.executionEngine.RegisterNodeType(typeID, factory)

	// Create a node instance to get metadata
	nodeInstance := factory()
	metadata := nodeInstance.GetMetadata()

	// Broadcast node type to connected clients
	s.wsManager.BroadcastMessage(MsgTypeNodeIntro, map[string]interface{}{
		"typeId":      metadata.TypeID,
		"name":        metadata.Name,
		"description": metadata.Description,
		"category":    metadata.Category,
		"version":     metadata.Version,
		"inputs":      convertPinsToInfo(nodeInstance.GetInputPins()),
		"outputs":     convertPinsToInfo(nodeInstance.GetOutputPins()),
		"properties":  convertPropertiesToInfo(nodeInstance.GetProperties()),
	})
}

// SetupRoutes sets up the HTTP routes for the API server
func (s *APIServerWithDB) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", s.wsManager.HandleWebSocket)

	// Create a blueprint handler
	blueprintHandler := NewBlueprintHandler(s.blueprintService)
	blueprintHandler.RegisterRoutes(r)

	userHandler := NewUserHandler(s.userService)
	userHandler.RegisterRoutes(r)

	workspaceHandler := NewWorkspaceHandler(s.workspaceService)
	workspaceHandler.RegisterRoutes(r)

	executionHandler := NewExecutionHandler(s.executionService)
	executionHandler.RegisterRoutes(r)

	// API endpoints that aren't handled by the blueprint handler
	api := r.PathPrefix("/api").Subrouter()

	// Node types
	api.HandleFunc("/nodes", s.handleGetNodeTypes).Methods("GET")

	// Setup error API
	s.setupErrorAPI(api)

	// Serve static files for the frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))

	return r
}

// setupErrorAPI sets up API endpoints for error handling
func (s *APIServerWithDB) setupErrorAPI(router *mux.Router) {
	// Create API handler for errors
	errorAPI := NewErrorHandlingAPI(
		s.errorManager,
		s.recoveryManager,
		nil, // We'll need to implement this if needed
	)

	// Register API routes for error handling
	router.HandleFunc("/errors/list", errorAPI.HandleListErrors).Methods("GET")
	router.HandleFunc("/errors/analysis", errorAPI.HandleGetErrorAnalysis).Methods("GET")
	router.HandleFunc("/errors/info", errorAPI.HandleGetExecutionInfo).Methods("GET")
	router.HandleFunc("/errors/recover", errorAPI.HandleAttemptRecovery).Methods("POST")
	router.HandleFunc("/errors/clear", errorAPI.HandleClearErrors).Methods("POST")

	// Setup test API for development
	testErrorAPI := NewTestErrorHandler(
		s.errorManager,
		s.recoveryManager,
		s.wsManager,
	)

	router.HandleFunc("/test/generate-error", testErrorAPI.HandleGenerateTestError).Methods("POST")
	router.HandleFunc("/test/generate-scenario", testErrorAPI.HandleGenerateErrorScenario).Methods("POST")
}

func (s *APIServerWithDB) handleGetNodeTypes(w http.ResponseWriter, request *http.Request) {
	nodeTypes := make([]map[string]interface{}, 0)

	for _, factory := range registry.GetInstance().GetAllNodeFactories() {
		_node := factory()
		metadata := _node.GetMetadata()

		nodeType := map[string]interface{}{
			"typeId":      metadata.TypeID,
			"name":        metadata.Name,
			"description": metadata.Description,
			"category":    metadata.Category,
			"version":     metadata.Version,
			"inputs":      convertPinsToInfo(_node.GetInputPins()),
			"outputs":     convertPinsToInfo(_node.GetOutputPins()),
			"properties":  convertPropertiesToInfo(_node.GetProperties()),
		}

		nodeTypes = append(nodeTypes, nodeType)
	}

	respondWithJSON(w, http.StatusOK, nodeTypes)
}

// GetErrorAwareEngine returns the error-aware execution engine
func (s *APIServerWithDB) GetErrorAwareEngine() *engine.ErrorAwareEngine {
	return s.errorAwareEngine
}

// convertPinsToInfo converts pins to a format suitable for the client
func convertPinsToInfo(pins []types.Pin) []map[string]interface{} {
	result := make([]map[string]interface{}, len(pins))

	for i, pin := range pins {
		result[i] = map[string]interface{}{
			"id":          pin.ID,
			"name":        pin.Name,
			"description": pin.Description,
			"type": map[string]string{
				"id":          pin.Type.ID,
				"name":        pin.Type.Name,
				"description": pin.Type.Description,
			},
			"optional": pin.Optional,
		}

		if pin.Default != nil {
			result[i]["default"] = pin.Default
		}
	}

	return result
}

func convertPropertiesToInfo(pins []types.Property) []map[string]interface{} {
	result := make([]map[string]interface{}, len(pins))
	defaultType := map[string]string{
		"id":          types.PinTypes.Any.ID,
		"displayName": types.PinTypes.Any.Name,
		"name":        types.PinTypes.Any.Name,
		"description": types.PinTypes.Any.Description,
	}

	for i, pin := range pins {
		if pin.Type != nil {
			defaultType = map[string]string{
				"id":          pin.Type.ID,
				"name":        pin.Type.Name,
				"description": pin.Type.Description,
			}
		}

		result[i] = map[string]interface{}{
			"name":        pin.Name,
			"displayName": pin.DisplayName,
			"description": pin.Description,
			"value":       pin.Value,
			"type":        defaultType,
		}
	}

	return result
}

// Response helpers
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

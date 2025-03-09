package api

import (
	"encoding/json"
	"log"
	"net/http"
	"webblueprint/internal/db"
	"webblueprint/internal/engine"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/registry"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"

	"github.com/gorilla/mux"
)

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

// SetupRoutes sets up the HTTP routes for the API server
func (s *APIServer) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", s.wsManager.HandleWebSocket)

	// API endpoints
	api := r.PathPrefix("/api").Subrouter()

	// Node types
	api.HandleFunc("/nodes", s.handleGetNodeTypes).Methods("GET")

	// Blueprints
	api.HandleFunc("/blueprints", s.handleGetBlueprints).Methods("GET")
	api.HandleFunc("/blueprints", s.handleCreateBlueprint).Methods("POST")
	api.HandleFunc("/blueprints/{id}", s.handleGetBlueprint).Methods("GET")
	api.HandleFunc("/blueprints/{id}", s.handleUpdateBlueprint).Methods("PUT")
	api.HandleFunc("/blueprints/{id}", s.handleDeleteBlueprint).Methods("DELETE")
	api.HandleFunc("/blueprints/{id}/variable", s.handleCreateVariable).Methods("POST")
	api.HandleFunc("/blueprints/{id}/execute", s.handleExecuteBlueprint).Methods("POST")

	// Debug data
	api.HandleFunc("/executions/{id}", s.handleGetExecution).Methods("GET")
	api.HandleFunc("/executions/{id}/nodes/{nodeId}", s.handleGetNodeDebugData).Methods("GET")

	// Engine configuration
	api.HandleFunc("/engine/config", s.handleGetEngineConfig).Methods("GET")
	api.HandleFunc("/engine/config", s.handleUpdateEngineConfig).Methods("PUT")

	// Serve static files for the frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))

	return r
}

// Response helpers
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	//response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Engine configuration handlers
func (s *APIServer) handleGetEngineConfig(w http.ResponseWriter, r *http.Request) {
	// Get engine configuration
	config := map[string]interface{}{
		"executionMode": s.executionEngine.GetExecutionMode(),
	}

	respondWithJSON(w, http.StatusOK, config)
}

func (s *APIServer) handleUpdateEngineConfig(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var config struct {
		ExecutionMode string `json:"executionMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Update execution mode if provided
	if config.ExecutionMode != "" {
		var mode engine.ExecutionMode
		switch config.ExecutionMode {
		case "actor":
			mode = engine.ModeActor
		case "standard":
			mode = engine.ModeStandard
		default:
			respondWithError(w, http.StatusBadRequest, "Invalid execution mode: must be 'standard' or 'actor'")
			return
		}

		s.executionEngine.SetExecutionMode(mode)
	}

	// Return updated configuration
	updatedConfig := map[string]interface{}{
		"executionMode": s.executionEngine.GetExecutionMode(),
	}

	respondWithJSON(w, http.StatusOK, updatedConfig)
}

// Node types handler
func (s *APIServer) handleGetNodeTypes(w http.ResponseWriter, r *http.Request) {
	nodeTypes := make([]map[string]interface{}, 0)

	for _, factory := range s.nodeRegistry {
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
		}

		nodeTypes = append(nodeTypes, nodeType)
	}

	respondWithJSON(w, http.StatusOK, nodeTypes)
}

// Blueprint handlers
func (s *APIServer) handleGetBlueprints(w http.ResponseWriter, r *http.Request) {
	blueprintList := make([]*blueprint.Blueprint, 0, len(db.Blueprints))
	for _, bp := range db.Blueprints {
		blueprintList = append(blueprintList, bp)
	}

	respondWithJSON(w, http.StatusOK, blueprintList)
}

func (s *APIServer) handleCreateBlueprint(w http.ResponseWriter, r *http.Request) {
	var bp blueprint.Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid blueprint format")
		return
	}

	// Process functions in the blueprint
	for i, function := range bp.Functions {
		log.Printf("Registering user function: %s", function.Name)

		// Ensure functions have execution pins in their NodeType
		hasExecutionInput := false
		hasExecutionOutput := false

		for _, input := range function.NodeType.Inputs {
			if input.Type != nil && input.Type.ID == "execution" {
				hasExecutionInput = true
				break
			}
		}

		for _, output := range function.NodeType.Outputs {
			if output.Type != nil && output.Type.ID == "execution" {
				hasExecutionOutput = true
				break
			}
		}

		// Add default execution input if needed
		if !hasExecutionInput {
			bp.Functions[i].NodeType.Inputs = append(bp.Functions[i].NodeType.Inputs,
				blueprint.NodePin{
					ID:          "execute",
					Name:        "Execute",
					Description: "Execution input",
					Type: &blueprint.NodePinType{
						ID:          "execution",
						Name:        "Execution",
						Description: "Execution flow",
					},
					Optional: false,
				})
		}

		// Add default execution output if needed
		if !hasExecutionOutput {
			bp.Functions[i].NodeType.Outputs = append(bp.Functions[i].NodeType.Outputs,
				blueprint.NodePin{
					ID:          "then",
					Name:        "Then",
					Description: "Execution output",
					Type: &blueprint.NodePinType{
						ID:          "execution",
						Name:        "Execution",
						Description: "Execution flow",
					},
					Optional: false,
				})
		}

		// Register the function as a node type
		nodeFactory := utility.NewUserFunctionNode(bp.Functions[i])

		// Register with API server (which will also register with the global registry)
		s.RegisterNodeType(function.Name, nodeFactory)
	}

	// Process variables in the blueprint
	for _, variable := range bp.Variables {
		// Create node factories for this variable
		getterFactory := data.NewVariableGetNodeFor(variable.Name, variable.Type)
		setterFactory := data.NewVariableSetNodeFor(variable.Name, variable.Type)

		// Register with API server (which will also register with the global registry)
		getterTypeID := "get-variable-" + variable.Name
		setterTypeID := "set-variable-" + variable.Name

		s.RegisterNodeType(getterTypeID, getterFactory)
		s.RegisterNodeType(setterTypeID, setterFactory)
	}

	// Store blueprint
	s.rw.Lock()
	defer s.rw.Unlock()
	db.Blueprints[bp.ID] = &bp

	// Register with execution engine
	s.executionEngine.LoadBlueprint(&bp)

	respondWithJSON(w, http.StatusCreated, bp)
}

func (s *APIServer) handleGetBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.rw.RLock()
	defer s.rw.RUnlock()
	bp, exists := db.Blueprints[id]
	if !exists {
		respondWithError(w, http.StatusNotFound, "Blueprint not found")
		return
	}

	respondWithJSON(w, http.StatusOK, bp)
}

func (s *APIServer) handleUpdateBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var bp blueprint.Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid blueprint format")
		return
	}

	// Ensure IDs match
	if bp.ID != id {
		respondWithError(w, http.StatusBadRequest, "Blueprint ID mismatch")
		return
	}

	// Process functions in the blueprint
	for i, function := range bp.Functions {
		log.Printf("Registering user function: %s", function.Name)

		// Ensure functions have execution pins in their NodeType
		hasExecutionInput := false
		hasExecutionOutput := false

		for _, input := range function.NodeType.Inputs {
			if input.Type != nil && input.Type.ID == "execution" {
				hasExecutionInput = true
				break
			}
		}

		for _, output := range function.NodeType.Outputs {
			if output.Type != nil && output.Type.ID == "execution" {
				hasExecutionOutput = true
				break
			}
		}

		// Add default execution input if needed
		if !hasExecutionInput {
			bp.Functions[i].NodeType.Inputs = append(bp.Functions[i].NodeType.Inputs,
				blueprint.NodePin{
					ID:          "execute",
					Name:        "Execute",
					Description: "Execution input",
					Type: &blueprint.NodePinType{
						ID:          "execution",
						Name:        "Execution",
						Description: "Execution flow",
					},
					Optional: false,
				})
		}

		// Add default execution output if needed
		if !hasExecutionOutput {
			bp.Functions[i].NodeType.Outputs = append(bp.Functions[i].NodeType.Outputs,
				blueprint.NodePin{
					ID:          "then",
					Name:        "Then",
					Description: "Execution output",
					Type: &blueprint.NodePinType{
						ID:          "execution",
						Name:        "Execution",
						Description: "Execution flow",
					},
					Optional: false,
				})
		}

		// Register the function as a node type
		nodeFactory := utility.NewUserFunctionNode(bp.Functions[i])

		// Register with API server (which will also register with the global registry)
		s.RegisterNodeType(function.Name, nodeFactory)
	}

	// Process variables in the blueprint
	for _, variable := range bp.Variables {
		// Create node factories for this variable
		getterFactory := data.NewVariableGetNodeFor(variable.Name, variable.Type)
		setterFactory := data.NewVariableSetNodeFor(variable.Name, variable.Type)

		// Register with API server (which will also register with the global registry)
		getterTypeID := "get-variable-" + variable.Name
		setterTypeID := "set-variable-" + variable.Name

		s.RegisterNodeType(getterTypeID, getterFactory)
		s.RegisterNodeType(setterTypeID, setterFactory)
	}

	// Update blueprint
	s.rw.Lock()
	defer s.rw.Unlock()
	db.Blueprints[id] = &bp

	// Re-register with execution engine
	err := s.executionEngine.LoadBlueprint(&bp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, bp)
}

func (s *APIServer) handleDeleteBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.rw.Lock()
	defer s.rw.Unlock()
	if _, exists := db.Blueprints[id]; !exists {
		respondWithError(w, http.StatusNotFound, "Blueprint not found")
		return
	}

	delete(db.Blueprints, id)

	w.WriteHeader(http.StatusNoContent)
}

type CreateBlueprintVariableReqeust struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (s *APIServer) handleCreateVariable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.rw.RLock()
	bp, exists := db.Blueprints[id]
	s.rw.RUnlock()
	if !exists {
		respondWithError(w, http.StatusNotFound, "Blueprint not found")
		return
	}

	var req CreateBlueprintVariableReqeust
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload "+err.Error())
		return
	}

	s.rw.Lock()
	defer s.rw.Unlock()
	variable := blueprint.Variable{
		Name:  req.Name,
		Type:  req.Type,
		Value: req.Value,
	}
	bp.Variables = append(bp.Variables, variable)

	log.Printf("Registering variable nodes for: %s", req.Name)

	// Create node factories for this variable
	getterFactory := data.NewVariableGetNodeFor(req.Name, req.Type)
	setterFactory := data.NewVariableSetNodeFor(req.Name, req.Type)

	// Register with API server (which will also register with the global registry)
	s.RegisterNodeType("get-variable-"+req.Name, getterFactory)
	s.RegisterNodeType("set-variable-"+req.Name, setterFactory)

	respondWithJSON(w, http.StatusCreated, variable)
}

// Blueprint execution
type ExecuteBlueprintRequest struct {
	InitialVariables map[string]interface{} `json:"initialVariables,omitempty"`
}

func (s *APIServer) handleExecuteBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the blueprint
	s.rw.RLock()
	bp, exists := db.Blueprints[id]
	s.rw.RUnlock()

	if !exists {
		respondWithError(w, http.StatusNotFound, "Blueprint not found")
		return
	}

	// Parse request (optional)
	var req ExecuteBlueprintRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request format")
			return
		}
	}

	// Convert initial variables to types.Value
	initialVars := make(map[string]types.Value)
	for k, v := range req.InitialVariables {
		// Determine type based on Go type
		var pinType *types.PinType

		switch v.(type) {
		case string:
			pinType = types.PinTypes.String
		case float64:
			pinType = types.PinTypes.Number
		case bool:
			pinType = types.PinTypes.Boolean
		case map[string]interface{}:
			pinType = types.PinTypes.Object
		case []interface{}:
			pinType = types.PinTypes.Array
		default:
			pinType = types.PinTypes.Any
		}

		initialVars[k] = types.NewValue(pinType, v)
	}

	// Ensure all functions in the blueprint are registered
	for i, function := range bp.Functions {
		// Check if the function is already registered
		globalRegistry := registry.GetInstance()
		if _, exists := globalRegistry.GetNodeFactory(function.Name); !exists {
			log.Printf("Registering missing user function: %s", function.Name)

			// Ensure functions have execution pins in their NodeType
			hasExecutionInput := false
			hasExecutionOutput := false

			for _, input := range function.NodeType.Inputs {
				if input.Type != nil && input.Type.ID == "execution" {
					hasExecutionInput = true
					break
				}
			}

			for _, output := range function.NodeType.Outputs {
				if output.Type != nil && output.Type.ID == "execution" {
					hasExecutionOutput = true
					break
				}
			}

			// Add default execution input if needed
			if !hasExecutionInput {
				bp.Functions[i].NodeType.Inputs = append(bp.Functions[i].NodeType.Inputs,
					blueprint.NodePin{
						ID:          "execute",
						Name:        "Execute",
						Description: "Execution input",
						Type: &blueprint.NodePinType{
							ID:          "execution",
							Name:        "Execution",
							Description: "Execution flow",
						},
						Optional: false,
					})
			}

			// Add default execution output if needed
			if !hasExecutionOutput {
				bp.Functions[i].NodeType.Outputs = append(bp.Functions[i].NodeType.Outputs,
					blueprint.NodePin{
						ID:          "then",
						Name:        "Then",
						Description: "Execution output",
						Type: &blueprint.NodePinType{
							ID:          "execution",
							Name:        "Execution",
							Description: "Execution flow",
						},
						Optional: false,
					})
			}

			// Register the function as a node type
			nodeFactory := utility.NewUserFunctionNode(bp.Functions[i])

			// Register with API server
			s.RegisterNodeType(function.Name, nodeFactory)
		}
	}

	// Ensure all variables in the blueprint are registered
	for _, variable := range bp.Variables {
		// Check if variable nodes are already registered
		globalRegistry := registry.GetInstance()
		getterName := "get-variable-" + variable.Name
		setterName := "set-variable-" + variable.Name

		if _, getExists := globalRegistry.GetNodeFactory(getterName); !getExists {
			log.Printf("Registering missing variable getter: %s", getterName)
			getterFactory := data.NewVariableGetNodeFor(variable.Name, variable.Type)
			s.RegisterNodeType(getterName, getterFactory)
		}

		if _, setExists := globalRegistry.GetNodeFactory(setterName); !setExists {
			log.Printf("Registering missing variable setter: %s", setterName)
			setterFactory := data.NewVariableSetNodeFor(variable.Name, variable.Type)
			s.RegisterNodeType(setterName, setterFactory)
		}
	}

	// Execute the blueprint
	result, err := s.executionEngine.Execute(id, initialVars)
	if err != nil {
		log.Printf("Error executing blueprint: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Execution failed: "+err.Error())
		return
	}

	// Return the execution result
	response := map[string]interface{}{
		"executionId": result.ExecutionID,
		"success":     result.Success,
		"startTime":   result.StartTime,
		"endTime":     result.EndTime,
		"duration":    result.EndTime.Sub(result.StartTime).String(),
	}

	if !result.Success && result.Error != nil {
		response["error"] = result.Error.Error()
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Debug data handlers
func (s *APIServer) handleGetExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status, exists := s.executionEngine.GetExecutionStatus(id)
	if !exists {
		respondWithError(w, http.StatusNotFound, "Execution not found")
		return
	}

	// Convert to response format
	response := map[string]interface{}{
		"executionId": status.ExecutionID,
		"status":      status.Status,
		"startTime":   status.StartTime,
		"endTime":     status.EndTime,
		"duration":    status.EndTime.Sub(status.StartTime).String(),
		"nodes":       map[string]interface{}{},
	}

	// Add node statuses
	for nodeID, nodeStatus := range status.NodeStatuses {
		response["nodes"].(map[string]interface{})[nodeID] = map[string]interface{}{
			"status":    nodeStatus.Status,
			"startTime": nodeStatus.StartTime,
			"endTime":   nodeStatus.EndTime,
		}

		if nodeStatus.Error != nil {
			response["nodes"].(map[string]interface{})[nodeID].(map[string]interface{})["error"] = nodeStatus.Error.Error()
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleGetNodeDebugData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	executionID := vars["id"]
	nodeID := vars["nodeId"]

	// Check if execution exists
	status, exists := s.executionEngine.GetExecutionStatus(executionID)
	if !exists {
		respondWithError(w, http.StatusNotFound, "Execution not found")
		return
	}

	// Check if node exists in this execution
	nodeStatus, exists := status.NodeStatuses[nodeID]
	if !exists {
		respondWithError(w, http.StatusNotFound, "Node not found in execution")
		return
	}

	// Get debug data from debug manager
	debugData, exists := s.debugManager.GetNodeDebugData(executionID, nodeID)
	if !exists {
		// Return empty debug data
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"nodeId":      nodeID,
			"executionId": executionID,
			"status":      nodeStatus.Status,
			"debug":       map[string]interface{}{},
		})
		return
	}

	// Return debug data
	response := map[string]interface{}{
		"nodeId":      nodeID,
		"executionId": executionID,
		"status":      nodeStatus.Status,
		"debug":       debugData,
	}

	if nodeStatus.Error != nil {
		response["error"] = nodeStatus.Error.Error()
	}

	respondWithJSON(w, http.StatusOK, response)
}

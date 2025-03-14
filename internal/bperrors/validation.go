package bperrors

import (
	"webblueprint/internal/common"
	"webblueprint/pkg/blueprint"
)

// BlueprintValidator validates blueprint structures
type BlueprintValidator struct {
	errorManager *ErrorManager
}

// NewBlueprintValidator creates a new blueprint validator
func NewBlueprintValidator(errorManager *ErrorManager) *BlueprintValidator {
	return &BlueprintValidator{
		errorManager: errorManager,
	}
}

// ValidateBlueprint performs validation checks on a blueprint
func (v *BlueprintValidator) ValidateBlueprint(bp *blueprint.Blueprint) common.ValidationResult {
	result := common.ValidationResult{
		Valid:      true,
		Errors:     make([]error, 0),
		Warnings:   make([]error, 0),
		NodeIssues: make(map[string][]string),
	}

	// Basic structure validation
	if bp.ID == "" {
		err := New(ErrorTypeValidation, ErrInvalidBlueprintStructure, "Blueprint is missing an ID", SeverityHigh)
		result.Errors = append(result.Errors, err)
		result.Valid = false
	}

	if bp.Name == "" {
		err := New(ErrorTypeValidation, ErrInvalidBlueprintStructure, "Blueprint is missing a name", SeverityMedium)
		result.Warnings = append(result.Warnings, err)
	}

	if len(bp.Nodes) == 0 {
		err := New(ErrorTypeValidation, ErrInvalidBlueprintStructure, "Blueprint has no nodes", SeverityHigh)
		result.Errors = append(result.Errors, err)
		result.Valid = false
	}

	// Validate nodes
	for _, node := range bp.Nodes {
		nodeIssues := v.validateNode(bp, &node)
		if len(nodeIssues) > 0 {
			result.NodeIssues[node.ID] = nodeIssues
		}
	}

	// Validate connections
	connectionIssues := v.validateConnections(bp)
	if len(connectionIssues) > 0 {
		for _, issue := range connectionIssues {
			if issue.Severity == SeverityHigh || issue.Severity == SeverityCritical {
				result.Errors = append(result.Errors, issue)
				result.Valid = false
			} else {
				result.Warnings = append(result.Warnings, issue)
			}
		}
	}

	// Check for cycles
	if hasCycles := v.checkForCycles(bp); hasCycles {
		err := New(ErrorTypeValidation, ErrCircularDependency, "Blueprint contains circular dependencies", SeverityHigh)
		result.Errors = append(result.Errors, err)
		result.Valid = false
	}

	// Check for entry points
	if entryPoints := bp.FindEntryPoints(); len(entryPoints) == 0 {
		err := New(ErrorTypeValidation, ErrNoEntryPoints, "Blueprint has no entry points", SeverityHigh)
		result.Errors = append(result.Errors, err)
		result.Valid = false
	}

	return result
}

// validateNode validates a single node in a blueprint
func (v *BlueprintValidator) validateNode(bp *blueprint.Blueprint, node *blueprint.BlueprintNode) []string {
	issues := make([]string, 0)

	// Check for empty ID
	if node.ID == "" {
		issues = append(issues, "Node is missing an ID")
	}

	// Check for empty type
	if node.Type == "" {
		issues = append(issues, "Node is missing a type")
	}

	// Check for disconnected nodes
	if len(bp.GetNodeInputConnections(node.ID)) == 0 && len(bp.GetNodeOutputConnections(node.ID)) == 0 {
		if entryPoints := bp.FindEntryPoints(); !contains(entryPoints, node.ID) {
			issues = append(issues, "Node is disconnected from the blueprint")
		}
	}

	return issues
}

// validateConnections validates the connections in a blueprint
func (v *BlueprintValidator) validateConnections(bp *blueprint.Blueprint) []*BlueprintError {
	issues := make([]*BlueprintError, 0)

	// Check all connections for validity
	for _, conn := range bp.Connections {
		// Check for valid source and target nodes
		sourceNode := bp.FindNode(conn.SourceNodeID)
		if sourceNode == nil {
			err := New(ErrorTypeValidation, ErrInvalidConnection,
				"Connection references non-existent source node", SeverityHigh)
			err.WithNodeInfo(conn.SourceNodeID, conn.SourcePinID)
			issues = append(issues, err)
			continue
		}

		targetNode := bp.FindNode(conn.TargetNodeID)
		if targetNode == nil {
			err := New(ErrorTypeValidation, ErrInvalidConnection,
				"Connection references non-existent target node", SeverityHigh)
			err.WithNodeInfo(conn.TargetNodeID, conn.TargetPinID)
			issues = append(issues, err)
			continue
		}
	}

	return issues
}

// checkForCycles detects circular dependencies in the blueprint
func (v *BlueprintValidator) checkForCycles(bp *blueprint.Blueprint) bool {
	// Create a directed graph representation of the blueprint
	graph := make(map[string][]string)
	for _, node := range bp.Nodes {
		graph[node.ID] = make([]string, 0)
	}

	for _, conn := range bp.Connections {
		graph[conn.SourceNodeID] = append(graph[conn.SourceNodeID], conn.TargetNodeID)
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range graph {
		if !visited[nodeID] {
			if v.isCyclicUtil(graph, nodeID, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// isCyclicUtil is a helper function for cycle detection using DFS
func (v *BlueprintValidator) isCyclicUtil(graph map[string][]string, nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, neighbor := range graph[nodeID] {
		if !visited[neighbor] {
			if v.isCyclicUtil(graph, neighbor, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// Helper function to check if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

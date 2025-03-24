package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"webblueprint/pkg/blueprint"
)

// TestBlueprintConversion tests the conversion between different blueprint formats
func TestBlueprintConversion(t *testing.T) {
	// Create a blueprint programmatically
	bp := blueprint.NewBlueprint("test-conversion", "Test Conversion Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "start",
		Type: "start",
		Position: blueprint.Position{
			X: 100,
			Y: 100,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "process",
		Type: "process-a",
		Position: blueprint.Position{
			X: 300,
			Y: 100,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "end",
		Type: "end",
		Position: blueprint.Position{
			X: 500,
			Y: 100,
		},
	})

	// Add connections
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_process",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "process",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process_to_end",
		SourceNodeID:   "process",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	// Convert to JSON
	bpJSON, err := json.Marshal(bp)
	assert.NoError(t, err, "Should be able to marshal blueprint to JSON")

	// Convert back from JSON to Blueprint
	var bp2 blueprint.Blueprint
	err = json.Unmarshal(bpJSON, &bp2)
	assert.NoError(t, err, "Should be able to unmarshal JSON to blueprint")

	// Verify the conversion preserved all data
	assert.Equal(t, bp.ID, bp2.ID, "Blueprint ID should be preserved")
	assert.Equal(t, bp.Name, bp2.Name, "Blueprint name should be preserved")
	assert.Equal(t, bp.Version, bp2.Version, "Blueprint version should be preserved")
	assert.Len(t, bp2.Nodes, len(bp.Nodes), "Same number of nodes should be preserved")
	assert.Len(t, bp2.Connections, len(bp.Connections), "Same number of connections should be preserved")

	// Check specific nodes
	startNode := bp2.FindNode("start")
	assert.NotNil(t, startNode, "Start node should exist")
	assert.Equal(t, "start", startNode.Type, "Start node type should be preserved")

	processNode := bp2.FindNode("process")
	assert.NotNil(t, processNode, "Process node should exist")
	assert.Equal(t, "process-a", processNode.Type, "Process node type should be preserved")

	endNode := bp2.FindNode("end")
	assert.NotNil(t, endNode, "End node should exist")
	assert.Equal(t, "end", endNode.Type, "End node type should be preserved")

	// Check connections
	startToProcess := bp2.FindConnection("start_to_process")
	assert.NotNil(t, startToProcess, "Start to process connection should exist")
	assert.Equal(t, "start", startToProcess.SourceNodeID, "Source node should be preserved")
	assert.Equal(t, "out", startToProcess.SourcePinID, "Source pin should be preserved")
	assert.Equal(t, "process", startToProcess.TargetNodeID, "Target node should be preserved")
	assert.Equal(t, "in", startToProcess.TargetPinID, "Target pin should be preserved")
	assert.Equal(t, "execution", startToProcess.ConnectionType, "Connection type should be preserved")

	processToEnd := bp2.FindConnection("process_to_end")
	assert.NotNil(t, processToEnd, "Process to end connection should exist")
	assert.Equal(t, "process", processToEnd.SourceNodeID, "Source node should be preserved")
	assert.Equal(t, "out", processToEnd.SourcePinID, "Source pin should be preserved")
	assert.Equal(t, "end", processToEnd.TargetNodeID, "Target node should be preserved")
	assert.Equal(t, "in", processToEnd.TargetPinID, "Target pin should be preserved")
	assert.Equal(t, "execution", processToEnd.ConnectionType, "Connection type should be preserved")
}

// TestBlueprintNodeProperties tests the handling of node properties
func TestBlueprintNodeProperties(t *testing.T) {
	// Create a blueprint with nodes that have properties
	bp := blueprint.NewBlueprint("test-properties", "Property Test Blueprint", "1.0.0")

	// Add a node with properties
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "string-node",
		Type: "string-node",
		Position: blueprint.Position{
			X: 100,
			Y: 100,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "text",
				DisplayName: "Text Value",
				Description: "A string property",
				Value:       "test string",
				Type:        nil,
			},
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "number-node",
		Type: "number-node",
		Position: blueprint.Position{
			X: 300,
			Y: 100,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "value",
				DisplayName: "Number Value",
				Description: "A number property",
				Value:       42,
				Type:        nil,
			},
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "boolean-node",
		Type: "boolean-node",
		Position: blueprint.Position{
			X: 500,
			Y: 100,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "flag",
				DisplayName: "Boolean Flag",
				Description: "A boolean property",
				Value:       true,
				Type:        nil,
			},
		},
	})

	// Convert to JSON and back
	bpJSON, err := json.Marshal(bp)
	assert.NoError(t, err, "Should be able to marshal blueprint to JSON")

	var bp2 blueprint.Blueprint
	err = json.Unmarshal(bpJSON, &bp2)
	assert.NoError(t, err, "Should be able to unmarshal JSON to blueprint")

	// Verify properties were preserved
	stringNode := bp2.FindNode("string-node")
	assert.NotNil(t, stringNode, "String node should exist")
	assert.Len(t, stringNode.Properties, 1, "String node should have 1 property")
	assert.Equal(t, "text", stringNode.Properties[0].Name, "Property name should be preserved")
	assert.Equal(t, "test string", stringNode.Properties[0].Value, "Property value should be preserved")

	numberNode := bp2.FindNode("number-node")
	assert.NotNil(t, numberNode, "Number node should exist")
	assert.Len(t, numberNode.Properties, 1, "Number node should have 1 property")
	assert.Equal(t, "value", numberNode.Properties[0].Name, "Property name should be preserved")

	// Note: JSON unmarshaling will convert numbers to float64 by default
	floatValue, ok := numberNode.Properties[0].Value.(float64)
	assert.True(t, ok, "Number should be stored as float64")
	assert.Equal(t, float64(42), floatValue, "Property value should be preserved")

	booleanNode := bp2.FindNode("boolean-node")
	assert.NotNil(t, booleanNode, "Boolean node should exist")
	assert.Len(t, booleanNode.Properties, 1, "Boolean node should have 1 property")
	assert.Equal(t, "flag", booleanNode.Properties[0].Name, "Property name should be preserved")
	assert.Equal(t, true, booleanNode.Properties[0].Value, "Property value should be preserved")
}

// TestBlueprintModification tests the ability to modify blueprints
func TestBlueprintModification(t *testing.T) {
	// Create a blueprint
	bp := blueprint.NewBlueprint("test-modification", "Modification Test Blueprint", "1.0.0")

	// Add some nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "node1",
		Type: "test-node",
		Position: blueprint.Position{
			X: 100,
			Y: 100,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "node2",
		Type: "test-node",
		Position: blueprint.Position{
			X: 300,
			Y: 100,
		},
	})

	// Add a connection
	bp.AddConnection(blueprint.Connection{
		ID:             "conn1",
		SourceNodeID:   "node1",
		SourcePinID:    "out",
		TargetNodeID:   "node2",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	// Now modify the blueprint
	// Remove a node
	bp.RemoveNode("node1")

	// Verify the node is gone
	assert.Nil(t, bp.FindNode("node1"), "Node should be removed")

	// Verify that the connection is also gone
	assert.Nil(t, bp.FindConnection("conn1"), "Connection should be removed with node")

	// Add a new node
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "node3",
		Type: "test-node",
		Position: blueprint.Position{
			X: 500,
			Y: 100,
		},
	})

	// Add a new connection
	bp.AddConnection(blueprint.Connection{
		ID:             "conn2",
		SourceNodeID:   "node2",
		SourcePinID:    "out",
		TargetNodeID:   "node3",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	// Verify the new node and connection exist
	assert.NotNil(t, bp.FindNode("node3"), "New node should be added")
	assert.NotNil(t, bp.FindConnection("conn2"), "New connection should be added")

	// Remove a connection
	bp.RemoveConnection("conn2")

	// Verify the connection is gone
	assert.Nil(t, bp.FindConnection("conn2"), "Connection should be removed")

	// Verify nodes still exist
	assert.NotNil(t, bp.FindNode("node2"), "Node should still exist")
	assert.NotNil(t, bp.FindNode("node3"), "Node should still exist")
}

// TestBlueprintConnectionQueries tests the node connection query functions
func TestBlueprintConnectionQueries(t *testing.T) {
	// Create a blueprint with multiple connections
	bp := blueprint.NewBlueprint("test-connections", "Connection Query Test Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "process1",
		Type:     "process",
		Position: blueprint.Position{X: 300, Y: 50},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "process2",
		Type:     "process",
		Position: blueprint.Position{X: 300, Y: 150},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 500, Y: 100},
	})

	// Add connections
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_process1",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "process1",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_process2",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "process2",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process1_to_end",
		SourceNodeID:   "process1",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process2_to_end",
		SourceNodeID:   "process2",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "data_connection",
		SourceNodeID:   "process1",
		SourcePinID:    "data",
		TargetNodeID:   "process2",
		TargetPinID:    "input",
		ConnectionType: "data",
		Data:           map[string]any{},
	})

	// Test GetNodeConnections
	startConns := bp.GetNodeConnections("start")
	assert.Len(t, startConns, 2, "Start node should have 2 connections")

	process1Conns := bp.GetNodeConnections("process1")
	assert.Len(t, process1Conns, 3, "Process1 node should have 3 connections (1 input, 2 outputs)")

	process2Conns := bp.GetNodeConnections("process2")
	assert.Len(t, process2Conns, 3, "Process2 node should have 3 connections (2 inputs, 1 output)")

	endConns := bp.GetNodeConnections("end")
	assert.Len(t, endConns, 2, "End node should have 2 connections")

	// Test GetNodeInputConnections
	startInputs := bp.GetNodeInputConnections("start")
	assert.Len(t, startInputs, 0, "Start node should have 0 input connections")

	process1Inputs := bp.GetNodeInputConnections("process1")
	assert.Len(t, process1Inputs, 1, "Process1 node should have 1 input connection")

	process2Inputs := bp.GetNodeInputConnections("process2")
	assert.Len(t, process2Inputs, 2, "Process2 node should have 2 input connections (execution + data)")

	endInputs := bp.GetNodeInputConnections("end")
	assert.Len(t, endInputs, 2, "End node should have 2 input connections")

	// Test GetNodeOutputConnections
	startOutputs := bp.GetNodeOutputConnections("start")
	assert.Len(t, startOutputs, 2, "Start node should have 2 output connections")

	process1Outputs := bp.GetNodeOutputConnections("process1")
	assert.Len(t, process1Outputs, 2, "Process1 node should have 2 output connections (execution + data)")

	process2Outputs := bp.GetNodeOutputConnections("process2")
	assert.Len(t, process2Outputs, 1, "Process2 node should have 1 output connection")

	endOutputs := bp.GetNodeOutputConnections("end")
	assert.Len(t, endOutputs, 0, "End node should have 0 output connections")
}

// TestBlueprintEntryPoints tests detection of entry point nodes
func TestBlueprintEntryPoints(t *testing.T) {
	// Create a blueprint with multiple connections
	bp := blueprint.NewBlueprint("test-entry-points", "Entry Points Test Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start1",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start2",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "dom-event",
		Type:     "dom-event",
		Position: blueprint.Position{X: 100, Y: 300},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "process1",
		Type:     "process",
		Position: blueprint.Position{X: 300, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "process2",
		Type:     "process",
		Position: blueprint.Position{X: 300, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "process3",
		Type:     "process",
		Position: blueprint.Position{X: 300, Y: 300},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 500, Y: 200},
	})

	// Add connections
	bp.AddConnection(blueprint.Connection{
		ID:             "start1_to_process1",
		SourceNodeID:   "start1",
		SourcePinID:    "out",
		TargetNodeID:   "process1",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "start2_to_process2",
		SourceNodeID:   "start2",
		SourcePinID:    "out",
		TargetNodeID:   "process2",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "dom_to_process3",
		SourceNodeID:   "dom-event",
		SourcePinID:    "out",
		TargetNodeID:   "process3",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process1_to_end",
		SourceNodeID:   "process1",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process2_to_end",
		SourceNodeID:   "process2",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process3_to_end",
		SourceNodeID:   "process3",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
		Data:           map[string]any{},
	})

	// Find entry points
	entryPoints := bp.FindEntryPoints()

	// We should have three entry points: start1, start2, and dom-event
	assert.Len(t, entryPoints, 3, "Blueprint should have 3 entry points")
	assert.Contains(t, entryPoints, "start1", "start1 should be an entry point")
	assert.Contains(t, entryPoints, "start2", "start2 should be an entry point")
	assert.Contains(t, entryPoints, "dom-event", "dom-event should be an entry point")
}

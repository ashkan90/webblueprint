package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"webblueprint/pkg/blueprint"
)

// The JSON data from the user is a complete blueprint definition with nodes, connections, etc.
// This test verifies that we can parse it correctly and register it with the engine.

// TestBlueprintJsonParsing tests the ability to parse blueprint JSON
func TestBlueprintJsonParsing(t *testing.T) {
	// Sample JSON data
	jsonData := `{
		"id": "861a1c9b-0486-4975-b7f3-45572ae63a83",
		"name": "New Blueprint7",
		"version": "1.0.0",
		"nodes": [
			{
				"id": "0753e85b-a169-4036-97b5-c5c972c815e7",
				"type": "constant-string",
				"position": {
					"x": 1032.3333333333335,
					"y": 468.2222222222223
				},
				"properties": [
					{
						"name": "constantValue",
						"displayName": "",
						"description": "",
						"value": "default value on prop",
						"type": null
					}
				]
			},
			{
				"id": "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
				"type": "print",
				"position": {
					"x": 1265.2222222222224,
					"y": 218.33333333333331
				},
				"properties": null
			}
		],
		"functions": null,
		"connections": [
			{
				"id": "e9c6e13b-61e4-4aab-a6c6-400f52eb8d49",
				"sourceNodeId": "0753e85b-a169-4036-97b5-c5c972c815e7",
				"sourcePinId": "value",
				"targetNodeId": "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
				"targetPinId": "message",
				"connectionType": "data",
				"data": {}
			}
		]
	}`

	// Parse JSON
	var bp blueprint.Blueprint
	err := json.Unmarshal([]byte(jsonData), &bp)
	assert.NoError(t, err, "Should be able to parse blueprint JSON")

	// Verify parsed data
	assert.Equal(t, "861a1c9b-0486-4975-b7f3-45572ae63a83", bp.ID, "Blueprint ID should be preserved")
	assert.Equal(t, "New Blueprint7", bp.Name, "Blueprint name should be preserved")
	assert.Equal(t, "1.0.0", bp.Version, "Blueprint version should be preserved")

	// Check nodes
	assert.Len(t, bp.Nodes, 2, "Blueprint should have 2 nodes")

	// Verify first node (constant-string)
	constantNode := bp.FindNode("0753e85b-a169-4036-97b5-c5c972c815e7")
	assert.NotNil(t, constantNode, "Constant node should exist")
	assert.Equal(t, "constant-string", constantNode.Type, "Node type should be constant-string")
	assert.Len(t, constantNode.Properties, 1, "Constant node should have 1 property")
	assert.Equal(t, "constantValue", constantNode.Properties[0].Name, "Property name should be constantValue")
	assert.Equal(t, "default value on prop", constantNode.Properties[0].Value, "Property value should be correct")

	// Verify connections
	assert.Len(t, bp.Connections, 1, "Blueprint should have 1 connection")
	conn := bp.Connections[0]
	assert.Equal(t, "e9c6e13b-61e4-4aab-a6c6-400f52eb8d49", conn.ID, "Connection ID should be preserved")
	assert.Equal(t, "0753e85b-a169-4036-97b5-c5c972c815e7", conn.SourceNodeID, "Connection source should be correct")
	assert.Equal(t, "value", conn.SourcePinID, "Connection source pin should be correct")
	assert.Equal(t, "fbbd33a5-626b-4a2e-997b-b9eb1647a095", conn.TargetNodeID, "Connection target should be correct")
	assert.Equal(t, "message", conn.TargetPinID, "Connection target pin should be correct")
	assert.Equal(t, "data", conn.ConnectionType, "Connection type should be data")

}

// TestCompleteJsonBlueprint tests using the complete JSON blueprint from the sample
func TestCompleteJsonBlueprint(t *testing.T) {
	// Sample JSON data from the paste
	jsonData := `{
		"id": "861a1c9b-0486-4975-b7f3-45572ae63a83",
		"name": "New Blueprint7",
		"version": "1.0.0",
		"nodes": [
			{
				"id": "0753e85b-a169-4036-97b5-c5c972c815e7",
				"type": "constant-string",
				"position": {
					"x": 1032.3333333333335,
					"y": 468.2222222222223
				},
				"properties": [
					{
						"name": "constantValue",
						"displayName": "",
						"description": "",
						"value": "default value on prop",
						"type": null
					}
				]
			},
			{
				"id": "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
				"type": "print",
				"position": {
					"x": 1265.2222222222224,
					"y": 218.33333333333331
				},
				"properties": null
			},
			{
				"id": "1dda3f01-6d79-4a31-bd7a-6786ceb80abb",
				"type": "http-request",
				"position": {
					"x": 534,
					"y": 195
				},
				"properties": [
					{
						"name": "input_method",
						"displayName": "",
						"description": "",
						"value": "GET",
						"type": null
					},
					{
						"name": "input_url",
						"displayName": "",
						"description": "",
						"value": "https://jsonplaceholder.typicode.com/todos/1",
						"type": null
					}
				],
				"data": {
					"defaults": {
						"url": "https://jsonplaceholder.typicode.com/todos/1"
					}
				}
			},
			{
				"id": "90a3774f-0fe8-4601-915b-c6a721e91153",
				"type": "object-operations",
				"position": {
					"x": 792,
					"y": 155.55555555555566
				},
				"properties": [
					{
						"name": "input_operation",
						"displayName": "",
						"description": "",
						"value": "get",
						"type": null
					},
					{
						"name": "input_key",
						"displayName": "",
						"description": "",
						"value": "title",
						"type": null
					}
				],
				"data": {
					"defaults": {
						"key": "title"
					}
				}
			},
			{
				"id": "50df077e-5303-4cf7-a912-da95f47dc722",
				"type": "print",
				"position": {
					"x": 1052.0000000000007,
					"y": 204.8888888888888
				},
				"properties": null
			},
			{
				"id": "33df614d-e269-4113-ba42-c9794c115e23",
				"type": "variable-set",
				"position": {
					"x": 1173,
					"y": 638.5
				},
				"properties": [
					{
						"name": "input_name",
						"displayName": "",
						"description": "",
						"value": "test_variable",
						"type": null
					},
					{
						"name": "input_value",
						"displayName": "",
						"description": "",
						"value": "default_value",
						"type": null
					}
				],
				"data": {
					"defaults": {
						"name": "test_variable",
						"value": "default_value"
					}
				}
			},
			{
				"id": "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
				"type": "print",
				"position": {
					"x": 1542,
					"y": 261.5
				},
				"properties": null
			},
			{
				"id": "c482a3e0-99bc-427d-91db-b994f683d14c",
				"type": "variable-get",
				"position": {
					"x": 1281,
					"y": 469.5
				},
				"properties": [
					{
						"name": "input_name",
						"displayName": "",
						"description": "",
						"value": "test_variable",
						"type": null
					}
				],
				"data": {
					"defaults": {
						"name": "test_variable"
					}
				}
			},
			{
				"id": "result_setter",
				"type": "set-variable-result",
				"position": {
					"x": 1800,
					"y": 300
				}
			}
		],
		"functions": null,
		"connections": [
			{
				"id": "e9c6e13b-61e4-4aab-a6c6-400f52eb8d49",
				"sourceNodeId": "0753e85b-a169-4036-97b5-c5c972c815e7",
				"sourcePinId": "value",
				"targetNodeId": "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
				"targetPinId": "message",
				"connectionType": "data",
				"data": {}
			},
			{
				"id": "89179cba-e2b2-464f-a947-b60aeacb5c1d",
				"sourceNodeId": "1dda3f01-6d79-4a31-bd7a-6786ceb80abb",
				"sourcePinId": "then",
				"targetNodeId": "90a3774f-0fe8-4601-915b-c6a721e91153",
				"targetPinId": "exec",
				"connectionType": "execution",
				"data": {}
			},
			{
				"id": "569da363-f9dd-47d9-9c52-02f10f735001",
				"sourceNodeId": "1dda3f01-6d79-4a31-bd7a-6786ceb80abb",
				"sourcePinId": "response",
				"targetNodeId": "90a3774f-0fe8-4601-915b-c6a721e91153",
				"targetPinId": "object",
				"connectionType": "data",
				"data": {}
			},
			{
				"id": "e23f9e26-52f2-4e50-8250-cfc12140d015",
				"sourceNodeId": "90a3774f-0fe8-4601-915b-c6a721e91153",
				"sourcePinId": "then",
				"targetNodeId": "50df077e-5303-4cf7-a912-da95f47dc722",
				"targetPinId": "exec",
				"connectionType": "execution",
				"data": {}
			},
			{
				"id": "52d141f7-2b31-47f0-95ca-395e62351de2",
				"sourceNodeId": "90a3774f-0fe8-4601-915b-c6a721e91153",
				"sourcePinId": "result",
				"targetNodeId": "50df077e-5303-4cf7-a912-da95f47dc722",
				"targetPinId": "message",
				"connectionType": "data",
				"data": {}
			},
			{
				"id": "ba14bdce-114b-42c5-98ce-271c1b48aa43",
				"sourceNodeId": "50df077e-5303-4cf7-a912-da95f47dc722",
				"sourcePinId": "then",
				"targetNodeId": "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
				"targetPinId": "exec",
				"connectionType": "execution",
				"data": {}
			},
			{
				"id": "fd3de6f8-96f6-4827-9a5c-304a757b6808",
				"sourceNodeId": "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
				"sourcePinId": "then",
				"targetNodeId": "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
				"targetPinId": "exec",
				"connectionType": "execution",
				"data": {}
			},
			{
				"id": "059a0d4a-ae71-460d-818f-3abee09bab8a",
				"sourceNodeId": "c482a3e0-99bc-427d-91db-b994f683d14c",
				"sourcePinId": "value",
				"targetNodeId": "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
				"targetPinId": "message",
				"connectionType": "data",
				"data": {}
			},
			{
				"id": "constant_to_result",
				"sourceNodeId": "0753e85b-a169-4036-97b5-c5c972c815e7",
				"sourcePinId": "value",
				"targetNodeId": "result_setter",
				"targetPinId": "value",
				"connectionType": "data",
				"data": {}
			},
			{
				"id": "print_to_result_setter",
				"sourceNodeId": "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
				"sourcePinId": "then",
				"targetNodeId": "result_setter",
				"targetPinId": "in",
				"connectionType": "execution",
				"data": {}
			}
		]
	}`

	// Parse JSON
	var bp blueprint.Blueprint
	err := json.Unmarshal([]byte(jsonData), &bp)
	assert.NoError(t, err, "Should be able to parse blueprint JSON")

	// Verify basic parsed data
	assert.Equal(t, "861a1c9b-0486-4975-b7f3-45572ae63a83", bp.ID, "Blueprint ID should be preserved")
	assert.Equal(t, "New Blueprint7", bp.Name, "Blueprint name should be preserved")
	assert.Equal(t, "1.0.0", bp.Version, "Blueprint version should be preserved")
	assert.Len(t, bp.Nodes, 9, "Blueprint should have 9 nodes")
	assert.Len(t, bp.Connections, 10, "Blueprint should have 10 connections")
}

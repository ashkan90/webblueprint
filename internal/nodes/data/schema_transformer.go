package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"webblueprint/internal/db"
	"webblueprint/internal/engineext"
	"webblueprint/internal/node"
	"webblueprint/internal/types"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader" // Enable HTTP(s) loading for $ref
)

// SchemaNode applies a selected schema transformation to input JSON data.
// (Renamed struct and updated description)
type SchemaNode struct {
	node.BaseNode
}

// NewSchemaNode creates a new SchemaNode instance.
// (Renamed function)
func NewSchemaNode() node.Node {
	return &SchemaNode{ // Use renamed struct
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "schema-transformer",                                                 // Updated TypeID
				Name:        "Schema Transformer",                                                 // Kept name
				Description: "Applies a predefined schema component transformation to JSON data.", // Updated description
				Category:    "Data",
				Version:     "1.0.0",
			},
			// Updated Input Pins
			Inputs: []types.Pin{
				{
					ID:          "exec", // Standard execution pin ID
					Name:        "Execute",
					Description: "Triggers the node's processing logic.",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "value", // Input data pin ID
					Name:        "Value",
					Description: "Input JSON data (object or array).",
					Type:        types.PinTypes.Object, // Expect Object primarily
				},
			},
			// Updated Output Pins
			Outputs: []types.Pin{
				{
					ID:          "then", // Standard success execution pin ID
					Name:        "Then",
					Description: "Activated upon successful transformation.",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "value", // Output data pin ID
					Name:        "Value",
					Description: "Transformed JSON data.",
					Type:        types.PinTypes.Object, // Output matches input type
				},
				{
					ID:          "onError", // Standard error execution pin ID
					Name:        "OnError",
					Description: "Activated if an error occurs during processing.",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "errorMessage", // Standard error message pin ID
					Name:        "Error Message",
					Description: "Details about any error encountered.",
					Type:        types.PinTypes.String,
				},
			},
			// Updated Properties
			Properties: []types.Property{
				{
					Name:         "schemaComponentId", // Updated property name
					DisplayName:  "Schema Component",
					Description:  "Select the Schema Definition/Component to apply.",
					Type:         types.PinTypes.String,
					Value:        "",
					DefaultValue: "",
					// Options should ideally be populated dynamically.
				},
			},
		},
	}
}

// Execute performs the schema transformation.
// (Renamed receiver type)
func (n *SchemaNode) Execute(ctx node.ExecutionContext) error {
	var _ db.SchemaComponentStore // Dummy variable to ensure db import is used
	logger := ctx.Logger()

	// --- 1. Get Schema Component ID from Property ---
	var schemaID string
	var foundProp bool
	// Use updated property name
	for _, prop := range n.BaseNode.GetProperties() {
		if prop.Name == "schemaComponentId" {
			if idVal, ok := prop.Value.(string); ok {
				schemaID = idVal
				foundProp = true
			}
			break
		}
	}

	// Use updated property name in error message
	if !foundProp || schemaID == "" {
		errMsg := "Schema Component property ('schemaComponentId') is not set or invalid"
		logger.Error(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}

	// --- 2. Get Schema Component Store from Context ---
	// Use the standard SchemaAccessContext interface from node package
	schemaCtx, ok := engineext.GetExtendedContext(ctx).(node.SchemaAccessContext)
	if !ok {
		errMsg := "Execution context does not support schema component access"
		logger.Error(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}
	schemaStore := schemaCtx.GetSchemaComponentStore()
	if schemaStore == nil {
		errMsg := "Schema Component Store is not available in the execution context"
		logger.Error(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}

	// --- 3. Fetch Schema Definition from DB ---
	schemaComponent, err := schemaStore.GetSchemaComponent(schemaID)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to fetch schema component %s: %v", schemaID, err)
		logger.Error(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}

	// --- 4. Compile the Schema (assuming JSON Schema validation for now) ---
	// TODO: Extend this section based on how transformation rules are defined.
	// For now, we only validate.
	compiler := jsonschema.NewCompiler()
	err = compiler.AddResource("schema.json", strings.NewReader(schemaComponent.SchemaDefinition))
	if err != nil {
		errMsg := fmt.Sprintf("Failed to add schema resource for %s: %v", schemaID, err)
		logger.Error(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to compile schema %s: %v", schemaID, err)
		logger.Error(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}

	// --- 5. Get Input Data (use updated pin ID 'value') ---
	inputValue, exists := ctx.GetInputValue("value")
	if !exists {
		errMsg := "Input value 'value' is missing" // Updated pin name in message
		logger.Warn(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}

	// --- 6. Prepare Input Data for Validation/Transformation ---
	var dataToProcess interface{}
	// inputIsString variable is no longer needed
	// Attempt to unmarshal if input is a JSON string
	if inputValue.Type.ID == types.PinTypes.String.ID {
		// inputIsString = true // Removed usage of inputIsString
		strVal, _ := inputValue.AsString()
		if strings.TrimSpace(strVal) == "" {
			// Handle empty string input - depends on schema requirements
			// Maybe treat as null or empty object? For now, try unmarshalling as null.
			dataToProcess = nil
		} else {
			err := json.Unmarshal([]byte(strVal), &dataToProcess)
			if err != nil {
				// If it's not valid JSON, validation will likely fail anyway.
				// Log the error but proceed with the raw string for validation attempt.
				logger.Warn("Input data is string but not valid JSON, attempting validation on raw string", map[string]interface{}{"error": err.Error()})
				// Depending on schema, maybe error out here? For now, let schema validator decide.
				dataToProcess = strVal // Fallback to raw string
			}
		}
	} else {
		// Assume it's already a Go map/slice/primitive from an Object/Array/Number/Boolean pin
		dataToProcess = inputValue.RawValue
	}

	// --- 7. Validate Data using JSON Schema ---
	validationErr := schema.Validate(dataToProcess)
	if validationErr != nil {
		errMsg := fmt.Sprintf("Schema validation failed: %v", validationErr)
		logger.Warn(errMsg, nil)
		// Use updated error pins
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, errMsg))
		return ctx.ActivateOutputFlow("onError")
	}

	// --- 8. Transformation ---
	// Placeholder: The actual transformation logic based on schemaComponent.SchemaDefinition goes here.
	// This depends on the chosen transformation language/rules (e.g., JSONata, JMESPath).
	// For demonstration, we'll use the validated data and add some metadata.
	transformedData := dataToProcess // Start with the validated data

	// --- 9. Prepare Output Object ---
	// Ensure the output pin 'value' (type Object) receives a map[string]interface{}.
	var outputObject map[string]interface{}
	if obj, ok := transformedData.(map[string]interface{}); ok {
		// If the validated data is already a map, use it as the base.
		outputObject = obj
	} else {
		// If the validated data is not a map, wrap it.
		logger.Debug("Validated data is not a JSON object, wrapping output.", map[string]interface{}{"dataType": fmt.Sprintf("%T", transformedData)})
		outputObject = map[string]interface{}{
			"result": transformedData,
		}
	}

	// Add placeholder transformation info
	outputObject["_transformation_info"] = map[string]interface{}{
		"schema_id":      schemaID,
		"transformed_at": time.Now().UTC().Format(time.RFC3339Nano),
		"status":         "ValidatedOnly", // Indicate only validation occurred
	}

	// --- 9. Set Output & Activate Flow (use updated pin IDs) ---
	logger.Info("Schema validation/transformation successful", nil)
	ctx.SetOutputValue("value", types.NewValue(types.PinTypes.Object, outputObject)) // Use 'value' output pin
	ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, ""))    // Clear error message on success

	return ctx.ActivateOutputFlow("then") // Use 'then' output pin
}

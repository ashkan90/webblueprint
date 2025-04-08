package math

import (
	"fmt"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// SafeDivideNode implements a division node with error recovery
type SafeDivideNode struct {
	inputs     []types.Pin
	outputs    []types.Pin
	properties []types.Property
}

// NewSafeDivideNode creates a new safe divide node
func NewSafeDivideNode() node.Node {
	return &SafeDivideNode{
		inputs: []types.Pin{
			{
				ID:          "exec",
				Name:        "Execute",
				Description: "Execution input",
				Type:        types.PinTypes.Execution,
			},
			{
				ID:          "dividend",
				Name:        "Dividend",
				Description: "The number to divide (numerator)",
				Type:        types.PinTypes.Number,
			},
			{
				ID:          "divisor",
				Name:        "Divisor",
				Description: "The number to divide by (denominator)",
				Type:        types.PinTypes.Number,
			},
			{
				ID:          "default",
				Name:        "Default Value",
				Description: "Value to return if division by zero occurs",
				Type:        types.PinTypes.Number,
			},
		},
		outputs: []types.Pin{
			{
				ID:          "then",
				Name:        "Then",
				Description: "Execution continues when operation succeeds",
				Type:        types.PinTypes.Execution,
			},
			{
				ID:          "catch",
				Name:        "Catch",
				Description: "Execution continues when error occurs",
				Type:        types.PinTypes.Execution,
			},
			{
				ID:          "result",
				Name:        "Result",
				Description: "The division result",
				Type:        types.PinTypes.Number,
			},
			{
				ID:          "error",
				Name:        "Error",
				Description: "Error information if division failed",
				Type:        types.PinTypes.Object,
			},
		},
		properties: []types.Property{
			{
				Name:  "defaultValue",
				Type:  types.PinTypes.Number,
				Value: 0,
			},
			{
				Name:  "errorHandlingMode",
				Type:  types.PinTypes.String,
				Value: "auto", // "auto", "manual", "default"
			},
		},
	}
}

// NewSafeDivideNodeFactory returns a factory function for creating safe divide nodes
func NewSafeDivideNodeFactory() node.NodeFactory {
	return func() node.Node {
		return NewSafeDivideNode()
	}
}

// GetMetadata returns node metadata
func (n *SafeDivideNode) GetMetadata() node.NodeMetadata {
	return node.NodeMetadata{
		TypeID:      "safe-divide",
		Name:        "Safe Divide",
		Description: "Divides two numbers with error handling",
		Category:    "Math",
		Version:     "1.0.0",
	}
}

// GetInputPins returns input pins
func (n *SafeDivideNode) GetInputPins() []types.Pin {
	return n.inputs
}

// SetInputPins sets the input pins
func (n *SafeDivideNode) SetInputPins(pins []types.Pin) {
	n.inputs = pins
}

// GetOutputPins returns output pins
func (n *SafeDivideNode) GetOutputPins() []types.Pin {
	return n.outputs
}

// SetOutputPins sets the output pins
func (n *SafeDivideNode) SetOutputPins(pins []types.Pin) {
	n.outputs = pins
}

// GetProperties returns node properties
func (n *SafeDivideNode) GetProperties() []types.Property {
	return n.properties
}

func (n *SafeDivideNode) SetProperty(name string, value interface{}) {
	for i := range n.properties {
		if n.properties[i].Name == name {
			n.properties[i].Value = value
		}
	}
}

// Execute runs the node
func (n *SafeDivideNode) Execute(ctx node.ExecutionContext) error {
	// Check if context supports error handling
	errorCtx, isErrorAware := ctx.(bperrors.ErrorAwareContext)

	// Get input values
	dividendValue, dividendExists := ctx.GetInputValue("dividend")
	divisorValue, divisorExists := ctx.GetInputValue("divisor")
	defaultValue, defaultExists := ctx.GetInputValue("default")

	// Check for missing inputs
	if !dividendExists {
		// Handle missing dividend
		if isErrorAware {
			err := errorCtx.ReportError(
				bperrors.ErrorTypeValidation,
				bperrors.ErrMissingRequiredInput,
				"Missing required dividend input",
				nil,
			)

			// Try to recover with default
			recovered, _ := errorCtx.AttemptRecovery(err)
			if recovered {
				dividendValue = types.NewValue(types.PinTypes.Number, 0)
			} else {
				ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
					"message": "Missing dividend input",
					"code":    string(bperrors.ErrMissingRequiredInput),
				}))
				return ctx.ActivateOutputFlow("catch")
			}
		} else {
			return fmt.Errorf("missing required dividend input")
		}
	}

	if !divisorExists {
		// Handle missing divisor
		if isErrorAware {
			err := errorCtx.ReportError(
				bperrors.ErrorTypeValidation,
				bperrors.ErrMissingRequiredInput,
				"Missing required divisor input",
				nil,
			)

			// Try to recover with default
			recovered, _ := errorCtx.AttemptRecovery(err)
			if recovered {
				divisorValue = types.NewValue(types.PinTypes.Number, 1)
			} else {
				ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
					"message": "Missing divisor input",
					"code":    string(bperrors.ErrMissingRequiredInput),
				}))
				return ctx.ActivateOutputFlow("catch")
			}
		} else {
			return fmt.Errorf("missing required divisor input")
		}
	}

	// Convert inputs to numbers
	dividend, err := dividendValue.AsNumber()
	if err != nil {
		if isErrorAware {
			apiErr := errorCtx.ReportError(
				bperrors.ErrorTypeValidation,
				bperrors.ErrTypeMismatch,
				"Failed to convert dividend to number",
				err,
			)
			ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"message": apiErr.Message,
				"code":    string(apiErr.Code),
			}))
		}
		return ctx.ActivateOutputFlow("catch")
	}

	divisor, err := divisorValue.AsNumber()
	if err != nil {
		if isErrorAware {
			apiErr := errorCtx.ReportError(
				bperrors.ErrorTypeValidation,
				bperrors.ErrTypeMismatch,
				"Failed to convert divisor to number",
				err,
			)
			ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"message": apiErr.Message,
				"code":    string(apiErr.Code),
			}))
		}
		return ctx.ActivateOutputFlow("catch")
	}

	// Check for division by zero
	if divisor == 0 {
		// Get the error handling mode from properties
		errorHandlingMode := "auto"
		for _, prop := range n.GetProperties() {
			if prop.Name == "errorHandlingMode" {
				if mode, ok := prop.Value.(string); ok {
					errorHandlingMode = mode
				}
			}
		}

		// Handle division by zero based on mode
		if errorHandlingMode == "default" || (errorHandlingMode == "auto" && defaultExists) {
			// Use default value
			var resultValue float64
			if defaultExists {
				defaultVal, err := defaultValue.AsNumber()
				if err != nil {
					defaultVal = 0 // Fallback if default is invalid
				}
				resultValue = defaultVal
			} else {
				// Get default from properties
				resultValue = 0
				for _, prop := range n.GetProperties() {
					if prop.Name == "defaultValue" {
						if val, ok := prop.Value.(float64); ok {
							resultValue = val
						}
					}
				}
			}

			// Log the error but use default
			if isErrorAware {
				errorCtx.ReportError(
					bperrors.ErrorTypeExecution,
					bperrors.ErrNodeExecutionFailed,
					"Division by zero, using default value",
					fmt.Errorf("division by zero"),
				)
			}

			ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, resultValue))
			return ctx.ActivateOutputFlow("then")
		} else {
			// Report error and activate catch flow
			if isErrorAware {
				apiErr := errorCtx.ReportError(
					bperrors.ErrorTypeExecution,
					bperrors.ErrNodeExecutionFailed,
					"Division by zero",
					fmt.Errorf("division by zero"),
				)
				ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
					"message": apiErr.Message,
					"code":    string(apiErr.Code),
					"details": map[string]interface{}{
						"dividend": dividend,
						"divisor":  divisor,
					},
				}))
			} else {
				ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
					"message": "Division by zero",
				}))
			}

			return ctx.ActivateOutputFlow("catch")
		}
	}

	// Perform division
	result := dividend / divisor
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, result))

	// Execute success path
	return ctx.ActivateOutputFlow("then")
}

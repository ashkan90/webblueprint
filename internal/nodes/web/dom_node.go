package web

import (
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// DOMElementNode implements a node that creates or modifies a DOM element
type DOMElementNode struct {
	node.BaseNode
}

// NewDOMElementNode creates a new DOM Element node
func NewDOMElementNode() node.Node {
	return &DOMElementNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "dom-element",
				Name:        "DOM Element",
				Description: "Creates or modifies a DOM element",
				Category:    "Web",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "exec",
					Name:        "Execute",
					Description: "Execution input",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "selector",
					Name:        "Selector",
					Description: "CSS selector to find element(s). If empty, creates a new element",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "tagName",
					Name:        "Tag Name",
					Description: "HTML tag name for new element (e.g., 'div', 'span')",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     "div",
				},
				{
					ID:          "innerHTML",
					Name:        "Inner HTML",
					Description: "HTML content to set inside the element",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "textContent",
					Name:        "Text Content",
					Description: "Text content to set inside the element",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "attributes",
					Name:        "Attributes",
					Description: "Object with element attributes to set",
					Type:        types.PinTypes.Object,
					Optional:    true,
				},
				{
					ID:          "styles",
					Name:        "Styles",
					Description: "Object with CSS styles to apply",
					Type:        types.PinTypes.Object,
					Optional:    true,
				},
				{
					ID:          "parentSelector",
					Name:        "Parent Selector",
					Description: "CSS selector for parent element to append to",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     "body",
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Execution continues",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "element",
					Name:        "Element",
					Description: "The created or modified DOM element",
					Type:        types.PinTypes.Object,
				},
				{
					ID:          "success",
					Name:        "Success",
					Description: "Whether the operation was successful",
					Type:        types.PinTypes.Boolean,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *DOMElementNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing DOM Element node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	selectorValue, selectorExists := ctx.GetInputValue("selector")
	tagNameValue, tagNameExists := ctx.GetInputValue("tagName")
	innerHTMLValue, innerHTMLExists := ctx.GetInputValue("innerHTML")
	textContentValue, textContentExists := ctx.GetInputValue("textContent")
	attributesValue, attributesExists := ctx.GetInputValue("attributes")
	stylesValue, stylesExists := ctx.GetInputValue("styles")
	parentSelectorValue, parentSelectorExists := ctx.GetInputValue("parentSelector")

	// Default values if not provided
	tagName := "div"
	if tagNameExists {
		tagNameStr, err := tagNameValue.AsString()
		if err == nil {
			tagName = tagNameStr
		}
	}

	parentSelector := "body"
	if parentSelectorExists {
		parentSelectorStr, err := parentSelectorValue.AsString()
		if err == nil && parentSelectorStr != "" {
			parentSelector = parentSelectorStr
		}
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"hasSelector":       selectorExists,
		"hasTagName":        tagNameExists,
		"hasInnerHTML":      innerHTMLExists,
		"hasTextContent":    textContentExists,
		"hasAttributes":     attributesExists,
		"hasStyles":         stylesExists,
		"hasParentSelector": parentSelectorExists,
	}

	// Build a JavaScript object to represent the DOM operation
	jsOperation := make(map[string]interface{})

	if selectorExists {
		selector, _ := selectorValue.AsString()
		jsOperation["selector"] = selector
		jsOperation["mode"] = "modify"
	} else {
		jsOperation["tagName"] = tagName
		jsOperation["mode"] = "create"
		jsOperation["parentSelector"] = parentSelector
	}

	if innerHTMLExists {
		innerHTML, _ := innerHTMLValue.AsString()
		jsOperation["innerHTML"] = innerHTML
	}

	if textContentExists {
		textContent, _ := textContentValue.AsString()
		jsOperation["textContent"] = textContent
	}

	if attributesExists {
		attributes, _ := attributesValue.AsObject()
		jsOperation["attributes"] = attributes
	}

	if stylesExists {
		styles, _ := stylesValue.AsObject()
		jsOperation["styles"] = styles
	}

	// Create the output element object that will be sent to the client
	elementObj := make(map[string]interface{})
	elementObj["operation"] = jsOperation
	elementObj["timestamp"] = time.Now().UnixNano()
	elementObj["nodeId"] = ctx.GetNodeID()

	// Set outputs
	ctx.SetOutputValue("element", types.NewValue(types.PinTypes.Object, elementObj))
	ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, true))

	debugData["output"] = elementObj

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "DOM Element Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("DOM element operation prepared", map[string]interface{}{
		"mode":     jsOperation["mode"],
		"selector": jsOperation["selector"],
		"tagName":  jsOperation["tagName"],
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

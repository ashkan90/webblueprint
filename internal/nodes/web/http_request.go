package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// HTTPRequestNode implements an HTTP request node
type HTTPRequestNode struct {
	node.BaseNode
}

// NewHTTPRequestNode creates a new HTTP request node
func NewHTTPRequestNode() node.Node {
	return &HTTPRequestNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "http-request",
				Name:        "HTTP Request",
				Description: "Makes an HTTP request to a specified URL",
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
					ID:          "url",
					Name:        "URL",
					Description: "URL to send the request to",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "method",
					Name:        "Method",
					Description: "HTTP method (GET, POST, etc.)",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     "GET",
				},
				{
					ID:          "headers",
					Name:        "Headers",
					Description: "HTTP headers to include",
					Type:        types.PinTypes.Object,
					Optional:    true,
				},
				{
					ID:          "body",
					Name:        "Body",
					Description: "Request body (for POST, PUT, etc.)",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Executed on successful request",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "catch",
					Name:        "Catch",
					Description: "Executed if an error occurs",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "response",
					Name:        "Response",
					Description: "Response data",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "status",
					Name:        "Status Code",
					Description: "HTTP status code",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *HTTPRequestNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing HTTP Request node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get URL and method
	urlValue, urlExists := ctx.GetInputValue("url")
	methodValue, methodExists := ctx.GetInputValue("method")
	headersValue, headersExist := ctx.GetInputValue("headers")
	bodyValue, bodyExists := ctx.GetInputValue("body")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"url":     urlExists,
		"method":  methodExists,
		"headers": headersExist,
		"body":    bodyExists,
	}

	// Validate URL
	if !urlExists {
		err := fmt.Errorf("missing required input: url")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing URL",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("catch")
	}

	url, err := urlValue.AsString()
	if err != nil {
		logger.Error("Invalid URL", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_url",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid URL",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("catch")
	}

	// Get method (default to GET)
	method := "GET"
	if methodExists {
		methodStr, err := methodValue.AsString()
		if err == nil {
			method = methodStr
		}
	}

	debugData["request"] = map[string]interface{}{
		"url":    url,
		"method": method,
	}

	logger.Info("Preparing HTTP request", map[string]interface{}{
		"url":    url,
		"method": method,
	})

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Prepare request
	var req *http.Request

	if bodyExists && bodyValue.RawValue != nil {
		// Convert body to JSON if it's not already a string
		var bodyContent []byte

		if bodyStr, ok := bodyValue.RawValue.(string); ok {
			bodyContent = []byte(bodyStr)
		} else {
			// Try to marshal as JSON
			jsonData, err := json.Marshal(bodyValue.RawValue)
			if err != nil {
				logger.Error("Failed to marshal request body", map[string]interface{}{"error": err.Error()})

				debugData["error"] = map[string]string{
					"type":    "body_marshal",
					"message": "Failed to marshal request body",
					"details": err.Error(),
				}
				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "Error: Invalid body format",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
				return ctx.ActivateOutputFlow("catch")
			}
			bodyContent = jsonData
		}

		debugData["requestBody"] = string(bodyContent)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(bodyContent))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		logger.Error("Failed to create HTTP request", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "request_creation",
			"message": "Failed to create HTTP request",
			"details": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Failed to create request",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("catch")
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WebBlueprint/2.0")

	// Add custom headers if provided
	if headersExist {
		headers, err := headersValue.AsObject()
		if err == nil {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					req.Header.Set(key, strValue)
				}
			}
		}
	}

	// Record request headers for debugging
	headerMap := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	debugData["requestHeaders"] = headerMap

	// Record debug snapshot before making the request
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Making HTTP Request",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Execute the request
	startTime := time.Now()
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)

	// Add timing information
	debugData["timing"] = map[string]interface{}{
		"start":    startTime.Format(time.RFC3339),
		"duration": requestDuration.String(),
	}

	if err != nil {
		logger.Error("HTTP request failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "request_execution",
			"message": "HTTP request failed",
			"details": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Request failed",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		ctx.SetOutputValue("status", types.NewValue(types.PinTypes.Number, float64(0)))
		return ctx.ActivateOutputFlow("catch")
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "response_reading",
			"message": "Failed to read response body",
			"details": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Failed to read response",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		ctx.SetOutputValue("status", types.NewValue(types.PinTypes.Number, float64(resp.StatusCode)))
		return ctx.ActivateOutputFlow("catch")
	}

	// Try to parse as JSON first
	var responseData interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		// Not valid JSON, return as string
		responseData = string(responseBody)
		debugData["responseFormat"] = "string"
	} else {
		debugData["responseFormat"] = "json"
	}

	// Record response information
	debugData["response"] = map[string]interface{}{
		"statusCode": resp.StatusCode,
		"headers":    resp.Header,
		"size":       len(responseBody),
	}

	// Set output values
	ctx.SetOutputValue("response", types.NewValue(types.PinTypes.Any, responseData))
	ctx.SetOutputValue("status", types.NewValue(types.PinTypes.Number, float64(resp.StatusCode)))

	// Final debug snapshot
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "HTTP Request Completed",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("HTTP request completed", map[string]interface{}{
		"statusCode": resp.StatusCode,
		"size":       len(responseBody),
		"body":       string(responseBody),
	})

	return ctx.ActivateOutputFlow("then")
}

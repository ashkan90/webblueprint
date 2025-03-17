package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// HTTPRequestWithRecoveryNode is an HTTP request node with enhanced error handling
type HTTPRequestWithRecoveryNode struct {
}

// NewHTTPRequestWithRecoveryNode creates a new HTTP request node with error handling
func NewHTTPRequestWithRecoveryNode() node.Node {
	return &HTTPRequestWithRecoveryNode{}
}

// NewHttpRequestWithRecoveryNodeFactory returns a factory function for creating HTTP request with recovery nodes
func NewHttpRequestWithRecoveryNodeFactory() node.NodeFactory {
	return func() node.Node {
		return NewHTTPRequestWithRecoveryNode()
	}
}

// GetMetadata returns node metadata
func (n *HTTPRequestWithRecoveryNode) GetMetadata() node.NodeMetadata {
	return node.NodeMetadata{
		TypeID:      "http-request-with-recovery",
		Name:        "HTTP Request (With Recovery)",
		Description: "Makes an HTTP request with enhanced error handling",
		Category:    "Web",
		Version:     "1.0.0",
	}
}

// GetInputPins returns input pins
func (n *HTTPRequestWithRecoveryNode) GetInputPins() []types.Pin {
	return []types.Pin{
		{
			ID:          "exec",
			Name:        "Execute",
			Description: "Execution input",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "url",
			Name:        "URL",
			Description: "The URL to request",
			Type:        types.PinTypes.String,
		},
		{
			ID:          "method",
			Name:        "Method",
			Description: "HTTP method (GET, POST, etc.)",
			Type:        types.PinTypes.String,
		},
		{
			ID:          "headers",
			Name:        "Headers",
			Description: "HTTP headers",
			Type:        types.PinTypes.Object,
		},
		{
			ID:          "body",
			Name:        "Body",
			Description: "Request body",
			Type:        types.PinTypes.String,
		},
		{
			ID:          "timeout",
			Name:        "Timeout (ms)",
			Description: "Request timeout in milliseconds",
			Type:        types.PinTypes.Number,
		},
	}
}

// GetOutputPins returns output pins
func (n *HTTPRequestWithRecoveryNode) GetOutputPins() []types.Pin {
	return []types.Pin{
		{
			ID:          "then",
			Name:        "Then",
			Description: "Execution output when successful",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "catch",
			Name:        "Catch",
			Description: "Execution output when error occurs",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "response",
			Name:        "Response",
			Description: "HTTP response",
			Type:        types.PinTypes.Object,
		},
		{
			ID:          "statusCode",
			Name:        "Status Code",
			Description: "HTTP status code",
			Type:        types.PinTypes.Number,
		},
		{
			ID:          "body",
			Name:        "Response Body",
			Description: "Response body as string",
			Type:        types.PinTypes.String,
		},
		{
			ID:          "headers",
			Name:        "Response Headers",
			Description: "Response headers",
			Type:        types.PinTypes.Object,
		},
		{
			ID:          "error",
			Name:        "Error",
			Description: "Error object if request failed",
			Type:        types.PinTypes.Object,
		},
	}
}

// GetProperties returns node properties
func (n *HTTPRequestWithRecoveryNode) GetProperties() []types.Property {
	return []types.Property{
		{
			Name:        "defaultMethod",
			DisplayName: "Default Method",
			Type:        types.PinTypes.String,
			Value:       "GET",
		},
		{
			Name:        "defaultTimeout",
			DisplayName: "Default Timeout",
			Type:        types.PinTypes.Number,
			Value:       5000,
		},
		{
			Name:        "retryCount",
			DisplayName: "Retry Count",
			Type:        types.PinTypes.Number,
			Value:       3,
		},
		{
			Name:        "fallbackUrl",
			DisplayName: "Fallback URL",
			Type:        types.PinTypes.String,
			Value:       "",
		},
	}
}

// Execute runs the node
func (n *HTTPRequestWithRecoveryNode) Execute(ctx node.ExecutionContext) error {
	// Check if we have error-aware context
	errorAware, isErrorAware := ctx.(bperrors.ErrorAwareContext)

	// Get inputs with recovery for missing values
	urlValue, urlExists := ctx.GetInputValue("url")

	// If URL doesn't exist and we have error-aware context, handle it
	if !urlExists && isErrorAware {
		err := errorAware.ReportError(
			bperrors.ErrorTypeValidation,
			bperrors.ErrMissingRequiredInput,
			"Missing required URL input",
			nil,
		)

		// Try to recover
		success, _ := errorAware.AttemptRecovery(err)
		if !success {
			// If recovery failed, activate catch flow and return error info
			ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"message": "Missing URL and recovery failed",
				"code":    string(bperrors.ErrMissingRequiredInput),
			}))
			ctx.ActivateOutputFlow("catch")
			return err
		}

		// If recovery succeeded, check for fallback URL in properties
		var fallbackUrl string
		for _, prop := range n.GetProperties() {
			if prop.Name == "fallbackUrl" {
				if strValue, ok := prop.Value.(string); ok && strValue != "" {
					fallbackUrl = strValue
				}
			}
		}

		if fallbackUrl != "" {
			urlValue = types.NewValue(types.PinTypes.String, fallbackUrl)
			ctx.Logger().Info("Using fallback URL", map[string]interface{}{
				"fallbackUrl": fallbackUrl,
			})
		} else {
			// No fallback, create empty response
			ctx.SetOutputValue("response", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"recovered": true,
				"message":   "Request skipped due to missing URL",
			}))
			ctx.SetOutputValue("statusCode", types.NewValue(types.PinTypes.Number, 0))
			ctx.SetOutputValue("body", types.NewValue(types.PinTypes.String, ""))
			ctx.SetOutputValue("headers", types.NewValue(types.PinTypes.Object, map[string]interface{}{}))

			return ctx.ActivateOutputFlow("then")
		}
	} else if !urlExists {
		// Standard error handling without recovery if not error-aware
		return fmt.Errorf("missing required URL input")
	}

	// Get other inputs with fallbacks to properties
	method, methodExists := ctx.GetInputValue("method")
	if !methodExists {
		// Default to GET or property value
		methodStr := "GET"
		for _, prop := range n.GetProperties() {
			if prop.Name == "defaultMethod" {
				if strValue, ok := prop.Value.(string); ok {
					methodStr = strValue
				}
			}
		}
		method = types.NewValue(types.PinTypes.String, methodStr)
	}

	timeout, timeoutExists := ctx.GetInputValue("timeout")
	if !timeoutExists {
		// Default to 5000 or property value
		timeoutVal := 5000.0
		for _, prop := range n.GetProperties() {
			if prop.Name == "defaultTimeout" {
				if numValue, ok := prop.Value.(float64); ok {
					timeoutVal = numValue
				} else if intValue, ok := prop.Value.(int); ok {
					timeoutVal = float64(intValue)
				}
			}
		}
		timeout = types.NewValue(types.PinTypes.Number, timeoutVal)
	}

	// Parse timeout as milliseconds
	timeoutVal, err := timeout.AsNumber()
	if err != nil {
		ctx.Logger().Error("Invalid Timeout", map[string]interface{}{"error": err.Error()})

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid Timeout",
			Value: map[string]string{
				"type":    "invalid_timeout",
				"message": err.Error(),
			},
			Timestamp: time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("catch")
	}
	timeoutDuration := time.Duration(timeoutVal) * time.Millisecond

	urlVal, err := urlValue.AsString()
	if err != nil {
		ctx.Logger().Error("Invalid URL", map[string]interface{}{"error": err.Error()})

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid URL",
			Value: map[string]string{
				"type":    "invalid_url",
				"message": err.Error(),
			},
			Timestamp: time.Now(),
		})

		ctx.SetOutputValue("response", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("catch")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeoutDuration,
	}

	// Create request
	methodVal := "GET"
	if methodExists {
		methodStr, err := method.AsString()
		if err == nil {
			methodVal = methodStr
		}
	}

	req, err := http.NewRequest(methodVal, urlVal, nil)
	if err != nil {
		// Handle request creation error
		if isErrorAware {
			apiErr := errorAware.ReportError(
				bperrors.ErrorTypeExecution,
				bperrors.ErrNodeExecutionFailed,
				"Failed to create HTTP request",
				err,
			)

			// Set error output
			ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"message": apiErr.Message,
				"code":    string(apiErr.Code),
			}))
			ctx.ActivateOutputFlow("catch")
			return apiErr
		}

		return err
	}

	// Add headers if provided
	headers, headersExist := ctx.GetInputValue("headers")
	if headersExist && headers.Type == types.PinTypes.Object {
		headerMap, ok := headers.RawValue.(map[string]interface{})
		if ok {
			for key, value := range headerMap {
				valueStr, ok := value.(string)
				if ok {
					req.Header.Add(key, valueStr)
				}
			}
		}
	}

	// Add body if provided
	body, bodyExists := ctx.GetInputValue("body")
	if bodyExists && body.Type == types.PinTypes.String {
		// In a real implementation, you would set the body on the request
	}

	// Execute request with retry logic
	var resp *http.Response
	maxRetries := 3

	// Get retry count from properties
	for _, prop := range n.GetProperties() {
		if prop.Name == "retryCount" {
			if numValue, ok := prop.Value.(float64); ok {
				maxRetries = int(numValue)
			} else if intValue, ok := prop.Value.(int); ok {
				maxRetries = intValue
			}
		}
	}

	retryCount := 0

	for retryCount <= maxRetries {
		resp, err = client.Do(req)
		if err == nil {
			break
		}

		retryCount++

		// If we have an error-aware context, log the retry
		if isErrorAware && retryCount <= maxRetries {
			ctx.Logger().Warn("HTTP request failed, retrying", map[string]interface{}{
				"url":        urlVal,
				"attempt":    retryCount,
				"maxRetries": maxRetries,
				"error":      err.Error(),
			})

			// Add a small delay between retries
			time.Sleep(time.Duration(retryCount*500) * time.Millisecond)
		}
	}

	// If all retries failed
	if err != nil {
		if isErrorAware {
			apiErr := errorAware.ReportError(
				bperrors.ErrorTypeConnection,
				bperrors.ErrNodeExecutionFailed,
				fmt.Sprintf("HTTP request failed after %d retries", retryCount),
				err,
			)

			// Set error output
			ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"message": apiErr.Message,
				"code":    string(apiErr.Code),
				"details": apiErr.Details,
				"retries": retryCount,
			}))
			ctx.ActivateOutputFlow("catch")
			return apiErr
		}

		return err
	}

	// Process response
	defer resp.Body.Close()

	// Read response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if isErrorAware {
			apiErr := errorAware.ReportError(
				bperrors.ErrorTypeExecution,
				bperrors.ErrNodeExecutionFailed,
				"Failed to read response body",
				err,
			)
			ctx.SetOutputValue("error", types.NewValue(types.PinTypes.Object, map[string]interface{}{
				"message": apiErr.Message,
				"code":    string(apiErr.Code),
			}))
			ctx.ActivateOutputFlow("catch")
			return apiErr
		}

		return err
	}

	// Convert headers to map
	headerMap := make(map[string]interface{})
	for key, values := range resp.Header {
		if len(values) == 1 {
			headerMap[key] = values[0]
		} else {
			headerMap[key] = values
		}
	}

	// Set outputs
	ctx.SetOutputValue("response", types.NewValue(types.PinTypes.Object, map[string]interface{}{
		"statusCode": resp.StatusCode,
		"headers":    headerMap,
		"body":       string(responseBody),
	}))

	ctx.SetOutputValue("statusCode", types.NewValue(types.PinTypes.Number, float64(resp.StatusCode)))
	ctx.SetOutputValue("body", types.NewValue(types.PinTypes.String, string(responseBody)))
	ctx.SetOutputValue("headers", types.NewValue(types.PinTypes.Object, headerMap))

	// Activate success flow
	ctx.ActivateOutputFlow("then")

	return nil
}

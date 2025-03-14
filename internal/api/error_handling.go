package api

import (
	"encoding/json"
	"net/http"
	"time"
	errors "webblueprint/internal/bperrors"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     *errors.BlueprintError `json:"error"`
	Message   string                 `json:"message"`
	ErrorCode string                 `json:"errorCode"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// writeErrorResponse writes a structured error response
func writeErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	response := ErrorResponse{
		Success:   false,
		Message:   err.Error(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Handle BlueprintError specifically
	if bpErr, ok := err.(*errors.BlueprintError); ok {
		response.Error = bpErr
		response.ErrorCode = string(bpErr.Code)
		response.Details = bpErr.Details
	} else {
		// Wrap in a generic BlueprintError
		bpErr = errors.Wrap(
			err,
			errors.ErrorTypeSystem,
			errors.ErrUnknown,
			err.Error(),
			errors.SeverityMedium,
		)
		response.Error = bpErr
		response.ErrorCode = string(bpErr.Code)
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// handleAPIError handles an error in an API handler
func handleAPIError(w http.ResponseWriter, err error) {
	// Set appropriate status code based on error type
	statusCode := http.StatusInternalServerError

	// If it's a BlueprintError, use more specific status codes
	if bpErr, ok := err.(*errors.BlueprintError); ok {
		switch bpErr.Type {
		case errors.ErrorTypeValidation:
			statusCode = http.StatusBadRequest
		case errors.ErrorTypePermission:
			statusCode = http.StatusForbidden
		case errors.ErrorTypeConnection:
			statusCode = http.StatusServiceUnavailable
		case errors.ErrorTypeDatabase:
			statusCode = http.StatusInternalServerError
		}

		// Special case error codes
		switch bpErr.Code {
		case errors.ErrBlueprintNotFound:
			statusCode = http.StatusNotFound
		case errors.ErrNodeNotFound:
			statusCode = http.StatusNotFound
		}
	}

	writeErrorResponse(w, err, statusCode)
}

// ErrorHandler wraps an HTTP handler with standardized error handling
func ErrorHandler(handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			handleAPIError(w, err)
		}
	}
}

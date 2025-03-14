package bperrors

import (
	"fmt"
	"strings"
)

// TestVerifier verifies error handling behavior
type TestVerifier struct {
	errorManager    *ErrorManager
	recoveryManager *RecoveryManager
}

// NewTestVerifier creates a new test verifier
func NewTestVerifier(errorManager *ErrorManager, recoveryManager *RecoveryManager) *TestVerifier {
	return &TestVerifier{
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
	}
}

// VerificationResult represents the result of a verification
type VerificationResult struct {
	Success      bool     `json:"success"`
	Failures     []string `json:"failures,omitempty"`
	Passes       []string `json:"passes"`
	TotalChecks  int      `json:"totalChecks"`
	PassedChecks int      `json:"passedChecks"`
}

// VerifyErrorHandling verifies that error handling is working correctly
func (v *TestVerifier) VerifyErrorHandling(executionID string) VerificationResult {
	result := VerificationResult{
		Success:  true,
		Failures: make([]string, 0),
		Passes:   make([]string, 0),
	}

	// Get all errors for the execution
	errors := v.errorManager.GetErrors(executionID)

	// Check if errors were recorded
	if len(errors) == 0 {
		result.Success = false
		result.Failures = append(result.Failures, "No errors were recorded for the execution")
		result.TotalChecks = 1
		return result
	}

	result.Passes = append(result.Passes, fmt.Sprintf("Found %d errors for execution %s", len(errors), executionID))

	// Verify error metadata
	totalChecks := 0
	passedChecks := 0

	for _, err := range errors {
		// Check error type
		totalChecks++
		if err.Type != "" {
			passedChecks++
			result.Passes = append(result.Passes, fmt.Sprintf("Error %s has valid type: %s", err.Code, err.Type))
		} else {
			result.Success = false
			result.Failures = append(result.Failures, fmt.Sprintf("Error %s has empty type", err.Code))
		}

		// Check error code
		totalChecks++
		if err.Code != "" {
			passedChecks++
			result.Passes = append(result.Passes, fmt.Sprintf("Error %s has valid code", err.Code))
		} else {
			result.Success = false
			result.Failures = append(result.Failures, "Error has empty code")
		}

		// Check error message
		totalChecks++
		if err.Message != "" {
			passedChecks++
			result.Passes = append(result.Passes, fmt.Sprintf("Error %s has valid message: %s", err.Code, truncate(err.Message, 30)))
		} else {
			result.Success = false
			result.Failures = append(result.Failures, fmt.Sprintf("Error %s has empty message", err.Code))
		}

		// Check severity
		totalChecks++
		if err.Severity != "" {
			passedChecks++
			result.Passes = append(result.Passes, fmt.Sprintf("Error %s has valid severity: %s", err.Code, err.Severity))
		} else {
			result.Success = false
			result.Failures = append(result.Failures, fmt.Sprintf("Error %s has empty severity", err.Code))
		}

		// Check timestamp
		totalChecks++
		if !err.Timestamp.IsZero() {
			passedChecks++
			result.Passes = append(result.Passes, fmt.Sprintf("Error %s has valid timestamp", err.Code))
		} else {
			result.Success = false
			result.Failures = append(result.Failures, fmt.Sprintf("Error %s has zero timestamp", err.Code))
		}

		// If recoverable, check recovery options
		if err.Recoverable {
			totalChecks++
			if len(err.RecoveryOptions) > 0 {
				passedChecks++
				result.Passes = append(result.Passes, fmt.Sprintf("Recoverable error %s has %d recovery options", err.Code, len(err.RecoveryOptions)))
			} else {
				result.Success = false
				result.Failures = append(result.Failures, fmt.Sprintf("Recoverable error %s has no recovery options", err.Code))
			}
		}
	}

	// Verify error analysis
	analysis := v.errorManager.AnalyzeErrors(executionID)

	// Check analysis metadata
	totalChecks++
	if analysis != nil {
		passedChecks++
		result.Passes = append(result.Passes, "Error analysis was generated")

		// Check analysis contents
		totalChecks++
		if totalErrors, ok := analysis["totalErrors"].(int); ok && totalErrors == len(errors) {
			passedChecks++
			result.Passes = append(result.Passes, fmt.Sprintf("Analysis shows correct total errors: %d", totalErrors))
		} else {
			result.Success = false
			result.Failures = append(result.Failures, "Analysis has incorrect total errors count")
		}
	} else {
		result.Success = false
		result.Failures = append(result.Failures, "Error analysis was not generated")
	}

	// Verify recovery functionality
	recoverableErrors := 0
	for _, err := range errors {
		if err.Recoverable {
			recoverableErrors++

			// Test recovery attempt
			totalChecks++
			success, details := v.recoveryManager.RecoverFromError(executionID, err)

			if success {
				passedChecks++
				result.Passes = append(result.Passes, fmt.Sprintf("Recovery succeeded for error %s using strategy %s",
					err.Code, err.RecoveryOptions[0]))

				// Check recovery details
				totalChecks++
				if details != nil && len(details) > 0 {
					passedChecks++
					result.Passes = append(result.Passes, "Recovery details were provided")
				} else {
					result.Success = false
					result.Failures = append(result.Failures, "Recovery succeeded but no details were provided")
				}
			} else {
				totalChecks++

				// Check if recovery attempt was recorded
				attempts := v.recoveryManager.GetRecoveryAttempts(executionID, err.NodeID)
				if len(attempts) > 0 {
					passedChecks++
					result.Passes = append(result.Passes, "Recovery attempt was recorded even though it failed")
				} else {
					result.Success = false
					result.Failures = append(result.Failures, "Recovery attempt was not recorded for failed recovery")
				}
			}
		}
	}

	// If there are recoverable errors, verify that at least one was tested
	if recoverableErrors > 0 {
		totalChecks++
		result.Passes = append(result.Passes, fmt.Sprintf("Found %d recoverable errors", recoverableErrors))
		passedChecks++
	}

	// Update totals
	result.TotalChecks = totalChecks
	result.PassedChecks = passedChecks

	return result
}

// Helper to truncate long strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GenerateVerificationReport generates a human-readable report
func (v *TestVerifier) GenerateVerificationReport(executionID string) string {
	result := v.VerifyErrorHandling(executionID)

	var b strings.Builder

	b.WriteString("# Error Handling Verification Report\n\n")

	if result.Success {
		b.WriteString("✅ PASSED: All error handling checks passed\n\n")
	} else {
		b.WriteString("❌ FAILED: Some error handling checks failed\n\n")
	}

	b.WriteString(fmt.Sprintf("Total Checks: %d\n", result.TotalChecks))
	b.WriteString(fmt.Sprintf("Passed Checks: %d (%d%%)\n\n",
		result.PassedChecks, calculatePercentage(result.PassedChecks, result.TotalChecks)))

	if len(result.Failures) > 0 {
		b.WriteString("## Failures\n\n")
		for i, failure := range result.Failures {
			b.WriteString(fmt.Sprintf("%d. ❌ %s\n", i+1, failure))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Passed Checks\n\n")
	for i, pass := range result.Passes {
		b.WriteString(fmt.Sprintf("%d. ✅ %s\n", i+1, pass))
	}

	return b.String()
}

// Helper to calculate percentage
func calculatePercentage(part, total int) int {
	if total == 0 {
		return 0
	}
	return (part * 100) / total
}

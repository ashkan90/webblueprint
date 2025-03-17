package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	errors "webblueprint/internal/bperrors"
)

func main() {
	// Define flags
	scenarioType := flag.String("scenario", "execution_failure", "Type of error scenario to simulate")
	executionID := flag.String("execution", fmt.Sprintf("test-%d", time.Now().Unix()), "Execution ID to use")
	outputFile := flag.String("output", "", "File to write report to (optional)")
	flag.Parse()

	fmt.Printf("Running error handling test with scenario '%s' and execution ID '%s'\n\n", *scenarioType, *executionID)

	// Create test error generator
	generator := errors.NewTestErrorGenerator()

	// Generate errors
	fmt.Println("Generating test errors...")
	analysis, err := generator.SimulateErrorScenario(*scenarioType, *executionID)
	if err != nil {
		if bpErr, ok := err.(*errors.BlueprintError); ok {
			fmt.Printf("Generated main error: [%s-%s] %s\n", bpErr.Type, bpErr.Code, bpErr.Message)
		} else {
			fmt.Printf("Generated error: %s\n", err.Error())
		}
	}

	// Print analysis summary
	if analysis != nil {
		if totalErrors, ok := analysis["totalErrors"].(int); ok {
			fmt.Printf("Generated %d errors\n", totalErrors)
		}
		if recoverableErrors, ok := analysis["recoverableErrors"].(int); ok {
			fmt.Printf("Of which %d are recoverable\n", recoverableErrors)
		}
	}

	// Create test verifier
	fmt.Println("\nVerifying error handling...")
	verifier := errors.NewTestVerifier(generator.GetErrorManager(), generator.GetRecoveryManager())

	// Generate report
	report := verifier.GenerateVerificationReport(*executionID)

	// Print or save report
	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(report), 0644)
		if err != nil {
			fmt.Printf("Failed to write report to file: %s\n", err.Error())
			fmt.Println("\nReport:")
			fmt.Println(report)
		} else {
			fmt.Printf("Report written to %s\n", *outputFile)
		}
	} else {
		fmt.Println("\nReport:")
		fmt.Println(report)
	}
}

<template>
  <div class="error-testing-panel">
    <h2>Error Handling Testing Tool</h2>
    
    <div class="form-section">
      <h3>Generate Test Error</h3>
      
      <div class="form-group">
        <label>Execution ID:</label>
        <input v-model="executionId" placeholder="Enter execution ID or use default" />
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label>Error Type:</label>
          <select v-model="errorType">
            <option v-for="type in errorTypes" :key="type" :value="type">{{ type }}</option>
          </select>
        </div>
        
        <div class="form-group">
          <label>Severity:</label>
          <select v-model="severity">
            <option v-for="severity in severities" :key="severity" :value="severity">{{ severity }}</option>
          </select>
        </div>
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label>Error Code:</label>
          <select v-model="errorCode">
            <option v-for="code in errorCodes" :key="code.value" :value="code.value">
              {{ code.label }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label>Node ID:</label>
          <input v-model="nodeId" placeholder="Optional node ID" />
        </div>
      </div>
      
      <div class="form-group">
        <label>Message:</label>
        <input v-model="message" placeholder="Error message" />
      </div>
      
      <div class="form-group">
        <label>
          <input type="checkbox" v-model="recoverable" />
          Recoverable Error
        </label>
      </div>
      
      <button @click="generateError" class="generate-btn">Generate Error</button>
    </div>
    
    <div class="form-section">
      <h3>Generate Error Scenario</h3>
      
      <div class="form-group">
        <label>Scenario Type:</label>
        <select v-model="scenarioType">
          <option value="execution_failure">Execution Failure</option>
          <option value="connection_problem">Connection Problem</option>
          <option value="validation_error">Validation Error</option>
          <option value="database_error">Database Error</option>
          <option value="recoverable_errors">Recoverable Errors</option>
          <option value="multi_node_errors">Multi-Node Errors</option>
        </select>
      </div>
      
      <button @click="generateScenario" class="scenario-btn">Generate Scenario</button>
    </div>
    
    <div v-if="result" class="result-section">
      <h3>Result</h3>
      <div class="result-status" :class="{ success: result.success, error: !result.success }">
        {{ result.success ? 'Success' : 'Failed' }}
      </div>
      
      <div v-if="result.error" class="error-details">
        <h4>Error Details</h4>
        <pre>{{ JSON.stringify(result.error, null, 2) }}</pre>
      </div>
      
      <div v-if="result.analysis" class="analysis-details">
        <h4>Error Analysis</h4>
        <pre>{{ JSON.stringify(result.analysis, null, 2) }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue';
import {BlueprintErrorCode, ErrorSeverity, ErrorType} from '../../types/errors';

const executionId = ref(`test-${Date.now()}`);
const errorType = ref(ErrorType.Execution);
const errorCode = ref(BlueprintErrorCode.NodeExecutionFailed);
const severity = ref(ErrorSeverity.Medium);
const nodeId = ref('');
const message = ref('Test error message');
const recoverable = ref(true);
const scenarioType = ref('execution_failure');
const result = ref<any>(null);

// Error code options for dropdown
const errorCodes = [
  { value: BlueprintErrorCode.NodeExecutionFailed, label: 'E001 - Node Execution Failed' },
  { value: BlueprintErrorCode.NodeNotFound, label: 'E002 - Node Not Found' },
  { value: BlueprintErrorCode.NodeTypeNotRegistered, label: 'E003 - Node Type Not Registered' },
  { value: BlueprintErrorCode.ExecutionTimeout, label: 'E004 - Execution Timeout' },
  { value: BlueprintErrorCode.ExecutionCancelled, label: 'E005 - Execution Cancelled' },
  { value: BlueprintErrorCode.NoEntryPoints, label: 'E006 - No Entry Points' },
  { value: BlueprintErrorCode.InvalidConnection, label: 'C001 - Invalid Connection' },
  { value: BlueprintErrorCode.CircularDependency, label: 'C002 - Circular Dependency' },
  { value: BlueprintErrorCode.MissingRequiredInput, label: 'C003 - Missing Required Input' },
  { value: BlueprintErrorCode.TypeMismatch, label: 'C004 - Type Mismatch' },
  { value: BlueprintErrorCode.NodeDisconnected, label: 'C005 - Node Disconnected' },
  { value: BlueprintErrorCode.InvalidBlueprintStructure, label: 'V001 - Invalid Blueprint Structure' },
  { value: BlueprintErrorCode.InvalidNodeConfiguration, label: 'V002 - Invalid Node Configuration' },
  { value: BlueprintErrorCode.MissingProperty, label: 'V003 - Missing Property' },
  { value: BlueprintErrorCode.InvalidPropertyValue, label: 'V004 - Invalid Property Value' },
  { value: BlueprintErrorCode.DatabaseConnection, label: 'D001 - Database Connection' },
  { value: BlueprintErrorCode.BlueprintNotFound, label: 'D002 - Blueprint Not Found' },
  { value: BlueprintErrorCode.BlueprintVersionNotFound, label: 'D003 - Blueprint Version Not Found' },
  { value: BlueprintErrorCode.DatabaseQuery, label: 'D004 - Database Query' },
  { value: BlueprintErrorCode.InternalServerError, label: 'S001 - Internal Server Error' },
  { value: BlueprintErrorCode.ResourceExhausted, label: 'S002 - Resource Exhausted' },
  { value: BlueprintErrorCode.SystemUnavailable, label: 'S003 - System Unavailable' },
  { value: BlueprintErrorCode.Unknown, label: 'U001 - Unknown Error' },
];

// Generate arrays from enums for dropdowns
const errorTypes = Object.values(ErrorType);
const severities = Object.values(ErrorSeverity);

async function generateError() {
  try {
    // Ensure we have an execution ID
    if (!executionId.value) {
      executionId.value = `test-${Date.now()}`;
    }

    // Send API request to generate a test error
    const response = await fetch('/api/test/generate-error', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        executionId: executionId.value,
        errorType: errorType.value,
        errorCode: errorCode.value,
        message: message.value,
        severity: severity.value,
        nodeId: nodeId.value,
        recoverable: recoverable.value
      })
    });

    if (!response.ok) {
      throw new Error(`Failed to generate error: ${response.statusText}`);
    }

    // Parse response
    result.value = await response.json();

  } catch (error) {
    console.error('Error generating test error:', error);
    result.value = {
      success: false,
      error: {
        message: error instanceof Error ? error.message : String(error)
      }
    };
  }
}

async function generateScenario() {
  try {
    // Ensure we have an execution ID
    if (!executionId.value) {
      executionId.value = `test-${Date.now()}`;
    }

    // Send API request to generate a test scenario
    const response = await fetch('/api/test/generate-scenario', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        executionId: executionId.value,
        scenarioType: scenarioType.value
      })
    });

    if (!response.ok) {
      throw new Error(`Failed to generate scenario: ${response.statusText}`);
    }

    // Parse response
    result.value = await response.json();

  } catch (error) {
    console.error('Error generating scenario:', error);
    result.value = {
      success: false,
      error: {
        message: error instanceof Error ? error.message : String(error)
      }
    };
  }
}
</script>

<style scoped>
.error-testing-panel {
  background-color: #f9f9f9;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  color: #333;
}

h2, h3, h4 {
  margin-top: 0;
  color: #333;
}

.form-section {
  background-color: #fff;
  border-radius: 6px;
  padding: 15px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 15px;
}

.form-row {
  display: flex;
  gap: 15px;
}

.form-row .form-group {
  flex: 1;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: #555;
}

input, select {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

input[type="checkbox"] {
  width: auto;
  margin-right: 5px;
}

button {
  padding: 10px 16px;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.generate-btn {
  background-color: #3498db;
  color: white;
}

.generate-btn:hover {
  background-color: #2980b9;
}

.scenario-btn {
  background-color: #9b59b6;
  color: white;
}

.scenario-btn:hover {
  background-color: #8e44ad;
}

.result-section {
  background-color: #fff;
  border-radius: 6px;
  padding: 15px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.result-status {
  display: inline-block;
  padding: 5px 12px;
  border-radius: 4px;
  font-weight: 500;
  margin-bottom: 15px;
}

.result-status.success {
  background-color: #2ecc71;
  color: white;
}

.result-status.error {
  background-color: #e74c3c;
  color: white;
}

.error-details, .analysis-details {
  background-color: #f5f5f5;
  border-radius: 4px;
  padding: 15px;
  overflow-x: auto;
  margin-bottom: 15px;
}

pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  white-space: pre-wrap;
}
</style>

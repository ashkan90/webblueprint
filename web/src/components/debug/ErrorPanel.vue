<template>
  <div class="error-panel">
    <div class="panel-header">
      <h3>Errors & Diagnostics</h3>
      <div class="panel-controls">
        <button @click="clearErrors" class="clear-btn">
          <span class="icon">🗑️</span> Clear
        </button>
        <div class="error-filter">
          <label>
            <input type="checkbox" v-model="showLow" /> Low
          </label>
          <label>
            <input type="checkbox" v-model="showMedium" /> Medium
          </label>
          <label>
            <input type="checkbox" v-model="showHigh" /> High
          </label>
          <label>
            <input type="checkbox" v-model="showCritical" /> Critical
          </label>
        </div>
      </div>
    </div>

    <div v-if="analysis && Object.keys(analysis).length > 0" class="error-analysis">
      <h4>Error Analysis</h4>
      <div class="analysis-summary">
        <div class="analysis-item">
          <div class="label">Total Errors:</div>
          <div class="value">{{ analysis.totalErrors }}</div>
        </div>
        <div class="analysis-item">
          <div class="label">Recoverable:</div>
          <div class="value">{{ analysis.recoverableErrors }}</div>
        </div>
      </div>
      
      <div v-if="analysis.topProblemNodes && analysis.topProblemNodes.length > 0" class="problem-nodes">
        <h5>Problem Nodes</h5>
        <div v-for="(node, index) in analysis.topProblemNodes" :key="index" class="problem-node">
          <div class="node-id">{{ getNodeName(node.nodeId) }}</div>
          <div class="error-count">{{ node.count }} {{ node.count === 1 ? 'error' : 'errors' }}</div>
          <button @click="highlightNode(node.nodeId)" class="highlight-btn">Focus</button>
        </div>
      </div>
    </div>

    <div class="error-content" ref="errorContentRef">
      <div v-if="!hasErrors" class="empty-errors">
        No errors to display
      </div>

      <div v-if="hasErrors" class="filter-bar">
        <input 
          v-model="errorFilter" 
          type="text" 
          class="filter-input" 
          placeholder="Filter errors..."
        />
      </div>

      <div
        v-for="(error, index) in displayErrors"
        :key="index"
        :class="['error-entry', error.severity.toLowerCase()]"
      >
        <div class="error-header">
          <span class="error-code">[{{ error.type }}-{{ error.code }}]</span>
          <span class="error-severity">{{ error.severity }}</span>
          <span v-if="error.nodeId" class="error-node" @click="highlightNode(error.nodeId)">
            {{ getNodeName(error.nodeId) }}
          </span>
        </div>
        
        <div class="error-message">{{ error.message }}</div>
        
        <div v-if="error.recoverable" class="recovery-info">
          <span class="recovery-label">Recoverable:</span>
          <span class="recovery-options">
            {{ error.recoveryOptions?.join(', ') }}
          </span>
        </div>
        
        <div v-if="error.expanded && error.details" class="error-details">
          <h5>Error Details</h5>
          <pre>{{ JSON.stringify(error.details, null, 2) }}</pre>
        </div>
        
        <div class="error-footer">
          <span class="error-timestamp">{{ formatTime(error.timestamp) }}</span>
          
          <div class="error-actions">
            <button
              v-if="error.details"
              @click="toggleErrorDetails(error)"
              class="toggle-details-btn"
            >
              {{ error.expanded ? 'Hide Details' : 'Show Details' }}
            </button>
            
            <button
              v-if="error.recoverable"
              @click="recoverError(error)"
              class="recover-btn"
            >
              Recover
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useErrorStore } from '../../stores/errorStore'
import { BlueprintError, ErrorSeverity } from '../../types/errors'

// Props
const props = defineProps<{
  executionId?: string
}>()

// Emits
const emit = defineEmits(['highlightNode', 'recoverError'])

// Store
const errorStore = useErrorStore()

// Refs (reactive state)
const showLow = ref(true)
const showMedium = ref(true)
const showHigh = ref(true)
const showCritical = ref(true)
const errorContentRef = ref<HTMLElement | null>(null)
const errorFilter = ref('')

// Computed properties
const displayErrors = computed(() => {
  // Start with all errors from the store
  let filteredErrors = [...errorStore.errors]

  console.log('Raw errors from store:', errorStore.errors)

  // Filter by execution ID if provided
  if (props.executionId) {
    filteredErrors = filteredErrors.filter(err => err.executionId === props.executionId)
  }

  // Filter by severity
  filteredErrors = filteredErrors.filter(err => {
    switch (err.severity) {
      case ErrorSeverity.Low: return showLow.value
      case ErrorSeverity.Medium: return showMedium.value
      case ErrorSeverity.High: return showHigh.value
      case ErrorSeverity.Critical: return showCritical.value
      default: return true
    }
  })

  // Filter by text search
  if (errorFilter.value) {
    const searchText = errorFilter.value.toLowerCase()
    filteredErrors = filteredErrors.filter(err =>
      err.message.toLowerCase().includes(searchText) ||
      err.type.toLowerCase().includes(searchText) ||
      err.code.toLowerCase().includes(searchText) ||
      (err.nodeId && err.nodeId.toLowerCase().includes(searchText))
    )
  }

  console.log('Filtered errors count:', filteredErrors.length)
  return filteredErrors
})

const hasErrors = computed(() => displayErrors.value.length > 0)
const analysis = computed(() => errorStore.errorAnalysis)

// Lifecycle hooks
onMounted(() => {
  console.log('ErrorPanel mounted, current errors:', errorStore.errors)
  
  // Add watcher to check for error changes
  watch(() => errorStore.errors.length, (newCount, oldCount) => {
    console.log(`Errors count changed from ${oldCount} to ${newCount}`)
  })
})

// Methods
function clearErrors() {
  errorStore.clearErrors()
}

function formatTime(timestamp: string): string {
  if (!timestamp) return ''
  
  try {
    const date = new Date(timestamp)
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch (e) {
    console.error('Error formatting timestamp:', e)
    return timestamp
  }
}

function getNodeName(nodeId: string): string {
  // This would need to get the node name from your node registry
  return nodeId
}

function toggleErrorDetails(error: BlueprintError) {
  // Create a modified copy and update it
  const updatedError = { 
    ...error,
    expanded: !error.expanded 
  }
  
  // Update the error in the store
  errorStore.updateError(updatedError)
  console.log('Toggled error details:', updatedError.expanded)
}

function highlightNode(nodeId: string) {
  emit('highlightNode', nodeId)
}

function recoverError(error: BlueprintError) {
  emit('recoverError', error)
}
</script>

<style scoped>
.error-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252525;
  color: #e0e0e0;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: #333;
  border-bottom: 1px solid #444;
}

.panel-header h3 {
  margin: 0;
  font-size: 1rem;
}

.panel-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.clear-btn {
  background-color: #444;
  border: none;
  color: #e0e0e0;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.8rem;
}

.clear-btn:hover {
  background-color: #555;
}

.error-filter {
  display: flex;
  gap: 8px;
  font-size: 0.8rem;
}

.error-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 8px 48px;
}

.error-analysis {
  background-color: #2a2a2a;
  padding: 10px;
  margin-bottom: 10px;
  border-radius: 4px;
}

.error-analysis h4, .error-analysis h5 {
  margin-top: 0;
  margin-bottom: 8px;
}

.analysis-summary {
  display: flex;
  gap: 20px;
  margin-bottom: 10px;
}

.analysis-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.analysis-item .label {
  font-weight: 500;
  color: #bbb;
}

.problem-nodes {
  margin-top: 10px;
}

.problem-node {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px;
  background-color: #333;
  border-radius: 4px;
  margin-bottom: 4px;
}

.node-id {
  font-weight: 500;
  flex: 1;
}

.highlight-btn {
  background-color: #444;
  border: none;
  color: #e0e0e0;
  padding: 2px 6px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 0.7rem;
}

.highlight-btn:hover {
  background-color: #555;
}

.filter-bar {
  margin-bottom: 10px;
}

.filter-input {
  width: 100%;
  padding: 6px 8px;
  background-color: #333;
  border: 1px solid #444;
  color: #e0e0e0;
  border-radius: 4px;
  font-size: 0.9rem;
}

.error-entry {
  margin-bottom: 10px;
  padding: 10px;
  border-radius: 4px;
  background-color: #2a2a2a;
  font-size: 0.9rem;
}

.error-entry.critical {
  border-left: 4px solid #e74c3c;
}

.error-entry.high {
  border-left: 4px solid #e67e22;
}

.error-entry.medium {
  border-left: 4px solid #f39c12;
}

.error-entry.low {
  border-left: 4px solid #3498db;
}

.error-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.error-code {
  font-family: monospace;
  font-size: 0.8rem;
  color: #999;
}

.error-severity {
  font-weight: bold;
  font-size: 0.7rem;
  padding: 2px 5px;
  border-radius: 3px;
  text-transform: uppercase;
}

.critical .error-severity {
  background-color: #e74c3c;
  color: white;
}

.high .error-severity {
  background-color: #e67e22;
  color: white;
}

.medium .error-severity {
  background-color: #f39c12;
  color: black;
}

.low .error-severity {
  background-color: #3498db;
  color: white;
}

.error-node {
  font-size: 0.8rem;
  background-color: #333;
  padding: 2px 6px;
  border-radius: 3px;
  cursor: pointer;
}

.error-node:hover {
  background-color: #444;
}

.error-message {
  margin-bottom: 8px;
  word-break: break-word;
  line-height: 1.4;
}

.recovery-info {
  font-size: 0.8rem;
  background-color: #333;
  padding: 4px 8px;
  border-radius: 3px;
  margin-bottom: 8px;
}

.recovery-label {
  color: #2ecc71;
  font-weight: bold;
  margin-right: 6px;
}

.error-details {
  background-color: #333;
  padding: 8px;
  border-radius: 4px;
  margin: 10px 0;
  overflow-x: auto;
}

.error-details h5 {
  margin-top: 0;
  margin-bottom: 6px;
  font-size: 0.8rem;
  color: #bbb;
}

.error-details pre {
  margin: 0;
  white-space: pre-wrap;
  font-size: 0.8rem;
}

.error-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
  font-size: 0.8rem;
}

.error-timestamp {
  color: #888;
}

.error-actions {
  display: flex;
  gap: 6px;
}

.toggle-details-btn, .recover-btn {
  background: none;
  border: none;
  font-size: 0.8rem;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 3px;
}

.toggle-details-btn {
  color: #3498db;
}

.toggle-details-btn:hover {
  background-color: #3498db22;
}

.recover-btn {
  color: #2ecc71;
  font-weight: bold;
}

.recover-btn:hover {
  background-color: #2ecc7122;
}

.empty-errors {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100px;
  color: #666;
  font-style: italic;
}
</style>
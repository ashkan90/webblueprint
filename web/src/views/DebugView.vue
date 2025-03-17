<template>
  <div class="debug-view">
    <div class="toolbar">
      <h1>WebBlueprint Debugging</h1>
      <div class="view-tabs">
        <button 
          @click="activeTab = 'logs'" 
          :class="{ active: activeTab === 'logs' }"
        >
          Logs
        </button>
        <button 
          @click="activeTab = 'errors'" 
          :class="{ active: activeTab === 'errors' }"
        >
          Errors
          <span v-if="errorCount > 0" class="badge">{{ errorCount }}</span>
        </button>
        <button 
          @click="activeTab = 'execution'" 
          :class="{ active: activeTab === 'execution' }"
        >
          Execution
        </button>
        <button 
          @click="activeTab = 'testing'" 
          :class="{ active: activeTab === 'testing' }"
        >
          Testing
        </button>
      </div>
      <div class="actions">
        <button @click="clearAll" class="clear-button">
          <span class="icon">🗑️</span> Clear All
        </button>
      </div>
    </div>
    
    <div class="debug-content">
      <LogPanel 
        v-show="activeTab === 'logs'" 
        :executionId="selectedExecutionId"
      />
      
      <ErrorPanel 
        v-show="activeTab === 'errors'" 
        :executionId="selectedExecutionId"
        @highlight-node="highlightNode"
        @recover-error="recoverFromError"
      />
      
<!--      <ExecutionPanel -->
<!--        v-show="activeTab === 'execution'" -->
<!--        :executionId="selectedExecutionId"-->
<!--      />-->
      
      <div v-show="activeTab === 'testing'" class="testing-panel">
        <ErrorTestingPanel />
      </div>
    </div>
    
    <div class="status-bar">
      <div class="status-item">
        <span class="label">Execution:</span>
        <select v-model="selectedExecutionId">
          <option v-for="id in executionIds" :key="id" :value="id">
            {{ id }}
          </option>
        </select>
      </div>
      
      <div class="status-item">
        <span class="label">Status:</span>
        <span 
          class="status-badge" 
          :class="executionStatus"
        >
          {{ executionStatus }}
        </span>
      </div>
      
      <div class="status-item">
        <span class="label">Errors:</span>
        <span 
          class="status-badge" 
          :class="errorCount > 0 ? 'error' : 'success'"
        >
          {{ errorCount }}
        </span>
      </div>
      
      <div class="status-item">
        <span class="label">Logs:</span>
        <span class="status-badge">{{ logCount }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import LogPanel from '../components/debug/LogPanel.vue';
import ErrorPanel from '../components/debug/ErrorPanel.vue';
// import ExecutionPanel from '../components/debug/ExecutionPanel.vue';
import ErrorTestingPanel from '../components/debug/ErrorTestingPanel.vue';

import { useExecutionStore } from '../stores/execution';
import { useErrorStore } from '../stores/errorStore';
import { useErrorViewStore } from '../stores/errorViewStore';

// Stores
const executionStore = useExecutionStore();
const errorStore = useErrorStore();
const errorViewStore = useErrorViewStore();

// State
const activeTab = ref('logs');
const selectedExecutionId = ref('');

// Computed
const executionIds = computed(() => executionStore.executionIds || []);
const executionStatus = computed(() => executionStore.executionStatus || 'idle');
const errorCount = computed(() => errorStore.errors.length);
const logCount = computed(() => executionStore.logs.length);

// Methods
function clearAll() {
  executionStore.clearLogs();
  errorStore.clearErrors();
  errorStore.clearRecoveryAttempts();
}

function highlightNode(nodeId) {
  errorViewStore.selectNode(nodeId);
  // Emit to parent component to handle highlighting in the editor
  // In a real implementation, this would communicate with the editor view
  console.log('Highlight node:', nodeId);
}

async function recoverFromError(error) {
  if (!error.recoverable) return;
  
  try {
    const result = await errorViewStore.recoverFromError(error);
    console.log('Recovery result:', result);
    
    if (result && result.success) {
      // Switch to logs tab to see recovery details
      activeTab.value = 'logs';
    }
  } catch (err) {
    console.error('Failed to recover from error:', err);
  }
}

// Lifecycle
onMounted(() => {
  // Set up WebSocket listeners for error notifications
  errorViewStore.setupWebSocketListeners();
  
  // Set first execution ID if available
  if (executionIds.value.length > 0) {
    selectedExecutionId.value = executionIds.value[0];
  }
});
</script>

<style scoped>
.debug-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #1e1e1e;
  color: #e0e0e0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
}

.toolbar {
  display: flex;
  align-items: center;
  padding: 0 16px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  height: 60px;
}

.toolbar h1 {
  font-size: 1.2rem;
  margin: 0;
  margin-right: 24px;
  font-weight: 500;
}

.view-tabs {
  display: flex;
  gap: 4px;
}

.view-tabs button {
  background: none;
  border: none;
  color: #e0e0e0;
  padding: 8px 16px;
  font-size: 0.9rem;
  cursor: pointer;
  border-radius: 4px;
  position: relative;
}

.view-tabs button:hover {
  background-color: #3a3a3a;
}

.view-tabs button.active {
  background-color: #3c3c3c;
  font-weight: 500;
}

.badge {
  position: absolute;
  top: 2px;
  right: 2px;
  background-color: #e74c3c;
  color: white;
  font-size: 0.7rem;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.actions {
  margin-left: auto;
}

.clear-button {
  background-color: #444;
  border: none;
  color: #e0e0e0;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
}

.clear-button:hover {
  background-color: #555;
}

.debug-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.testing-panel {
  padding: 16px;
  overflow: auto;
  height: 100%;
}

.status-bar {
  height: 30px;
  background-color: #2d2d2d;
  border-top: 1px solid #3d3d3d;
  display: flex;
  align-items: center;
  padding: 0 16px;
  font-size: 0.8rem;
  color: #bbb;
}

.status-item {
  display: flex;
  align-items: center;
  margin-right: 20px;
}

.status-item .label {
  margin-right: 6px;
}

.status-item select {
  background-color: #333;
  border: 1px solid #444;
  color: #e0e0e0;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 0.8rem;
}

.status-badge {
  padding: 2px 6px;
  border-radius: 3px;
  background-color: #333;
}

.status-badge.running {
  background-color: #3498db;
  color: white;
}

.status-badge.completed {
  background-color: #2ecc71;
  color: white;
}

.status-badge.failed {
  background-color: #e74c3c;
  color: white;
}

.status-badge.success {
  background-color: #2ecc71;
  color: white;
}

.status-badge.error {
  background-color: #e74c3c;
  color: white;
}
</style>

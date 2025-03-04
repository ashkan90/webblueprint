<template>
  <div class="log-panel">
    <div class="panel-header">
      <h3>Log Messages</h3>
      <div class="panel-controls">
        <button @click="clearLogs" class="clear-btn">
          <span class="icon">🗑️</span> Clear
        </button>
        <div class="log-filter">
          <label>
            <input type="checkbox" v-model="showDebug" /> Debug
          </label>
          <label>
            <input type="checkbox" v-model="showInfo" /> Info
          </label>
          <label>
            <input type="checkbox" v-model="showWarn" /> Warn
          </label>
          <label>
            <input type="checkbox" v-model="showError" /> Error
          </label>
        </div>
      </div>
    </div>

    <div class="log-content" ref="logContent">
      <div v-if="filteredLogs.length === 0" class="empty-logs">
        No log messages to display
      </div>

      <div v-if="filteredLogs.length > 0" class="panel-header">
        <input
          v-model="logFilter"
          type="text"
          class="search-input"
        />
      </div>

      <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          :class="['log-entry', log.level.toLowerCase()]"
      >
        <div class="log-timestamp">{{ formatTime(log.timestamp) }}</div>
        <div class="log-level">{{ log.level }}</div>
        <div class="log-node">{{ getNodeName(log.nodeId) }}</div>
        <div class="log-message">{{ log.message }}</div>
        <div v-if="log.expanded && log.details" class="log-details">
          <pre>{{ JSON.stringify(log.details, null, 2) }}</pre>
        </div>
        <button
            v-if="log.details"
            @click="toggleLogDetails(index)"
            class="toggle-details"
        >
          {{ log.expanded ? 'Hide Details' : 'Show Details' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useExecutionStore } from '../../stores/execution'
import { useBlueprintStore } from '../../stores/blueprint'
import { useNodeRegistryStore } from '../../stores/nodeRegistry'

const props = defineProps<{
  executionId?: string
}>()

// Stores
const executionStore = useExecutionStore()
const blueprintStore = useBlueprintStore()
const nodeRegistryStore = useNodeRegistryStore()

// State
const showDebug = ref(true)
const showInfo = ref(true)
const showWarn = ref(true)
const showError = ref(true)
const logContent = ref<HTMLElement | null>(null)
const autoScroll = ref(true)
const logFilter = ref('')

// Get logs from execution store
const logs = computed(() => executionStore.logs)
const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    if (logFilter.value !== '') {
      return logFilter.value.includes('-')
        ? log.nodeId.includes(logFilter.value)
        : log.message.toLowerCase().includes(logFilter.value)
    }

    switch (log.level.toLowerCase()) {
      case 'debug': return showDebug.value
      case 'info': return showInfo.value
      case 'warn': return showWarn.value
      case 'error': return showError.value
      default: return true
    }
  })
})

// Methods
function formatTime(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function getNodeName(nodeId: string): string {
  if (!nodeId) return ''

  const node = blueprintStore.getNodeById(nodeId)
  if (!node) return nodeId

  const nodeType = nodeRegistryStore.getNodeTypeById(node.type)
  return nodeType?.name || node.type
}

function clearLogs() {
  executionStore.clearLogs()
}

function toggleLogDetails(index: number) {
  if (filteredLogs.value[index]) {
    filteredLogs.value[index].expanded = !filteredLogs.value[index].expanded
  }
}

function scrollToBottom() {
  if (logContent.value && autoScroll.value) {
    setTimeout(() => {
      if (logContent.value) {
        logContent.value.scrollTop = logContent.value.scrollHeight
      }
    }, 10)
  }
}

// Watch for new logs and scroll to bottom
watch(() => logs.value.length, () => {
  scrollToBottom()
})

// Initialize
onMounted(() => {
  scrollToBottom()
})
</script>

<style scoped>
.log-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252525;
  color: #e0e0e0;
  font-family: monospace;
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

.log-filter {
  display: flex;
  gap: 8px;
  font-size: 0.8rem;
}

.log-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 8px 48px;
}

.log-entry {
  margin-bottom: 4px;
  padding: 6px 8px;
  border-radius: 4px;
  background-color: #2a2a2a;
  font-size: 0.9rem;
  display: grid;
  grid-template-columns: auto auto auto 1fr;
  grid-gap: 8px;
  align-items: start;
}

.log-entry.debug {
  border-left: 3px solid #8e8e8e;
}

.log-entry.info {
  border-left: 3px solid #4a9ce2;
}

.log-entry.warn {
  border-left: 3px solid #e6b41c;
}

.log-entry.error {
  border-left: 3px solid #e74c3c;
}

.log-timestamp {
  color: #888;
  font-size: 0.8rem;
  white-space: nowrap;
}

.log-level {
  font-weight: bold;
  text-transform: uppercase;
  font-size: 0.7rem;
  padding: 2px 4px;
  border-radius: 3px;
  min-width: 40px;
  text-align: center;
}

.debug .log-level {
  background-color: #555;
  color: #fff;
}

.info .log-level {
  background-color: #3498db;
  color: #fff;
}

.warn .log-level {
  background-color: #f39c12;
  color: #333;
}

.error .log-level {
  background-color: #e74c3c;
  color: #fff;
}

.log-node {
  font-weight: 500;
  color: #9a9a9a;
}

.log-message {
  word-break: break-word;
}

.log-details {
  grid-column: 1 / -1;
  background-color: #333;
  padding: 8px;
  border-radius: 4px;
  margin-top: 4px;
  overflow-x: auto;
}

.log-details pre {
  margin: 0;
  white-space: pre-wrap;
}

.toggle-details {
  grid-column: 1 / -1;
  background: none;
  border: none;
  color: #4a9ce2;
  cursor: pointer;
  padding: 2px 0;
  font-size: 0.8rem;
  text-align: left;
}

.toggle-details:hover {
  text-decoration: underline;
}

.empty-logs {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100px;
  color: #666;
  font-style: italic;
}

.search-input,
.search-select {
  flex: 1;
  background-color: #444;
  border: 1px solid #555;
  border-radius: 4px;
  padding: 4px 8px;
  color: white;
}

.search-input:focus,
.search-select:focus {
  outline: none;
  border-color: var(--accent-blue);
}
</style>
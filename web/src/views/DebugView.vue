<template>
  <div class="debug-view">
    <div class="debug-header">
      <h1>Execution Debug</h1>

      <div v-if="executionId" class="execution-info">
        <div class="info-row">
          <div class="info-label">Execution ID:</div>
          <div class="info-value">{{ executionId }}</div>
        </div>

        <div class="info-row">
          <div class="info-label">Status:</div>
          <div :class="['info-value', statusClass]">{{ executionStatus }}</div>
        </div>

        <div class="info-row" v-if="startTime">
          <div class="info-label">Started:</div>
          <div class="info-value">{{ formatDateTime(startTime) }}</div>
        </div>

        <div class="info-row" v-if="endTime">
          <div class="info-label">Ended:</div>
          <div class="info-value">{{ formatDateTime(endTime) }}</div>
        </div>

        <div class="info-row" v-if="duration !== null">
          <div class="info-label">Duration:</div>
          <div class="info-value">{{ formatDuration(duration) }}</div>
        </div>

        <div v-if="errorMessage" class="error-message">
          <div class="info-label">Error:</div>
          <div class="info-value error">{{ errorMessage }}</div>
        </div>

        <div class="actions">
          <button class="btn" @click="goToEditor">
            <span class="icon">✏️</span> Edit Blueprint
          </button>
          <button class="btn" @click="rerunExecution">
            <span class="icon">🔄</span> Re-run
          </button>
        </div>
      </div>

      <div v-else class="no-execution">
        <p>No execution found. Please specify an execution ID.</p>
        <button class="btn" @click="goToHome">Go to Home</button>
      </div>
    </div>

    <DebugPanel
        v-if="executionId"
        :execution-id="executionId"
        :selected-node-id="selectedNodeId"
        @select-node="handleSelectNode"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useExecutionStore } from '../stores/execution'
import { useBlueprintStore } from '../stores/blueprint'
import DebugPanel from '../components/debug/DebugPanel.vue'

// Router
const route = useRoute()
const router = useRouter()

// Stores
const executionStore = useExecutionStore()
const blueprintStore = useBlueprintStore()

// State
const selectedNodeId = ref<string | null>(null)

// Computed
const executionId = computed(() => route.params.executionId as string)
const executionStatus = computed(() => executionStore.executionStatus)
const startTime = computed(() => executionStore.executionStartTime)
const endTime = computed(() => executionStore.executionEndTime)
const duration = computed(() => executionStore.executionDuration)
const errorMessage = computed(() => executionStore.errorMessage)

const statusClass = computed(() => {
  switch (executionStatus.value) {
    case 'running':
      return 'status-running'
    case 'completed':
      return 'status-completed'
    case 'error':
      return 'status-error'
    default:
      return 'status-idle'
  }
})

// Methods
function handleSelectNode(nodeId: string) {
  selectedNodeId.value = nodeId
}

function goToEditor() {
  const blueprintId = executionStore.blueprintId
  if (blueprintId) {
    router.push(`/editor/${blueprintId}`)
  }
}

function goToHome() {
  router.push('/')
}

async function rerunExecution() {
  try {
    const blueprintId = executionStore.blueprintId
    if (blueprintId) {
      const result = await executionStore.executeBlueprint(blueprintId)
      // Update URL to the new execution
      router.push(`/debug/${result.executionId}`)
    }
  } catch (error) {
    console.error('Failed to re-run execution:', error)
    alert('Failed to re-run execution. Please try again.')
  }
}

function formatDateTime(date: Date | null): string {
  if (!date) return ''
  return date.toLocaleString()
}

function formatDuration(ms: number | null): string {
  if (ms === null) return '0ms'

  if (ms < 1000) {
    return `${ms}ms`
  } else if (ms < 60000) {
    return `${(ms / 1000).toFixed(2)}s`
  } else {
    const minutes = Math.floor(ms / 60000)
    const seconds = ((ms % 60000) / 1000).toFixed(2)
    return `${minutes}m ${seconds}s`
  }
}

// Initialize
onMounted(async () => {
  if (executionId.value) {
    try {
      // Load execution data
      await executionStore.loadExecution(executionId.value)
    } catch (error) {
      console.error('Failed to load execution:', error)
      router.push('/')
    }
  }
})
</script>

<style scoped>
.debug-view {
  height: calc(100vh - 50px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.debug-header {
  padding: 20px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

h1 {
  font-size: 1.8rem;
  margin-bottom: 1rem;
  color: var(--accent-blue);
}

.execution-info {
  background-color: #333;
  padding: 15px;
  border-radius: 6px;
}

.info-row {
  display: flex;
  margin-bottom: 8px;
}

.info-label {
  width: 100px;
  color: #aaa;
  font-weight: 500;
}

.info-value {
  flex: 1;
}

.info-value.status-running {
  color: var(--accent-yellow);
}

.info-value.status-completed {
  color: var(--accent-green);
}

.info-value.status-error {
  color: var(--accent-red);
}

.error-message .info-value {
  color: var(--accent-red);
  font-weight: bold;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.btn {
  background-color: #444;
  border: none;
  color: white;
  padding: 8px 12px;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.2s;
}

.btn:hover {
  background-color: #555;
}

.btn .icon {
  margin-right: 4px;
}

.no-execution {
  text-align: center;
  padding: 2rem;
  color: #aaa;
}

.no-execution p {
  margin-bottom: 1rem;
}
</style>
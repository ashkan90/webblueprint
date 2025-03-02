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
  router.push(`/editor/${blueprintStore.blueprint.id}`)
}

function goToHome() {
  router.push('/')
}

async function rerunExecution() {
  try {
    if (blueprintStore.blueprint.id) {
      await executionStore.executeBlueprint(blueprintStore.blueprint.id)
      // Stay on the same page but update the URL with the new execution ID
      router.push(`/debug/${executionStore.currentExecutionId}`)
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
    // Load execution data if not already loaded
    if (executionStore.currentExecutionId !== executionId.value) {
      try {
        await executionStore.loadExecution(executionId.value)
      } catch (error) {
        console.error('Failed to load execution:', error)
      }
    }

    // Load blueprint if not already loaded
    if (!blueprintStore.blueprint.id && executionStore.blueprintId) {
      try {
        await blueprintStore.loadBlueprint(executionStore.blueprintId)
      } catch (error) {
        console.error('Failed to load blueprint:', error)
      }
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
  margin-bottom: 1rem;
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

.no-execution {
  text-align: center;
  padding: 2rem;
  color: #aaa;
}

.no-execution p {
  margin-bottom: 1rem;
}
</style>
<template>
  <div class="debug-panel">
    <div class="debug-header">
      <div class="tab-controls">
        <button
            v-for="tab in tabs"
            :key="tab.id"
            class="tab-button"
            :class="{ 'active': activeTab === tab.id }"
            @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="debug-actions">
        <button class="action-button" @click="refreshDebugData">
          <span class="icon">🔄</span>
        </button>
        <button class="action-button" @click="clearDebugData">
          <span class="icon">🗑️</span>
        </button>
      </div>
    </div>

    <div class="debug-content">
      <!-- Overview Tab -->
      <div v-if="activeTab === 'overview'" class="tab-content">
        <div class="execution-info">
          <div class="info-row">
            <div class="info-label">Status</div>
            <div class="info-value" :class="statusClass">{{ executionStatus }}</div>
          </div>

          <div class="info-row">
            <div class="info-label">Execution ID</div>
            <div class="info-value">{{ executionId || 'None' }}</div>
          </div>

          <div class="info-row" v-if="startTime">
            <div class="info-label">Started</div>
            <div class="info-value">{{ formatTime(startTime) }}</div>
          </div>

          <div class="info-row" v-if="endTime">
            <div class="info-label">Completed</div>
            <div class="info-value">{{ formatTime(endTime) }}</div>
          </div>

          <div class="info-row" v-if="duration !== null">
            <div class="info-label">Duration</div>
            <div class="info-value">{{ formatDuration(duration) }}</div>
          </div>
        </div>

        <div class="node-statuses">
          <h3>Node Execution Status</h3>

          <div v-if="Object.keys(nodeStatuses).length === 0" class="empty-state">
            No nodes have been executed yet.
          </div>

          <div v-else class="node-status-list">
            <div
                v-for="(status, nodeId) in nodeStatuses"
                :key="nodeId"
                class="node-status-item"
                :class="[status.status, { 'selected': selectedNodeId === nodeId }]"
                @click="selectNode(nodeId)"
            >
              <div class="node-status-header">
                <div class="node-name">{{ getNodeName(nodeId) }}</div>
                <div class="status-badge">{{ status.status }}</div>
              </div>

              <div class="node-status-details">
                <div v-if="status.message" class="node-message">
                  {{ status.message }}
                </div>

                <div v-if="status.timestamp" class="node-timestamp">
                  {{ formatTime(status.timestamp) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Node Data Tab -->
      <div v-if="activeTab === 'nodeData'" class="tab-content">
        <div v-if="selectedNodeId" class="node-debug-container">
          <div class="node-debug-header">
            <h3>{{ getNodeName(selectedNodeId) }}</h3>
            <div class="node-debug-status" :class="getNodeStatus(selectedNodeId)?.status || 'idle'">
              {{ getNodeStatus(selectedNodeId)?.status || 'idle' }}
            </div>
          </div>

          <div v-if="nodeDebugData" class="node-debug-content">
            <!-- Input Values -->
            <div class="debug-section">
              <div class="section-header">
                <h4>Inputs</h4>
              </div>

              <JsonTree
                  v-if="nodeDebugData.inputs && Object.keys(nodeDebugData.inputs).length > 0"
                  :data="nodeDebugData.inputs"
                  :expanded="true"
              />
              <div v-else class="empty-state">No input data available</div>
            </div>

            <!-- Output Values -->
            <div class="debug-section">
              <div class="section-header">
                <h4>Outputs</h4>
              </div>

              <JsonTree
                  v-if="nodeDebugData.outputs && Object.keys(nodeDebugData.outputs).length > 0"
                  :data="nodeDebugData.outputs"
                  :expanded="true"
              />
              <div v-else class="empty-state">No output data available</div>
            </div>

            <!-- Debug Snapshots -->
            <div class="debug-section">
              <div class="section-header">
                <h4>Debug Snapshots</h4>
              </div>

              <div v-if="nodeDebugData.snapshots && nodeDebugData.snapshots.length > 0" class="snapshots-list">
                <div
                    v-for="(snapshot, index) in nodeDebugData.snapshots"
                    :key="index"
                    class="snapshot-item"
                >
                  <div class="snapshot-header" @click="toggleSnapshot(index)">
                    <div class="snapshot-description">{{ snapshot.description }}</div>
                    <div class="snapshot-timestamp">{{ formatTime(snapshot.timestamp) }}</div>
                    <div class="snapshot-toggle">
                      {{ expandedSnapshots.includes(index) ? '▼' : '▶' }}
                    </div>
                  </div>

                  <div v-if="expandedSnapshots.includes(index)" class="snapshot-content">
                    <JsonTree :data="snapshot.data" :expanded="false" />
                  </div>
                </div>
              </div>
              <div v-else class="empty-state">No debug snapshots available</div>
            </div>
          </div>

          <div v-else class="empty-state centered">
            <div v-if="isLoadingNodeData">Loading node data...</div>
            <div v-else>No debug data available for this node</div>
          </div>
        </div>

        <div v-else class="empty-state centered">
          Select a node to view its debug data
        </div>
      </div>

      <!-- Data Flow Tab -->
      <div v-if="activeTab === 'dataFlow'" class="tab-content">
        <h3>Data Flow</h3>

        <div v-if="dataFlows.length === 0" class="empty-state">
          No data flow events recorded yet.
        </div>

        <div v-else class="data-flow-list">
          <div
              v-for="(flow, index) in dataFlows"
              :key="index"
              class="data-flow-item"
          >
            <div class="flow-header">
              <div class="flow-source">{{ getNodeName(flow.sourceNodeId) }}</div>
              <div class="flow-arrow">→</div>
              <div class="flow-target">{{ getNodeName(flow.targetNodeId) }}</div>
              <div class="flow-timestamp">{{ formatTime(flow.timestamp) }}</div>
            </div>

            <div class="flow-details">
              <div class="flow-pins">
                <span class="pin-name">{{ flow.sourcePinId }}</span>
                <span class="flow-arrow-small">→</span>
                <span class="pin-name">{{ flow.targetPinId }}</span>
              </div>

              <div class="flow-value-container">
                <JsonTree
                    :data="flow.value"
                    :expanded="false"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Timeline Tab -->
      <div v-if="activeTab === 'timeline'" class="tab-content">
        <h3>Execution Timeline</h3>

        <div v-if="timelineEvents?.length === 0" class="empty-state">
          No execution events recorded yet.
        </div>

        <div v-else class="timeline">
          <div
              v-for="(event, index) in timelineEvents"
              :key="index"
              class="timeline-event"
              :class="event.type"
          >
            <div class="event-time">{{ formatTime(event.timestamp) }}</div>
            <div
                class="event-node"
                :class="{ 'clickable': event.nodeId }"
                @click="event.nodeId && selectNode(event.nodeId)"
            >
              {{ event.nodeId ? getNodeName(event.nodeId) : '' }}
            </div>
            <div class="event-content">
              <div class="event-title">{{ event.title }}</div>
              <div class="event-description">{{ event.description }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Logs Tab -->
      <div v-if="activeTab === 'logs'" class="debug-tab">
        <LogPanel :execution-id="executionId" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useExecutionStore } from '../../stores/execution'
import { useBlueprintStore } from '../../stores/blueprint'
import { useNodeRegistryStore } from '../../stores/nodeRegistry'
import LogPanel from './LogPanel.vue'
import JsonTree from './JsonTree.vue'
import type { NodeExecutionStatus, NodeDebugData, DataFlow } from '../../types/execution'

const props = defineProps<{
  executionId?: string | null
  selectedNodeId?: string | null
}>()

const emit = defineEmits<{
  (e: 'select-node', nodeId: string): void
}>()

// Stores
const executionStore = useExecutionStore()
const blueprintStore = useBlueprintStore()
const nodeRegistryStore = useNodeRegistryStore()

// State
const activeTab = ref('overview')
const expandedSnapshots = ref<number[]>([])
const isLoadingNodeData = ref(false)
const localSelectedNodeId = ref<string | null>(null)

// Tabs
const tabs = [
  { id: 'overview', label: 'Overview' },
  { id: 'nodeData', label: 'Node Data' },
  { id: 'dataFlow', label: 'Data Flow' },
  { id: 'timeline', label: 'Timeline' },
  { id: 'logs', label: 'Logs' }
]

// Computed values
const selectedNodeId = computed(() => props.selectedNodeId || localSelectedNodeId.value)
const executionId = computed(() => props.executionId || executionStore.currentExecutionId)
const executionStatus = computed(() => executionStore.executionStatus)
const nodeStatuses = computed(() => executionStore.nodeStatuses)
const dataFlows = computed(() => executionStore.dataFlows)
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

const nodeDebugData = computed(() => {
  if (!selectedNodeId.value) return null
  return executionStore.getNodeDebugData(selectedNodeId.value)
})

// Timeline events
const timelineEvents = computed(() => {
  const events = []

  // Execution start/end events
  if (startTime.value) {
    events.push({
      type: 'execution',
      timestamp: startTime.value,
      title: 'Execution Started',
      description: `Execution ID: ${executionId.value}`,
      nodeId: null
    })
  }

  if (endTime.value) {
    events.push({
      type: executionStatus.value === 'completed' ? 'completed' : 'error',
      timestamp: endTime.value,
      title: `Execution ${executionStatus.value === 'completed' ? 'Completed' : 'Failed'}`,
      description: `Duration: ${formatDuration(duration.value)}`,
      nodeId: null
    })
  }

  // Node status events
  Object.entries(nodeStatuses.value).forEach(([nodeId, status]) => {
    events.push({
      type: status.status,
      timestamp: status.timestamp,
      title: `Node ${
          status.status === 'executing' ? 'Started' :
              status.status === 'completed' ? 'Completed' : 'Error'
      }`,
      description: status.message || '',
      nodeId
    })
  })

  // Data flow events
  dataFlows.value.forEach((flow) => {
    events.push({
      type: 'dataFlow',
      timestamp: flow.timestamp,
      title: 'Data Flow',
      description: `${flow.sourcePinId} → ${flow.targetPinId}`,
      nodeId: flow.sourceNodeId
    })
  })

  // Sort by timestamp
  return events
  // return events.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime())
})

// Methods
function getNodeName(nodeId: string): string {
  const node = blueprintStore.getNodeById(nodeId)
  if (!node) return nodeId

  const nodeType = nodeRegistryStore.getNodeTypeById(node.type)
  return nodeType?.name || node.type
}

function getNodeStatus(nodeId: string): NodeExecutionStatus | null {
  if (!nodeId) return null
  return executionStore.getNodeStatus(nodeId)
}

function selectNode(nodeId: string) {
  localSelectedNodeId.value = nodeId
  emit('select-node', nodeId)

  // If we're not in the node data tab, switch to it
  if (activeTab.value !== 'nodeData') {
    activeTab.value = 'nodeData'
  }

  // Fetch debug data for this node if we don't have it yet
  if (executionId.value && !executionStore.getNodeDebugData(nodeId)) {
    fetchNodeDebugData(nodeId)
  }
}

function formatTime(date: Date | string | null): string {
  if (!date) return ''

  const dateObj = typeof date === 'string' ? new Date(date) : date
  return dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
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

function toggleSnapshot(index: number) {
  const idx = expandedSnapshots.value.indexOf(index)
  if (idx === -1) {
    expandedSnapshots.value.push(index)
  } else {
    expandedSnapshots.value.splice(idx, 1)
  }
}

async function fetchNodeDebugData(nodeId: string) {
  if (!executionId.value) return

  isLoadingNodeData.value = true

  try {
    await executionStore.fetchNodeDebugData(executionId.value, nodeId)
  } catch (error) {
    console.error('Error fetching node debug data:', error)
  } finally {
    isLoadingNodeData.value = false
  }
}

function refreshDebugData() {
  if (selectedNodeId.value && executionId.value) {
    fetchNodeDebugData(selectedNodeId.value)
  }
}

function clearDebugData() {
  executionStore.clearDebugData()
}

// Watch for changes in selectedNodeId from props
watch(() => props.selectedNodeId, (newNodeId) => {
  if (newNodeId) {
    localSelectedNodeId.value = newNodeId

    // If we're not in the node data tab, switch to it
    if (activeTab.value !== 'nodeData') {
      activeTab.value = 'nodeData'
    }

    // Fetch debug data for this node if we don't have it yet
    if (executionId.value && !executionStore.getNodeDebugData(newNodeId)) {
      fetchNodeDebugData(newNodeId)
    }
  }
})

// Load debug data when the component is mounted
onMounted(() => {
  if (selectedNodeId.value && executionId.value && !executionStore.getNodeDebugData(selectedNodeId.value)) {
    fetchNodeDebugData(selectedNodeId.value)
  }
})
</script>

<style scoped>
.debug-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #2d2d2d;
  border-top: 1px solid #3d3d3d;
}

.debug-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background-color: #333;
  border-bottom: 1px solid #444;
}

.tab-controls {
  display: flex;
  gap: 2px;
}

.tab-button {
  background-color: #444;
  border: none;
  color: #ddd;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.tab-button:hover {
  background-color: #555;
}

.tab-button.active {
  background-color: var(--accent-blue);
  color: white;
}

.debug-actions {
  display: flex;
  gap: 4px;
}

.action-button {
  background-color: #444;
  border: none;
  color: #ddd;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s;
}

.action-button:hover {
  background-color: #555;
}

.debug-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.tab-content {
  height: 100%;
}

.empty-state {
  color: #aaa;
  margin: 16px 0;
  font-style: italic;
}

.empty-state.centered {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.execution-info {
  background-color: #333;
  padding: 16px;
  border-radius: 4px;
  margin-bottom: 16px;
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

.node-statuses h3,
.tab-content > h3 {
  margin-top: 0;
  margin-bottom: 16px;
  font-size: 1.1rem;
  color: #ddd;
  border-bottom: 1px solid #444;
  padding-bottom: 8px;
}

.node-status-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-status-item {
  background-color: #333;
  border-radius: 4px;
  padding: 8px;
  cursor: pointer;
  transition: transform 0.1s, box-shadow 0.1s;
}

.node-status-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.node-status-item.selected {
  outline: 2px solid var(--accent-blue);
}

.node-status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.node-name {
  font-weight: 500;
}

.status-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 10px;
  text-transform: uppercase;
  font-weight: bold;
  background-color: #444;
}

.node-status-item.executing .status-badge {
  background-color: var(--accent-yellow);
  color: #333;
}

.node-status-item.completed .status-badge {
  background-color: var(--accent-green);
  color: white;
}

.node-status-item.error .status-badge {
  background-color: var(--accent-red);
  color: white;
}

.node-status-details {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  color: #aaa;
}

.node-debug-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.node-debug-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.node-debug-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.node-debug-status {
  font-size: 10px;
  padding: 2px 10px;
  border-radius: 10px;
  text-transform: uppercase;
  font-weight: bold;
  background-color: #444;
}

.node-debug-status.executing {
  background-color: var(--accent-yellow);
  color: #333;
}

.node-debug-status.completed {
  background-color: var(--accent-green);
  color: white;
}

.node-debug-status.error {
  background-color: var(--accent-red);
  color: white;
}

.node-debug-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.debug-section {
  background-color: #333;
  border-radius: 4px;
  overflow: hidden;
}

.section-header {
  padding: 8px 12px;
  background-color: #444;
}

.section-header h4 {
  margin: 0;
  font-size: 0.9rem;
  color: #ddd;
}

.snapshots-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.snapshot-item {
  background-color: #333;
}

.snapshot-header {
  padding: 8px 12px;
  display: flex;
  align-items: center;
  cursor: pointer;
  background-color: #3d3d3d;
}

.snapshot-header:hover {
  background-color: #444;
}

.snapshot-description {
  flex: 1;
  font-weight: 500;
}

.snapshot-timestamp {
  color: #aaa;
  font-size: 0.8rem;
  margin: 0 8px;
}

.snapshot-toggle {
  color: #aaa;
  width: 16px;
  text-align: center;
}

.snapshot-content {
  padding: 12px;
}

.data-flow-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.data-flow-item {
  background-color: #333;
  border-radius: 4px;
  overflow: hidden;
}

.flow-header {
  padding: 8px 12px;
  background-color: #444;
  display: flex;
  align-items: center;
}

.flow-source, .flow-target {
  font-weight: 500;
}

.flow-arrow {
  margin: 0 8px;
  color: #aaa;
}

.flow-timestamp {
  margin-left: auto;
  font-size: 0.8rem;
  color: #aaa;
}

.flow-details {
  padding: 12px;
}

.flow-pins {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  color: #aaa;
  font-size: 0.9rem;
}

.flow-arrow-small {
  margin: 0 8px;
}

.pin-name {
  padding: 2px 6px;
  background-color: #444;
  border-radius: 4px;
}

.flow-value-container {
  background-color: #2d2d2d;
  border-radius: 4px;
  padding: 8px;
}

.timeline {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.timeline-event {
  display: flex;
  border-left: 3px solid #555;
  padding-left: 16px;
  position: relative;
}

.timeline-event:before {
  content: '';
  position: absolute;
  left: -7px;
  top: 0;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background-color: #555;
}

.timeline-event.executing:before {
  background-color: var(--accent-yellow);
}

.timeline-event.completed:before {
  background-color: var(--accent-green);
}

.timeline-event.error:before {
  background-color: var(--accent-red);
}

.timeline-event.dataFlow:before {
  background-color: var(--accent-blue);
}

.timeline-event.execution:before {
  background-color: #8e44ad;
}

.event-time {
  width: 110px;
  font-size: 0.8rem;
  color: #aaa;
}

.event-node {
  width: 150px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.event-node.clickable {
  cursor: pointer;
  text-decoration: underline;
}

.event-node.clickable:hover {
  color: var(--accent-blue);
}

.event-content {
  flex: 1;
}

.event-title {
  font-weight: 500;
}

.event-description {
  font-size: 0.9rem;
  color: #aaa;
}
</style>
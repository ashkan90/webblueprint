<template>
  <div class="enhanced-editor-view">
    <div class="toolbar">
      <div class="blueprint-info">
        <input
            v-model="blueprintName"
            class="blueprint-name"
            placeholder="Blueprint Name"
            @change="updateBlueprintName"
        />
        <div class="blueprint-id" v-if="blueprintStore.blueprint.id">ID: {{ blueprintStore.blueprint.id }}</div>
      </div>

      <div class="tool-buttons">
        <div class="execution-mode-selector">
          <label for="execution-mode">Execution: </label>
          <select id="execution-mode" v-model="executionMode" class="mode-select" :disabled="isExecuting">
            <option value="direct">Direct Mode</option>
            <option value="actor">Actor Mode</option>
          </select>
        </div>

        <button @click="executeBlueprint" :disabled="isExecuting" class="btn primary">
          <span class="icon">▶</span>
          {{ isExecuting ? 'Running...' : 'Execute' }}
        </button>

        <button @click="saveBlueprint" :disabled="isExecuting" class="btn" :class="{ 'has-changes': blueprintStore.hasUnsavedChanges }">
          <span class="icon">💾</span> Save
        </button>

        <button @click="toggleDebugPanel" class="btn" :class="{ 'active': showDebugPanel }">
          <span class="icon">🔍</span> Debug
        </button>
        
        <button @click="toggleVersionsPanel" class="btn" :class="{ 'active': showVersionsPanel }">
          <span class="icon">📋</span> Versions
        </button>
      </div>
    </div>

    <div class="editor-container" :class="{ 'with-debug': showDebugPanel || showVersionsPanel }">
      <div class="node-palette">
        <EnhancedNodePalette @node-added="handleNodeAdded" />
      </div>

      <div class="canvas-container" ref="canvasContainer">
        <BlueprintCanvas
            ref="canvas"
            :nodes="nodes"
            :connections="connections"
            :node-statuses="nodeStatuses"
            :active-connections="activeConnections"
            :active-pins="activePins"
            @node-added="handleNodeAdded"
            @node-selected="handleNodeSelected"
            @node-deselected="handleNodeDeselected"
            @node-moved="handleNodeMoved"
            @connection-created="handleConnectionCreated"
            @connection-deleted="handleConnectionDeleted"
            @node-deleted="handleNodeDeleted"
        />
      </div>

      <div v-if="selectedNode" class="node-properties">
        <EnhancedPropertyEditor
            :node="selectedNode"
            :node-type="getNodeType(selectedNode.type)"
            @property-changed="handlePropertyChanged"
            @pin-default-changed="handlePinDefaultChanged"
            selected/>
      </div>
    </div>

    <div v-if="showDebugPanel" class="bottom-panel">
      <DebugPanel
          :execution-id="currentExecutionId"
          :selected-node-id="selectedNodeId"
          @select-node="handleDebugNodeSelected"
      />
    </div>
    
    <div v-if="showVersionsPanel" class="bottom-panel versions-panel-container">
      <VersionsPanel />
    </div>

    <!-- Execution result modal -->
    <div v-if="showResultModal" class="modal-backdrop">
      <div class="modal">
        <div class="modal-header">
          <h3>Execution {{ executionResult?.success ? 'Completed' : 'Failed' }}</h3>
          <button class="close-btn" @click="closeResultModal">×</button>
        </div>
        <div class="modal-body">
          <p v-if="executionResult?.success" class="success-message">
            Blueprint executed successfully!
          </p>
          <p v-else class="error-message">
            Blueprint execution failed: {{ executionResult?.error }}
          </p>

          <div class="execution-info">
            <div class="info-item">
              <span class="label">Execution ID:</span>
              <span class="value">{{ executionResult?.executionId }}</span>
            </div>
            <div class="info-item">
              <span class="label">Duration:</span>
              <span class="value">{{ executionDuration }}</span>
            </div>
            <div class="info-item">
              <span class="label">Mode:</span>
              <span class="value">{{ executionMode }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeResultModal">Close</button>
          <button class="btn primary" @click="openDebugPanelWithResult">
            View Debug Data
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, provide } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { v4 as uuid } from 'uuid'
import { useBlueprintStore } from '../stores/blueprint'
import { useNodeRegistryStore } from '../stores/nodeRegistry'
import { useExecutionStore } from '../stores/execution'
import { useWorkspaceStore } from '../stores/workspace'
import type { Node, Connection } from '../types/blueprint'
import type { NodeTypeDefinition } from '../types/nodes'
import BlueprintCanvas from '../components/editor/BlueprintCanvas.vue'
import EnhancedNodePalette from '../components/editor/EnhancedNodePalette.vue'
import EnhancedPropertyEditor from '../components/editor/EnhancedPropertyEditor.vue'
import DebugPanel from '../components/debug/DebugPanel.vue'
import VersionsPanel from '../components/VersionsPanel.vue' // Import the new versions panel
import {
  executeBlueprint as executeBlueprintFn,
  executionManager,
  getExecutionMode,
  setExecutionMode
} from '../bootstrap/blueprintSystem'
import { ExecutionMode } from '../services/executionService'

const route = useRoute()
const router = useRouter()
const blueprintStore = useBlueprintStore()
const nodeRegistryStore = useNodeRegistryStore()
const executionStore = useExecutionStore()
const workspaceStore = useWorkspaceStore()

// State
const blueprintName = ref('')
const selectedNodeId = ref<string | null>(null)
const showDebugPanel = ref(false)
const showVersionsPanel = ref(false) // New state for versions panel
const canvas = ref<InstanceType<typeof BlueprintCanvas> | null>(null)
const canvasContainer = ref<HTMLElement | null>(null)
const showResultModal = ref(false)
const executionResult = ref<{
  duration: string;
  executionId: string;
  success: boolean;
  error?: string;
}>()
const executionMode = ref<ExecutionMode>(getExecutionMode())

// Computed values
const nodes = computed(() => blueprintStore.nodes)
const connections = computed(() => blueprintStore.connections)
const selectedNode = computed(() =>
    selectedNodeId.value ? blueprintStore.getNodeById(selectedNodeId.value) : null
)
const nodeStatuses = computed(() => executionStore.nodeStatuses)
const isExecuting = computed(() => executionStore.isExecuting)
const currentExecutionId = computed(() => executionStore.currentExecutionId)
const executionDuration = computed(() => {
  if (!executionStore.executionDuration) return '0ms'

  const ms = executionStore.executionDuration
  if (ms < 1000) {
    return `${ms}ms`
  } else {
    return `${(ms / 1000).toFixed(2)}s`
  }
})

// Get active connections and pins from execution manager
const activeConnections = computed(() => executionManager.getActiveConnections())
const activePins = computed(() => executionManager.getActivePins())

// Methods
function getNodeType(typeId: string): NodeTypeDefinition | null {
  return nodeRegistryStore.getNodeTypeById(typeId)
}

function handleNodeAdded(node: Node) {
  // Check for duplicates
  const existingNodeIndex = blueprintStore.nodes.findIndex(n => n.id === node.id)

  if (existingNodeIndex === -1) {
    // Node doesn't exist, add it
    blueprintStore.addNode(node)
  } else {
    // Node already exists with this ID, generate a new ID and add
    const newNode = {
      ...node,
      id: uuid() // Generate new ID
    }
    blueprintStore.addNode(newNode)
  }

  // Select the newly added node
  selectedNodeId.value = node.id
}

function handleNodeSelected(nodeId: string) {
  selectedNodeId.value = nodeId
}

function handleDebugNodeSelected(nodeId: string) {
  // Select the node both in the debug panel and highlight it in the canvas
  selectedNodeId.value = nodeId

  // TODO: Could also scroll the canvas to show the selected node
}

function handleNodeDeselected() {
  selectedNodeId.value = null
}

function handleNodeMoved(nodeId: string, position: { x: number, y: number }) {
  blueprintStore.updateNodePosition(nodeId, { x: position.x, y: position.y })
}

function handleConnectionCreated(connection: Connection) {
  blueprintStore.addConnection(connection)
}

function handleConnectionDeleted(connectionId: string) {
  blueprintStore.removeConnection(connectionId)
}

function handleNodeDeleted(nodeId: string) {
  blueprintStore.removeNode(nodeId)
  if (selectedNodeId.value === nodeId) {
    selectedNodeId.value = null
  }
}

function handlePropertyChanged(nodeId: string, propertyName: string, value: any) {
  blueprintStore.updateNodeProperty(nodeId, propertyName, value)
}

function handlePinDefaultChanged(nodeId: string, pinId: string, value: any) {
  // Store pin defaults with a special naming convention
  blueprintStore.updateNodeProperty(nodeId, `input_${pinId}`, value)
}

function updateBlueprintName() {
  blueprintStore.blueprint.name = blueprintName.value
}

async function saveBlueprint() {
  try {
    // If this is a new blueprint with no ID, generate one
    if (!blueprintStore.blueprint.id) {
      blueprintStore.blueprint.id = uuid()
    }

    // If the blueprint has no name, use a default
    if (!blueprintStore.blueprint.name) {
      blueprintName.value = 'Untitled Blueprint'
      blueprintStore.blueprint.name = blueprintName.value
    }

    // Get the current workspace ID
    const workspaceId = workspaceStore.currentWorkspace?.id
    if (!workspaceId) {
      throw new Error('No active workspace found')
    }

    await blueprintStore.saveBlueprint(workspaceId)

    // Update the route if this is a new blueprint
    if (route.params.id !== blueprintStore.blueprint.id) {
      router.push(`/editor2/${blueprintStore.blueprint.id}`)
    }
  } catch (error) {
    console.error('Failed to save blueprint:', error)
    alert('Failed to save blueprint. Please try again.')
  }
}

async function executeBlueprint() {
  try {
    // Save the blueprint first if it has changes
    if (!blueprintStore.blueprint.id || blueprintStore.hasUnsavedChanges) {
      await saveBlueprint()
    }

    // Update the execution mode before running
    setExecutionMode(executionMode.value)

    // Execute the blueprint
    const result = await executeBlueprintFn(blueprintStore.blueprint.id)

    // Show result modal
    executionResult.value = {
      duration: result.duration || executionDuration.value,
      executionId: result.executionId,
      success: result.success,
      error: result.error
    }
    showResultModal.value = true
  } catch (error) {
    console.error('Failed to execute blueprint:', error)
    alert('Failed to execute blueprint. Please try again.')
  }
}

function toggleDebugPanel() {
  showDebugPanel.value = !showDebugPanel.value
  if (showDebugPanel.value) {
    showVersionsPanel.value = false
  }
}

function toggleVersionsPanel() {
  showVersionsPanel.value = !showVersionsPanel.value
  if (showVersionsPanel.value) {
    showDebugPanel.value = false
  }
}

function closeResultModal() {
  showResultModal.value = false
}

function openDebugPanelWithResult() {
  showDebugPanel.value = true
  showVersionsPanel.value = false
  closeResultModal()
}

// Watch for changes in execution mode
watch(() => executionMode.value, (newMode) => {
  setExecutionMode(newMode)
})

// Provide canvas container to children components
provide('canvasContainer', canvasContainer)

// Load blueprint on mount
onMounted(async () => {
  // Load node types
  await nodeRegistryStore.fetchNodeTypes()

  // Load blueprint if ID is provided
  if (route.params.id) {
    try {
      await blueprintStore.loadBlueprint(route.params.id as string)
      blueprintName.value = blueprintStore.blueprint.name
    } catch (error) {
      console.error('Failed to load blueprint:', error)
      // Create a new blueprint if loading fails
      blueprintStore.createBlueprint('New Blueprint')
      blueprintName.value = blueprintStore.blueprint.name
    }
  } else {
    // Create a new blueprint
    blueprintStore.createBlueprint('New Blueprint')
    blueprintName.value = blueprintStore.blueprint.name
  }
})
</script>

<style scoped>
.enhanced-editor-view {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 50px); /* Subtract header height */
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  flex: 0 0 auto;
}

.blueprint-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.blueprint-name {
  font-size: 1.2rem;
  background-color: transparent;
  border: 1px solid transparent;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  margin-bottom: 4px;
}

.blueprint-name:hover, .blueprint-name:focus {
  border-color: #3d3d3d;
  background-color: #333;
  outline: none;
}

.blueprint-id {
  font-size: 0.7rem;
  color: #777;
  padding-left: 8px;
}

.tool-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.editor-container {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.editor-container.with-debug {
  height: 60%;
}

.bottom-panel {
  height: 40%;
  border-top: 1px solid #3d3d3d;
  background-color: #2d2d2d;
  overflow: hidden;
}

.versions-panel-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.node-palette {
  width: 250px;
  background-color: #2d2d2d;
  border-right: 1px solid #3d3d3d;
  overflow-y: auto;
  flex: 0 0 auto;
}

.canvas-container {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.node-properties {
  width: 300px;
  background-color: #2d2d2d;
  border-left: 1px solid #3d3d3d;
  overflow-y: auto;
  flex: 0 0 auto;
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

.btn.primary {
  background-color: #3a8cd7;
}

.btn.primary:hover {
  background-color: #4a9de7;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn.active {
  background-color: #555;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
}

.btn.has-changes {
  background-color: #3a8cd7;
}

.icon {
  font-size: 0.9em;
}

.execution-mode-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  color: #aaa;
}

.mode-select {
  background-color: #444;
  border: 1px solid #555;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9rem;
}

.mode-select:focus {
  outline: none;
  border-color: var(--accent-blue);
}

/* Modal styles */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background-color: #2d2d2d;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #3d3d3d;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
}

.close-btn {
  background: none;
  border: none;
  color: #aaa;
  font-size: 1.5rem;
  cursor: pointer;
}

.close-btn:hover {
  color: white;
}

.modal-body {
  padding: 16px;
}

.modal-footer {
  padding: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  border-top: 1px solid #3d3d3d;
}

.success-message {
  color: #4caf50;
  margin-bottom: 16px;
}

.error-message {
  color: #f44336;
  margin-bottom: 16px;
}

.execution-info {
  background-color: #333;
  padding: 12px;
  border-radius: 4px;
}

.info-item {
  display: flex;
  margin-bottom: 8px;
}

.info-item .label {
  font-weight: 500;
  width: 100px;
  color: #aaa;
}

.info-item .value {
  flex: 1;
}
</style>
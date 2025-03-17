<template>
  <div class="editor-view">
    <div class="toolbar">
      <div class="blueprint-info">
        <template v-if="currentEditingFunction">
          <button @click="backToMainEditor" class="back-btn">
            <span class="icon">←</span> Back to Main Blueprint
          </button>
          <span class="editor-title">Function: {{ getCurrentFunctionName() }}</span>
        </template>
        <template v-else>
          <input
              v-model="blueprintName"
              class="blueprint-name"
              placeholder="Blueprint Name"
              @change="updateBlueprintName"
          />
        </template>
      </div>

      <div class="tool-buttons">
        <button @click="executeBlueprint" :disabled="isExecuting || currentEditingFunction !== null" class="btn primary">
          <span class="icon">▶</span>
          {{ isExecuting ? 'Running...' : 'Execute' }}
        </button>

        <button @click="saveBlueprint" :disabled="isExecuting" class="btn" :class="{ 'has-changes': blueprintStore.hasUnsavedChanges }">
          <span class="icon">💾</span> Save
        </button>

        <button @click="toggleDebugPanel" :disabled="currentEditingFunction !== null" class="btn" :class="{ 'active': showDebugPanel }">
          <span class="icon">🔍</span> Debug
        </button>

        <button @click="toggleErrorPanel" :disabled="currentEditingFunction !== null" class="btn" :class="{ 'active': showErrorPanel }">
          <span class="icon">⚠️</span> Errors
          <span v-if="errorStore.hasErrors" class="badge" :class="{ 'critical': errorStore.hasCriticalErrors }">
            {{ errorStore.errors.length }}
          </span>
        </button>
      </div>
    </div>

    <div class="editor-container" :class="{ 'with-debug': showDebugPanel }">
      <div class="node-palette">
        <BlueprintLeftPanel
            @add-node="handleNodeAdded"
            @function-double-clicked="handleFunctionDoubleClicked"
        />
      </div>

      <div class="canvas-container" ref="canvasRef">
        <BlueprintCanvas
            ref="canvas"
            :nodes="currentEditingFunction ? getCurrentFunctionNodes() : nodes"
            :connections="currentEditingFunction ? getCurrentFunctionConnections() : connections"
            :node-statuses="nodeStatuses"
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
        <NodeProperties
            :node="selectedNode"
            :node-type="getNodeType(selectedNode.type)"
            @property-changed="handlePropertyChanged"
            @pin-default-changed="handlePinDefaultChanged"
            selected/>
      </div>
    </div>

    <DebugPanel
        v-if="showDebugPanel && !currentEditingFunction"
        :execution-id="currentExecutionId"
        :selected-node-id="selectedNodeId"
    />
    
    <!-- Error Panel - Always visible but can be toggled -->
    <div v-if="showErrorPanel" class="error-panel-container">
      <div class="error-panel-header">
        <h3>Errors & Diagnostics</h3>
        <button @click="toggleErrorPanel" class="close-btn">×</button>
      </div>
      <ErrorPanel 
        :execution-id="currentExecutionId"
        @highlight-node="handleNodeHighlighted"
        @recover-error="handleErrorRecovery"
      />
    </div>
    
    <!-- Bottom Drawer with Content and Versions tabs -->
    <BottomDrawer v-if="!currentEditingFunction">
      <template #content>
        <ContentBrowserPanel
            @asset-opened="handleAssetOpened"
        />
      </template>
      <template #versions>
        <VersionsPanel />
      </template>
    </BottomDrawer>

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
              <span class="value">{{ executionResult?.duration }}</span>
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
import { useErrorViewStore } from '../stores/errorViewStore'
import { useErrorStore } from '../stores/errorStore'
import type { Node, Connection } from '../types/blueprint'
import type { NodeTypeDefinition } from '../types/nodes'
import type { BlueprintError } from '../types/errors'
import { Asset, AssetType } from '../types/mockPersistent'
import BlueprintCanvas from '../components/editor/BlueprintCanvas.vue'
import NodeProperties from '../components/editor/NodeProperties.vue'
import DebugPanel from '../components/debug/DebugPanel.vue'
import ErrorPanel from '../components/debug/ErrorPanel.vue'
import VersionsPanel from '../components/VersionsPanel.vue'
import BlueprintLeftPanel from "../components/editor/BlueprintLeftPanel.vue"
import ContentBrowserPanel from "../components/ContentBrowserPanel.vue"
import BottomDrawer from "../components/drawer/BottomDrawer.vue"
import {useWorkspaceStore} from "../stores/workspace";

const route = useRoute()
const router = useRouter()
const blueprintStore = useBlueprintStore()
const workspaceStore = useWorkspaceStore()
const nodeRegistryStore = useNodeRegistryStore()
const executionStore = useExecutionStore()
const errorStore = useErrorStore()
const errorViewStore = useErrorViewStore()

// State
const blueprintName = ref('')
const selectedNodeId = ref<string | null>(null)
const showDebugPanel = ref(false)
const showErrorPanel = ref(false)
const canvas = ref<InstanceType<typeof BlueprintCanvas> | null>(null)
const canvasRef = ref<HTMLElement | null>(null)
const showResultModal = ref(false)
const executionResult = ref<{
  duration: string;
  executionId: string;
  success: boolean;
  error?: string;
}>()

provide('canvasContainer', canvasRef)

// Computed values
const nodes = computed(() => blueprintStore.nodes)
const connections = computed(() => blueprintStore.connections)
const currentEditingFunction = computed(() => blueprintStore.currentEditingFunction)
const selectedNode = computed(() => {
  if (!selectedNodeId.value) return null;

  if (currentEditingFunction.value) {
    // Get node from function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      return func.nodes.find(node => node.id === selectedNodeId.value) || null;
    }
    return null;
  } else {
    // Get node from main blueprint
    return blueprintStore.getNodeById(selectedNodeId.value);
  }
})
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

// Methods
function getNodeType(typeId: string): NodeTypeDefinition | null {
  return nodeRegistryStore.getNodeTypeById(typeId)
}

// Function editing methods
function handleFunctionDoubleClicked(functionId: string) {
  // Switch to function editing mode
  blueprintStore.currentEditingFunction = functionId;
  // Clear any selected node
  selectedNodeId.value = null;
  // Disable debug panel while editing a function
  showDebugPanel.value = false;
}

function backToMainEditor() {
  // Switch back to main blueprint editing
  blueprintStore.currentEditingFunction = null;
  // Clear any selected node
  selectedNodeId.value = null;
}

function getCurrentFunctionName(): string {
  if (!currentEditingFunction.value) return '';

  const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
  return func ? func.name : 'Unknown Function';
}

function getCurrentFunctionNodes(): Node[] {
  if (!currentEditingFunction.value) return [];

  const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
  return func ? func.nodes : [];
}

function getCurrentFunctionConnections(): Connection[] {
  if (!currentEditingFunction.value) return [];

  const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
  return func ? func.connections : [];
}

function handleNodeAdded(node: Node) {
  console.log('Adding node:', node); // Debug

  // Create a copy with a new ID if there's a duplicate
  const nodeToAdd = { ...node };

  // Check if we're editing a function or the main blueprint
  if (currentEditingFunction.value) {
    // Find the function we're editing
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      // Check for duplicates in the function's nodes
      const existingNodeIndex = func.nodes.findIndex(n => n.id === node.id);
      if (existingNodeIndex !== -1) {
        // Generate a new ID if node already exists
        nodeToAdd.id = uuid();
      }

      // Ensure properties is initialized
      if (!nodeToAdd.properties) {
        nodeToAdd.properties = [];
      }

      // Add the node to the function
      func.nodes.push(JSON.parse(JSON.stringify(nodeToAdd)));
    }
  } else {
    // We're in the main blueprint
    // Ensure we're not adding duplicates by checking IDs
    const existingNodeIndex = blueprintStore.nodes.findIndex(n => n.id === node.id);
    if (existingNodeIndex !== -1) {
      // Generate a new ID if node already exists
      nodeToAdd.id = uuid();
    }

    // Add to the main blueprint
    blueprintStore.addNode(nodeToAdd);
  }

  // Select the newly added node
  selectedNodeId.value = nodeToAdd.id;
}

function handleNodeSelected(nodeId: string) {
  selectedNodeId.value = nodeId
}

function handleNodeDeselected() {
  selectedNodeId.value = null
}

function handleNodeMoved(nodeId: string, position: { x: number, y: number }) {
  if (currentEditingFunction.value) {
    // Update node position within the function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      const node = func.nodes.find(n => n.id === nodeId);
      if (node) {
        node.position = { x: position.x, y: position.y };
      }
    }
  } else {
    // Update node position in the main blueprint
    blueprintStore.updateNodePosition(nodeId, { x: position.x, y: position.y });
  }
}

function handleConnectionCreated(connection: Connection) {
  if (currentEditingFunction.value) {
    // Add connection to the function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      // Ensure connection has an ID
      if (!connection.id) {
        connection.id = uuid();
      }

      // Check for duplicates
      const exists = func.connections.some(conn =>
          conn.sourceNodeId === connection.sourceNodeId &&
          conn.sourcePinId === connection.sourcePinId &&
          conn.targetNodeId === connection.targetNodeId &&
          conn.targetPinId === connection.targetPinId
      );

      if (!exists) {
        func.connections.push(connection);
      }
    }
  } else {
    // Add connection to the main blueprint
    blueprintStore.addConnection(connection);
  }
}

function handleConnectionDeleted(connectionId: string) {
  if (currentEditingFunction.value) {
    // Remove connection from the function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      func.connections = func.connections.filter(conn => conn.id !== connectionId);
    }
  } else {
    // Remove connection from the main blueprint
    blueprintStore.removeConnection(connectionId);
  }
}

function handleNodeDeleted(nodeId: string) {
  if (currentEditingFunction.value) {
    // Remove node from the function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      // Remove connections first
      func.connections = func.connections.filter(
          conn => conn.sourceNodeId !== nodeId && conn.targetNodeId !== nodeId
      );

      // Remove the node
      func.nodes = func.nodes.filter(node => node.id !== nodeId);
    }
  } else {
    // Remove node from the main blueprint
    blueprintStore.removeNode(nodeId);
  }

  if (selectedNodeId.value === nodeId) {
    selectedNodeId.value = null;
  }
}

function handlePropertyChanged(nodeId: string, propertyName: string, value: any) {
  if (currentEditingFunction.value) {
    // Update node property within the function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      const node = func.nodes.find(n => n.id === nodeId);
      if (node) {
        // Find the property
        const propIndex = node.properties.findIndex(p => p.name === propertyName);
        if (propIndex !== -1) {
          // Update existing property
          node.properties[propIndex].value = value;
        } else {
          // Add new property
          node.properties.push({ name: propertyName, value });
        }
      }
    }
  } else {
    // Update node property in the main blueprint
    blueprintStore.updateNodeProperty(nodeId, propertyName, value);
  }
}

function handlePinDefaultChanged(nodeId: string, pinId: string, value: any) {
  const propertyName = `input_${pinId}`;

  if (currentEditingFunction.value) {
    // Update pin default within the function
    const func = blueprintStore.functions.find(f => f.id === currentEditingFunction.value);
    if (func) {
      const node = func.nodes.find(n => n.id === nodeId);
      if (node) {
        // Find the property
        const propIndex = node.properties.findIndex(p => p.name === propertyName);
        if (propIndex !== -1) {
          // Update existing property
          node.properties[propIndex].value = value;
        } else {
          // Add new property
          node.properties.push({ name: propertyName, value });
        }
      }
    }
  } else {
    // Update pin default in the main blueprint
    blueprintStore.updateNodeProperty(nodeId, propertyName, value);
  }
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

    await blueprintStore.saveBlueprint(workspaceStore.currentWorkspace.id)

    // Update the route if this is a new blueprint
    if (route.params.id !== blueprintStore.blueprint.id) {
      await router.push(`/editor/${blueprintStore.blueprint.id}`)
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

    const result = await executionStore.executeBlueprint(blueprintStore.blueprint.id)

    // Show result modal
    executionResult.value = {
      duration: result.duration,
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
  // Only allow toggling debug panel when not editing a function
  if (!currentEditingFunction.value) {
    showDebugPanel.value = !showDebugPanel.value
  }
}

function toggleErrorPanel() {
  // Only allow toggling error panel when not editing a function
  if (!currentEditingFunction.value) {
    showErrorPanel.value = !showErrorPanel.value
  }
}

function handleNodeHighlighted(nodeId: string) {
  // Handle node highlighting - select the node and scroll to it
  selectedNodeId.value = nodeId
  if (canvas.value) {
    canvas.value.centerOnNode(nodeId)
  }
}

function handleErrorRecovery(error: BlueprintError) {
  // Attempt to recover from an error
  errorStore.recoverFromError(error).then(result => {
    if (result && result.success) {
      // If recovery was successful, update UI or take action
      console.log('Recovery successful:', result)
    }
  })
}

function closeResultModal() {
  showResultModal.value = false
}

function openDebugPanelWithResult() {
  showDebugPanel.value = true
  closeResultModal()
}

function handleAssetOpened(asset: Asset) {
  if (asset.type === AssetType.BLUEPRINT) {
    // If we're already editing a function, go back to main editor
    if (currentEditingFunction.value) {
      backToMainEditor();
    }
    
    // Navigate to the editor with the selected blueprint ID
    router.push(`/editor/${asset.id}`);
  }
}

// Watch for critical errors to auto-show error panel
watch(() => errorStore.hasCriticalErrors, (hasCriticalErrors) => {
  if (hasCriticalErrors && errorViewStore.autoShowErrors) {
    showErrorPanel.value = true
  }
})

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
.editor-view {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 50px); /* Subtract header height */
  overflow: hidden;
}

/* The editor container is flexible and will adjust based on content browser height */
.editor-container {
  flex: 1;
  min-height: 0;
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
  align-items: center;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background-color: #444;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  margin-right: 12px;
}

.back-btn:hover {
  background-color: #555;
}

.editor-title {
  font-size: 1.1rem;
  font-weight: 500;
  color: var(--accent-blue);
}

.blueprint-name {
  font-size: 1.2rem;
  background-color: transparent;
  border: 1px solid transparent;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
}

.blueprint-name:hover, .blueprint-name:focus {
  border-color: #3d3d3d;
  background-color: #333;
  outline: none;
}

.tool-buttons {
  display: flex;
  gap: 8px;
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
  height: 30%;
  border-top: 1px solid #3d3d3d;
  background-color: #2d2d2d;
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

/* Error panel styles */
.error-panel-container {
  position: absolute;
  right: 0;
  bottom: 30px; /* Adjust to match BottomDrawer height */
  width: 400px;
  background-color: #252525;
  border-left: 1px solid #3d3d3d;
  border-top: 1px solid #3d3d3d;
  height: 300px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  box-shadow: -2px -2px 10px rgba(0, 0, 0, 0.2);
}

.error-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: #333;
  border-bottom: 1px solid #444;
}

.error-panel-header h3 {
  margin: 0;
  font-size: 1rem;
  color: #e0e0e0;
}

.error-panel-header .close-btn {
  background: none;
  border: none;
  color: #aaa;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
}

.error-panel-header .close-btn:hover {
  color: white;
}

.badge {
  background-color: #555;
  color: white;
  border-radius: 10px;
  padding: 1px 6px;
  font-size: 0.7rem;
  margin-left: 5px;
}

.badge.critical {
  background-color: #e74c3c;
}
</style>
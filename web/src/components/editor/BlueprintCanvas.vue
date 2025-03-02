<template>
  <div
      ref="canvasContainer"
      class="blueprint-canvas-container"
      @wheel="handleWheel"
      @mousedown="handleMouseDown"
      @contextmenu.prevent="handleContextMenu"
  >
    <div
        ref="canvas"
        class="blueprint-canvas"
        :style="canvasStyle"
    >
      <!-- Grid background is handled by CSS -->

      <!-- Connections -->
      <svg class="connections-layer">
        <g v-for="connection in connections" :key="connection.id">
          <path
              :d="getConnectionPath(connection)"
              :class="getConnectionClass(connection)"
              @click="handleConnectionClick(connection)"
          />
        </g>

        <!-- Connection being created -->
        <path
            v-if="isCreatingConnection"
            :d="temporaryConnectionPath"
            :class="{
            'connection-path': true,
            'connection-exec': isCreatingExecutionConnection,
            'connection-data': !isCreatingExecutionConnection,
            'connection-valid': isValidConnection,
            'connection-invalid': !isValidConnection
          }"
        />
      </svg>

      <!-- Nodes -->
      <BlueprintNode
          v-for="node in nodes"
          :key="node.id"
          :node="node"
          :node-type="getNodeType(node.type)"
          :status="getNodeStatus(node.id)"
          :selected="selectedNodeId === node.id"
          @select="handleNodeSelect"
          @deselect="handleNodeDeselect"
          @move="handleNodeMove"
          @pin-mouse-down="handlePinMouseDown"
          @pin-mouse-up="handlePinMouseUp"
          @delete="handleNodeDelete"
      />
    </div>

    <!-- Context menu -->
    <div v-if="showContextMenu" class="context-menu" :style="contextMenuStyle">
      <div class="context-menu-item" @click="handleAddNodeAtCursor">
        Add Node
      </div>
      <div class="context-menu-item" @click="handlePasteNode" v-if="canPaste">
        Paste
      </div>
      <div class="context-menu-divider"></div>
      <div class="context-menu-item" @click="handleCenterView">
        Center View
      </div>
      <div class="context-menu-item" @click="handleResetZoom">
        Reset Zoom
      </div>
    </div>

    <!-- Node context menu -->
    <div v-if="showNodeContextMenu" class="context-menu" :style="contextMenuStyle">
      <div class="context-menu-item" @click="handleCopyNode">
        Copy
      </div>
      <div class="context-menu-item" @click="handleDuplicateNode">
        Duplicate
      </div>
      <div class="context-menu-divider"></div>
      <div class="context-menu-item delete" @click="handleDeleteSelectedNode">
        Delete
      </div>
    </div>

    <!-- Add node menu -->
    <div v-if="showAddNodeMenu" class="add-node-menu" :style="contextMenuStyle">
      <div class="search-container">
        <input
            ref="addNodeSearchInput"
            v-model="addNodeSearchQuery"
            type="text"
            placeholder="Search nodes..."
            class="search-input"
        />
      </div>

      <div class="node-categories">
        <template v-for="category in filteredCategories" :key="category">
          <div class="category-header">{{ category }}</div>
          <div
              v-for="nodeType in nodeTypesInCategory(category)"
              :key="nodeType.typeId"
              class="node-type-item"
              @click="handleAddSpecificNode(nodeType)"
          >
            {{ nodeType.name }}
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { v4 as uuid } from 'uuid'
import { useNodeRegistryStore } from '../../stores/nodeRegistry'
import type { Node, Connection, Position } from '../../types/blueprint'
import type { NodeTypeDefinition } from '../../types/nodes'
import type { NodeExecutionStatus } from '../../types/execution'
import BlueprintNode from './BlueprintNode.vue'

const props = defineProps<{
  nodes: Node[]
  connections: Connection[]
  nodeStatuses?: Record<string, NodeExecutionStatus>
}>()

const emit = defineEmits<{
  (e: 'node-selected', nodeId: string): void
  (e: 'node-deselected'): void
  (e: 'node-moved', nodeId: string, position: Position): void
  (e: 'connection-created', connection: Connection): void
  (e: 'connection-deleted', connectionId: string): void
  (e: 'node-deleted', nodeId: string): void
}>()

const nodeRegistryStore = useNodeRegistryStore()

// DOM Refs
const canvasContainer = ref<HTMLElement | null>(null)
const canvas = ref<HTMLElement | null>(null)
const addNodeSearchInput = ref<HTMLInputElement | null>(null)

// Canvas state
const scale = ref(1)
const offset = ref({ x: 0, y: 0 })
const isDraggingCanvas = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const selectedNodeId = ref<string | null>(null)
const isPanning = ref(false)

// Context menu state
const showContextMenu = ref(false)
const showNodeContextMenu = ref(false)
const showAddNodeMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const addNodeSearchQuery = ref('')
const canPaste = ref(false)
const copiedNode = ref<Node | null>(null)

// Connection creation state
const isCreatingConnection = ref(false)
const connectionSource = ref<{
  nodeId: string;
  pinId: string;
  position: { x: number, y: number };
  isExecution: boolean;
} | null>(null)
const connectionTarget = ref<{ x: number, y: number } | null>(null)
const isValidConnection = ref(false)
const isCreatingExecutionConnection = ref(false)

// Computed properties
const canvasStyle = computed(() => {
  return {
    transform: `translate(${offset.value.x}px, ${offset.value.y}px) scale(${scale.value})`
  }
})

const contextMenuStyle = computed(() => {
  return {
    left: `${contextMenuPosition.value.x}px`,
    top: `${contextMenuPosition.value.y}px`
  }
})

const temporaryConnectionPath = computed(() => {
  if (!isCreatingConnection.value || !connectionSource.value || !connectionTarget.value) {
    return ''
  }

  return generateConnectionPath(
      connectionSource.value.position.x,
      connectionSource.value.position.y,
      connectionTarget.value.x,
      connectionTarget.value.y
  )
})

const filteredCategories = computed(() => {
  const categories = nodeRegistryStore.categories
  if (!addNodeSearchQuery.value) {
    return categories
  }

  const query = addNodeSearchQuery.value.toLowerCase()

  return categories.filter(category => {
    // Check if category matches
    if (category.toLowerCase().includes(query)) {
      return true
    }

    // Check if any node type in this category matches
    const nodeTypes = nodeRegistryStore.nodeTypesByCategory[category] || []
    return nodeTypes.some(nodeType =>
        nodeType.name.toLowerCase().includes(query) ||
        nodeType.description.toLowerCase().includes(query)
    )
  })
})

// Methods
function getNodeType(typeId: string): NodeTypeDefinition | null {
  return nodeRegistryStore.getNodeTypeById(typeId)
}

function getNodeStatus(nodeId: string): string {
  if (!props.nodeStatuses) return 'idle'

  const status = props.nodeStatuses[nodeId]
  return status ? status.status : 'idle'
}

function handleWheel(event: WheelEvent) {
  event.preventDefault()

  // Zoom logic
  const delta = -Math.sign(event.deltaY) * 0.1
  const newScale = Math.max(0.1, Math.min(2, scale.value + delta))

  // Get mouse position relative to the canvas container
  const rect = canvasContainer.value!.getBoundingClientRect()
  const mouseX = event.clientX - rect.left - offset.value.x
  const mouseY = event.clientY - rect.top - offset.value.y

  // Adjust offset to zoom towards the mouse position
  offset.value.x -= mouseX * (newScale / scale.value - 1)
  offset.value.y -= mouseY * (newScale / scale.value - 1)

  scale.value = newScale
}

function handleMouseDown(event: MouseEvent) {
  // Only start panning on middle-button or shift+left-button
  if (event.button === 1 || (event.button === 0 && event.shiftKey)) {
    isPanning.value = true
    dragStart.value = { x: event.clientX, y: event.clientY }
    canvasContainer.value!.style.cursor = 'grabbing'
    event.preventDefault()
  }

  // Close context menus on any click elsewhere
  if (!event.target || !(event.target as HTMLElement).closest('.context-menu') &&
      !(event.target as HTMLElement).closest('.add-node-menu')) {
    showContextMenu.value = false
    showNodeContextMenu.value = false
    showAddNodeMenu.value = false
  }

  // Only propagate the click to the canvas when not panning
  if (!isPanning.value) {
    // If clicking on the canvas background (not a node), deselect the current node
    if (event.target === canvas.value || event.target === canvasContainer.value) {
      selectedNodeId.value = null
      emit('node-deselected')
    }
  }
}

function handleMouseMove(event: MouseEvent) {
  if (isPanning.value) {
    const dx = event.clientX - dragStart.value.x
    const dy = event.clientY - dragStart.value.y

    offset.value.x += dx
    offset.value.y += dy

    dragStart.value = { x: event.clientX, y: event.clientY }

    event.preventDefault()
  }

  // Update temporary connection path while drawing a connection
  if (isCreatingConnection.value && connectionSource.value) {
    // Convert mouse position to canvas coordinates
    const rect = canvasContainer.value!.getBoundingClientRect()
    const canvasX = (event.clientX - rect.left - offset.value.x) / scale.value
    const canvasY = (event.clientY - rect.top - offset.value.y) / scale.value

    connectionTarget.value = { x: canvasX, y: canvasY }

    // Check if the mouse is over a valid target pin
    const targetPinElement = document.elementFromPoint(event.clientX, event.clientY) as HTMLElement | null

    // Reset valid state
    isValidConnection.value = false

    if (targetPinElement && targetPinElement.classList.contains('node-pin')) {
      const targetNodeId = targetPinElement.getAttribute('data-node-id')
      const targetPinId = targetPinElement.getAttribute('data-pin-id')
      const targetPinType = targetPinElement.getAttribute('data-pin-type')
      const isTargetInput = targetPinElement.classList.contains('pin-input')

      if (targetNodeId && targetPinId && targetPinType && isTargetInput) {
        // Don't allow connecting to the same node
        if (targetNodeId !== connectionSource.value.nodeId) {
          // Check if execution pin types match
          const isTargetExecution = targetPinType === 'execution'

          if (isTargetExecution === connectionSource.value.isExecution) {
            isValidConnection.value = true
          }
        }
      }
    }
  }
}

function handleMouseUp(event: MouseEvent) {
  if (isPanning.value) {
    isPanning.value = false
    canvasContainer.value!.style.cursor = 'default'
    event.preventDefault()
  }

  // Handle connection creation
  if (isCreatingConnection.value && connectionSource.value) {
    // Find target pin under the mouse
    const targetPinElement = document.elementFromPoint(event.clientX, event.clientY) as HTMLElement | null

    if (targetPinElement && targetPinElement.classList.contains('node-pin')) {
      const targetNodeId = targetPinElement.getAttribute('data-node-id')
      const targetPinId = targetPinElement.getAttribute('data-pin-id')
      const targetPinType = targetPinElement.getAttribute('data-pin-type')
      const isTargetInput = targetPinElement.classList.contains('pin-input')

      if (targetNodeId && targetPinId && targetPinType && isTargetInput) {
        // Don't allow connecting to the same node
        if (targetNodeId !== connectionSource.value.nodeId) {
          // Check if execution pin types match
          const isTargetExecution = targetPinType === 'execution'

          if (isTargetExecution === connectionSource.value.isExecution) {
            // Create the connection
            const connection: Connection = {
              id: uuid(),
              sourceNodeId: connectionSource.value.nodeId,
              sourcePinId: connectionSource.value.pinId,
              targetNodeId: targetNodeId,
              targetPinId: targetPinId,
              connectionType: isTargetExecution ? 'execution' : 'data'
            }

            emit('connection-created', connection)
          }
        }
      }
    }

    // Reset connection creation state
    isCreatingConnection.value = false
    connectionSource.value = null
    connectionTarget.value = null
  }
}

function handleContextMenu(event: MouseEvent) {
  // Check if we're right-clicking on a node
  const nodeElement = (event.target as HTMLElement).closest('.blueprint-node')

  if (nodeElement) {
    const nodeId = nodeElement.getAttribute('data-node-id')
    if (nodeId) {
      // Show node context menu
      contextMenuPosition.value = { x: event.clientX, y: event.clientY }
      showNodeContextMenu.value = true
      showContextMenu.value = false
      showAddNodeMenu.value = false
      selectedNodeId.value = nodeId
      emit('node-selected', nodeId)
    }
  } else {
    // Show canvas context menu
    contextMenuPosition.value = { x: event.clientX, y: event.clientY }
    showContextMenu.value = true
    showNodeContextMenu.value = false
    showAddNodeMenu.value = false

    // Store the cursor position for adding nodes
    const rect = canvasContainer.value!.getBoundingClientRect()

    // Convert to canvas coordinates
    const canvasX = (event.clientX - rect.left - offset.value.x) / scale.value
    const canvasY = (event.clientY - rect.top - offset.value.y) / scale.value

    // Store for later use
    dragStart.value = { x: canvasX, y: canvasY }

    // Check if we have copied node data
    canPaste.value = copiedNode.value !== null
  }
}

function handleNodeSelect(nodeId: string) {
  selectedNodeId.value = nodeId
  emit('node-selected', nodeId)
}

function handleNodeDeselect() {
  selectedNodeId.value = null
  emit('node-deselected')
}

function handleNodeMove(nodeId: string, position: Position) {
  emit('node-moved', nodeId, position)
}

function handlePinMouseDown(data: {
  nodeId: string,
  pinId: string,
  isInput: boolean,
  isExecution: boolean,
  position: { x: number, y: number }
}) {
  // Only start connection from output pins
  if (!data.isInput) {
    isCreatingConnection.value = true
    isCreatingExecutionConnection.value = data.isExecution
    connectionSource.value = {
      nodeId: data.nodeId,
      pinId: data.pinId,
      position: data.position,
      isExecution: data.isExecution
    }
    connectionTarget.value = { ...data.position }
  }
}

function handlePinMouseUp(data: {
  nodeId: string,
  pinId: string,
  isInput: boolean,
  isExecution: boolean
}) {
  // Only connect if we have a valid connection source and this is an input pin
  if (isCreatingConnection.value && connectionSource.value && data.isInput) {
    // Check if pins are compatible
    if (connectionSource.value.isExecution === data.isExecution) {
      // Create the connection
      const connection: Connection = {
        id: uuid(),
        sourceNodeId: connectionSource.value.nodeId,
        sourcePinId: connectionSource.value.pinId,
        targetNodeId: data.nodeId,
        targetPinId: data.pinId,
        connectionType: data.isExecution ? 'execution' : 'data'
      }

      emit('connection-created', connection)
    }

    // Reset connection creation state
    isCreatingConnection.value = false
    connectionSource.value = null
    connectionTarget.value = null
  }
}

function handleConnectionClick(connection: Connection) {
  // Delete the connection when clicking on it
  if (confirm('Delete this connection?')) {
    emit('connection-deleted', connection.id)
  }
}

function handleNodeDelete(nodeId: string) {
  emit('node-deleted', nodeId)
}

function handleAddNodeAtCursor() {
  showContextMenu.value = false
  showAddNodeMenu.value = true

  // Focus the search input
  setTimeout(() => {
    if (addNodeSearchInput.value) {
      addNodeSearchInput.value.focus()
    }
  }, 0)
}

function nodeTypesInCategory(category: string) {
  const query = addNodeSearchQuery.value.toLowerCase()
  const nodeTypes = nodeRegistryStore.nodeTypesByCategory[category] || []

  if (!query) {
    return nodeTypes
  }

  return nodeTypes.filter(nodeType =>
      nodeType.name.toLowerCase().includes(query) ||
      nodeType.description.toLowerCase().includes(query)
  )
}

function handleAddSpecificNode(nodeType: NodeTypeDefinition) {
  // Create a new node
  const node: Node = {
    id: uuid(),
    type: nodeType.typeId,
    position: { x: dragStart.value.x, y: dragStart.value.y },
    properties: []
  }

  // Add the node to the blueprint
  const newNode = structuredClone(node)
  props.nodes.push(newNode)

  // Close the menu
  showAddNodeMenu.value = false
}

function handleCopyNode() {
  if (selectedNodeId.value) {
    const node = props.nodes.find(n => n.id === selectedNodeId.value)
    if (node) {
      copiedNode.value = structuredClone(node)
    }
  }

  showNodeContextMenu.value = false
}

function handlePasteNode() {
  if (copiedNode.value) {
    // Create a new node based on the copied one
    const node: Node = {
      ...structuredClone(copiedNode.value),
      id: uuid(),
      position: { x: dragStart.value.x, y: dragStart.value.y }
    }

    // Add the node to the blueprint
    props.nodes.push(node)
  }

  showContextMenu.value = false
}

function handleDuplicateNode() {
  if (selectedNodeId.value) {
    const node = props.nodes.find(n => n.id === selectedNodeId.value)
    if (node) {
      // Create a new node based on the selected one
      const newNode: Node = {
        ...structuredClone(node),
        id: uuid(),
        position: {
          x: node.position.x + 20,
          y: node.position.y + 20
        }
      }

      // Add the node to the blueprint
      props.nodes.push(newNode)
    }
  }

  showNodeContextMenu.value = false
}

function handleDeleteSelectedNode() {
  if (selectedNodeId.value) {
    emit('node-deleted', selectedNodeId.value)
  }

  showNodeContextMenu.value = false
}

function handleCenterView() {
  if (canvasContainer.value && canvas.value) {
    // Calculate the average position of all nodes
    if (props.nodes.length > 0) {
      const avgX = props.nodes.reduce((sum, node) => sum + node.position.x, 0) / props.nodes.length
      const avgY = props.nodes.reduce((sum, node) => sum + node.position.y, 0) / props.nodes.length

      // Center the view on this position
      const containerWidth = canvasContainer.value.clientWidth
      const containerHeight = canvasContainer.value.clientHeight

      offset.value.x = containerWidth / 2 - avgX * scale.value
      offset.value.y = containerHeight / 2 - avgY * scale.value
    } else {
      // If no nodes, just center the view
      offset.value.x = 0
      offset.value.y = 0
    }
  }

  showContextMenu.value = false
}

function handleResetZoom() {
  scale.value = 1
  showContextMenu.value = false
}

function getConnectionPath(connection: Connection): string {
  // Find source and target nodes
  const sourceNode = props.nodes.find(n => n.id === connection.sourceNodeId)
  const targetNode = props.nodes.find(n => n.id === connection.targetNodeId)

  if (!sourceNode || !targetNode) {
    return ''
  }

  // Find pin elements
  const sourceElement = document.querySelector(
      `.node-pin.pin-output[data-node-id="${connection.sourceNodeId}"][data-pin-id="${connection.sourcePinId}"]`
  ) as HTMLElement | null

  const targetElement = document.querySelector(
      `.node-pin.pin-input[data-node-id="${connection.targetNodeId}"][data-pin-id="${connection.targetPinId}"]`
  ) as HTMLElement | null

  if (!sourceElement || !targetElement) {
    return ''
  }

  // Get positions of pins
  const sourceRect = sourceElement.getBoundingClientRect()
  const targetRect = targetElement.getBoundingClientRect()
  const canvasRect = canvas.value!.getBoundingClientRect()

  // Calculate positions relative to canvas
  const x1 = (sourceRect.left + sourceRect.width / 2 - canvasRect.left) / scale.value
  const y1 = (sourceRect.top + sourceRect.height / 2 - canvasRect.top) / scale.value
  const x2 = (targetRect.left + targetRect.width / 2 - canvasRect.left) / scale.value
  const y2 = (targetRect.top + targetRect.height / 2 - canvasRect.top) / scale.value

  // Generate path
  return generateConnectionPath(x1, y1, x2, y2)
}

function generateConnectionPath(x1: number, y1: number, x2: number, y2: number): string {
  // Create a bezier curve path
  const dx = Math.max(Math.abs(x2 - x1) * 0.5, 50)

  return `M ${x1} ${y1} C ${x1 + dx} ${y1}, ${x2 - dx} ${y2}, ${x2} ${y2}`
}

function getConnectionClass(connection: Connection): string {
  const classes = ['connection-path']

  if (connection.connectionType === 'execution') {
    classes.push('connection-exec')
  } else {
    classes.push('connection-data')
  }

  return classes.join(' ')
}

// Event listeners
onMounted(() => {
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)

  // Center the view
  if (canvasContainer.value && canvas.value) {
    offset.value.x = canvasContainer.value.clientWidth / 2
    offset.value.y = canvasContainer.value.clientHeight / 2
  }
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
})

// Expose methods for parent components
defineExpose({
  centerView: handleCenterView,
  resetZoom: handleResetZoom
})
</script>

<style scoped>
.blueprint-canvas-container {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--bg-color);
}

.blueprint-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 10000px;
  height: 10000px;
  transform-origin: 0 0;
  /* Grid background */
  background-image:
      linear-gradient(to right, var(--grid-color) 1px, transparent 1px),
      linear-gradient(to bottom, var(--grid-color) 1px, transparent 1px);
  background-size: 20px 20px;
  background-position: 0 0;
}

.connections-layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.connection-path {
  fill: none;
  stroke-width: 2px;
  pointer-events: stroke;
}

.connection-exec {
  stroke: var(--conn-exec);
  stroke-dasharray: 4 2;
}

.connection-data {
  stroke: var(--conn-color);
}

.connection-valid {
  opacity: 1;
}

.connection-invalid {
  opacity: 0.5;
  stroke-dasharray: 4 2;
}

.context-menu, .add-node-menu {
  position: fixed;
  background-color: var(--context-menu-bg);
  border-radius: 5px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  padding: 5px 0;
  z-index: 1000;
  min-width: 120px;
  animation: fadeIn 0.15s ease-out;
}

.add-node-menu {
  width: 300px;
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.context-menu-item {
  padding: 8px 15px;
  cursor: pointer;
  transition: background-color 0.2s;
  color: var(--text-color);
}

.context-menu-item:hover {
  background-color: var(--context-menu-hover);
}

.context-menu-divider {
  height: 1px;
  background-color: #444;
  margin: 5px 0;
}

.context-menu-item.delete {
  color: #f44336;
}

.search-container {
  padding: 8px;
  border-bottom: 1px solid #444;
}

.search-input {
  width: 100%;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #444;
  background-color: #333;
  color: white;
}

.node-categories {
  overflow-y: auto;
  max-height: 350px;
}

.category-header {
  padding: 8px;
  background-color: #333;
  font-weight: 500;
  color: #aaa;
}

.node-type-item {
  padding: 8px 15px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.node-type-item:hover {
  background-color: var(--context-menu-hover);
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
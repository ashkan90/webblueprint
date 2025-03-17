<template>
  <div
      key="canvasContainer"
      ref="canvasContainer"
      class="blueprint-canvas-container"
      @wheel="handleWheel"
      @mousedown="handleMouseDown"
      @contextmenu.prevent="handleContextMenu"
      @dragover.prevent="handleDragOver"
      @drop.prevent="handleDrop"
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
              :class="[getConnectionClass(connection),{ 'connection-active': isConnectionActive(connection) }]"
              @click="handleConnectionClick(connection)"
              @mouseover="handleConnectionMouseOver($event, connection)"
              @mouseout="handleConnectionMouseOut"
          />

          <!-- Data flow animation -->
          <circle
              v-if="connection.connectionType === 'data' && isConnectionActive(connection)"
              :class="['data-particle', getConnectionParticleClass(connection), 'particle-active']"
              r="4"
          >
            <animateMotion
                :path="getConnectionPath(connection)"
                dur="0.5s"
                repeatCount="indefinite"
                rotate="auto"
            />
          </circle>
        </g>

        <!-- Custom data flow animations -->
        <g v-for="flow in dataFlows" :key="flow.id">
          <circle
              :class="['data-flow', `data-flow-${flow.sourceType}`]"
              r="5"
          >
            <animateMotion
                :path="flow.path"
                :keyPoints="`0;${flow.progress}`"
                dur="1s"
                keyTimes="0;1"
                calcMode="linear"
                fill="freeze"
            />
          </circle>
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
          :active-pins="activePins"
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

    <!-- Connection Feedback -->
    <ConnectionFeedback
        :show="showConnectionFeedback"
        :position="connectionFeedbackPosition"
        :result="connectionValidationResult"
    />

    <!-- Connection Tooltip -->
    <ConnectionValueTooltip
        :show="showConnectionTooltip"
        :position="connectionTooltipPosition"
        :pin-name="tooltipPinName"
        :pin-type="tooltipPinType"
        :pin-type-name="tooltipPinTypeName"
        :value="tooltipValue"
    />
  </div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, onUnmounted, watch, toRaw} from 'vue'
import { v4 as uuid } from 'uuid'
import { validateConnection } from '../../utils/connectionValidator';
import { useNodeRegistryStore } from '../../stores/nodeRegistry'
import type {Node, Connection, Position, NodeProperty} from '../../types/blueprint'
import type {NodePropertyDefinition, NodeTypeDefinition, PinDefinition} from '../../types/nodes'
import type { NodeExecutionStatus } from '../../types/execution'
import BlueprintNode from './BlueprintNode.vue'
import ConnectionValueTooltip from './ConnectionValueTooltip.vue';
import ConnectionFeedback from './ConnectionFeedback.vue';
import {useExecutionStore} from "../../stores/execution";


const props = defineProps<{
  nodes: Node[]
  connections: Connection[]
  nodeStatuses?: Record<string, NodeExecutionStatus>
}>()

const emit = defineEmits<{
  (e: 'node-selected', nodeId: string): void
  (e: 'node-added', nodeId: string): void
  (e: 'node-added', node: Node): void
  (e: 'node-deselected'): void
  (e: 'node-moved', nodeId: string, position: Position): void
  (e: 'connection-created', connection: Connection): void
  (e: 'connection-deleted', connectionId: string): void
  (e: 'node-deleted', nodeId: string): void
}>()

const nodeRegistryStore = useNodeRegistryStore()
const executionStore = useExecutionStore()

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

// Visual debugging
const showConnectionFeedback = ref(false);
const connectionFeedbackPosition = ref({ x: 0, y: 0 });
const connectionValidationResult = ref({ valid: false });
const potentialTargetNode = ref<string | null>(null);
const potentialTargetPin = ref<string | null>(null);
const showConnectionTooltip = ref(false);
const connectionTooltipPosition = ref({ x: 0, y: 0 });
const tooltipPinName = ref('');
const tooltipPinType = ref('');
const tooltipPinTypeName = ref('');
const tooltipValue = ref<any>(undefined);


// Context menu state
const showContextMenu = ref(false)
const showNodeContextMenu = ref(false)
const showAddNodeMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const addNodeSearchQuery = ref('')
const canPaste = ref(false)
const copiedNode = ref<Node | null>(null)

// Animation state
const activeConnections = ref<Set<string>>(new Set());
const activeNodes = ref<Set<string>>(new Set());
const activePins = ref<Set<string>>(new Set());
const dataFlows = ref<Array<{
  id: string;
  path: string;
  sourceType: string;
  progress: number;
  animationId: number;
}>>([]);

// Connection creation state
const isCreatingConnection = ref(false)
const connectionSource = ref<{
  nodeId: string;
  pinId: string;
  position: { x: number, y: number };
  isExecution: boolean;
  defaultValue?: any;
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
  // First try to get the node type directly
  const nodeType = nodeRegistryStore.getNodeTypeById(typeId);
  if (nodeType) {
    return nodeType;
  }
  
  // If not found, it might be a user-defined function with a different ID format
  // Try to find it by name in the registry
  console.log('Node type not found directly:', typeId);
  
  // Log all available node types for debugging
  console.log('Available node types:', Object.keys(nodeRegistryStore.nodeTypes));
  
  // Check if this is a node created from a dragged function
  // If it's a function node, its id in the node registry would be the function name
  const nodeTypeValues = Object.values(nodeRegistryStore.nodeTypes);
  const matchingNodeType = nodeTypeValues.find(nt => 
    nt.name === typeId || // Match by name
    nt.typeId === typeId  // Match by typeId
  );
  
  if (matchingNodeType) {
    console.log('Found matching node type by name:', matchingNodeType);
    return matchingNodeType;
  }
  
  return null;
}

function getNodeStatus(nodeId: string): string {
  // if (!props.nodeStatuses) return 'idle'
  //
  // const status = props.nodeStatuses[nodeId]
  // return status ? status.status : 'idle'
  if (!props.nodeStatuses) return 'idle';

  const status = props.nodeStatuses[nodeId];
  const statusStr = status ? status.status : 'idle';

  // Update active nodes tracking
  if (statusStr === 'executing') {
    activeNodes.value.add(nodeId);
  } else {
    activeNodes.value.delete(nodeId);
  }

  return statusStr;
}

// Function to handle hovering over connections
function handleConnectionMouseOver(event: MouseEvent, connection: Connection) {
  // Only show tooltips during execution when there's debug data
  if (!props.nodeStatuses || Object.keys(props.nodeStatuses).length === 0) {
    return;
  }

  // Get the connection data
  const sourceNode = props.nodes.find(n => n.id === connection.sourceNodeId);
  const sourceNodeType = sourceNode ? nodeRegistryStore.getNodeTypeById(sourceNode.type) : null;
  const sourcePin = sourceNodeType?.outputs.find(p => p.id === connection.sourcePinId);

  if (!sourceNode || !sourceNodeType || !sourcePin) {
    return;
  }

  // Try to get the debug data for this node
  const debugData = executionStore.getNodeDebugData(connection.sourceNodeId);

  // If we have debug data, show the tooltip
  if (debugData && debugData.outputs) {
    // Get the output value
    const value = debugData.outputs[connection.sourcePinId];

    if (value !== undefined) {
      // Show the tooltip
      showConnectionTooltip.value = true;
      connectionTooltipPosition.value = {
        x: event.clientX,
        y: event.clientY
      };

      // Set tooltip data
      tooltipPinName.value = sourcePin.name;
      tooltipPinType.value = sourcePin.type.id;
      tooltipPinTypeName.value = sourcePin.type.name;
      tooltipValue.value = value;
    }
  }

  showConnectionTooltip.value = false;
}

function handleConnectionMouseOut() {
  showConnectionTooltip.value = false;
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
    isValidConnection.value = false;
    showConnectionFeedback.value = false;

    // Reset potential target tracking
    potentialTargetNode.value = null;
    potentialTargetPin.value = null;

    if (targetPinElement && targetPinElement.classList.contains('node-pin')) {
      const targetNodeId = targetPinElement.getAttribute('data-node-id')
      const targetPinId = targetPinElement.getAttribute('data-pin-id')
      const targetPinType = targetPinElement.getAttribute('data-pin-type')
      const isTargetInput = targetPinElement.classList.contains('pin-input')

      if (targetNodeId && targetPinId && targetPinType && isTargetInput) {
        // Store the potential target
        potentialTargetNode.value = targetNodeId;
        potentialTargetPin.value = targetPinId;

        // Don't allow connecting to the same node
        if (targetNodeId !== connectionSource.value.nodeId) {
          // Check connection validity with our validation utility
          const validationResult = validateConnection(
              connectionSource.value.nodeId,
              connectionSource.value.pinId,
              targetNodeId,
              targetPinId
          );

          // Update validation state
          isValidConnection.value = validationResult.valid;
          connectionValidationResult.value = validationResult;

          // Show feedback tooltip near the mouse position
          showConnectionFeedback.value = true;
          connectionFeedbackPosition.value = {
            x: event.clientX,
            y: event.clientY
          };
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
    let connectionCreated = false;

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
          // Use our validation function
          const validationResult = validateConnection(
              connectionSource.value.nodeId,
              connectionSource.value.pinId,
              targetNodeId,
              targetPinId
          );

          if (validationResult.valid) {
            // Create the connection
            const connection: Connection = {
              id: uuid(),
              sourceNodeId: connectionSource.value.nodeId,
              sourcePinId: connectionSource.value.pinId,
              targetNodeId: targetNodeId,
              targetPinId: targetPinId,
              connectionType: targetPinType === 'execution' ? 'execution' : 'data',
              // Include any default value in the connection metadata
              data: connectionSource.value.defaultValue !== undefined ? {
                defaultValue: connectionSource.value.defaultValue
              } : undefined
            }

            emit('connection-created', connection);
            connectionCreated = true;
          } else {
            // Show validation error feedback
            showConnectionFeedback.value = true;
            connectionFeedbackPosition.value = {
              x: event.clientX,
              y: event.clientY
            };
            connectionValidationResult.value = validationResult;

            // Hide feedback after a short delay
            setTimeout(() => {
              showConnectionFeedback.value = false;
            }, 3000);
          }
        }
      }
    }

    // Reset connection creation state
    if (connectionCreated || !showConnectionFeedback.value) {
      isCreatingConnection.value = false;
      connectionSource.value = null;
      connectionTarget.value = null;
      showConnectionFeedback.value = false;
    }
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

function handleDragOver(event: DragEvent) {
  // Allow the drop
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy';
  }
}

function handleDrop(event: DragEvent) {
  // Get the dropped data
  if (!event.dataTransfer) return;

  // Completely ignore variable drags - let BlueprintLeftPanel handle them
  if (event.dataTransfer.types.includes('application/x-blueprint-variable')) {
    // This is a variable drag - don't do anything
    console.log("Variable drag detected - ignoring in canvas");
    return;
  }

  const jsonData = event.dataTransfer.getData('application/json');
  if (!jsonData) {
    console.error('No valid data found in drag operation');
    return;
  }

  try {
    const nodeData: Node = JSON.parse(jsonData);

    console.log('Dropped node data:', nodeData);

    // Convert to canvas coordinates
    const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
    const canvasX = (event.clientX - rect.left - offset.value.x) / scale.value;
    const canvasY = (event.clientY - rect.top - offset.value.y) / scale.value;

    // Update position
    nodeData.position = { x: canvasX, y: canvasY };

    // Ensure we have a complete node object
    if (!nodeData.id || !nodeData.type) {
      console.error('Invalid node data', nodeData);
      return;
    }

    // Ensure properties is defined
    if (!nodeData.properties) {
      nodeData.properties = [];
    }

    // Add the node
    emit('node-added', nodeData);

    // Select the new node
    if (selectedNodeId) {
      selectedNodeId.value = nodeData.id;
    }
  } catch (error) {
    console.error('Error handling drop:', error);
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
  position: { x: number, y: number },
  defaultValue?: any
}) {
  // Only start connection from output pins
  if (!data.isInput) {
    isCreatingConnection.value = true
    isCreatingExecutionConnection.value = data.isExecution
    connectionSource.value = {
      nodeId: data.nodeId,
      pinId: data.pinId,
      position: data.position,
      isExecution: data.isExecution,
      defaultValue: data.defaultValue
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

  if (nodeType.properties) {
    nodeType.properties.forEach((prop) => {
      node.properties.push({
        name: prop.name,
        displayName: prop.displayName,
        value: prop.value,
      })
    })
  }

  nodeType.inputs.forEach((input: PinDefinition) => {
    if (!input.default) {
      return
    }

    const nodeProperty: NodeProperty = {
      name: `input_${input.id}`,
      value: input.default,
    }

    node.properties.push(nodeProperty)
  })

  // Add the node to the blueprint
  const newNode = structuredClone(node)
  emit('node-added', newNode)
  // props.nodes.push(newNode)

  // Close the menu
  showAddNodeMenu.value = false
}

function handleCopyNode() {
  if (selectedNodeId.value) {
    const node = props.nodes.find(n => n.id === selectedNodeId.value)
    if (node) {
      copiedNode.value = structuredClone(toRaw(node))
    }
  }

  showNodeContextMenu.value = false
}

function handlePasteNode() {
  if (copiedNode.value) {
    // Create a new node based on the copied one
    const node: Node = {
      ...structuredClone(toRaw(copiedNode.value)),
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
        ...structuredClone(toRaw(node)),
        id: uuid(),
        position: {
          x: node.position.x + 20,
          y: node.position.y + 20
        }
      }

      // Add the node to the blueprint
      emit('node-added', newNode)
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

function isConnectionActive(connection: Connection): boolean {
  return activeConnections.value.has(connection.id);
  // Check if this connection is part of an active data flow
  // if (!props.nodeStatuses) return false
  //
  // // For execution connections, check if source node is executing or completed
  // // and target node is executing
  // if (connection.connectionType === 'execution') {
  //   const sourceStatus = props.nodeStatuses[connection.sourceNodeId]
  //   const targetStatus = props.nodeStatuses[connection.targetNodeId]
  //
  //   return (sourceStatus?.status === 'completed' && targetStatus?.status === 'executing')
  // }
  //
  // // For data connections, check if source node is completed and target is executing
  // const sourceStatus = props.nodeStatuses[connection.sourceNodeId]
  // const targetStatus = props.nodeStatuses[connection.targetNodeId]
  //
  // return (sourceStatus && targetStatus &&
  //     (sourceStatus.status === 'completed' || sourceStatus.status === 'executing') &&
  //     targetStatus.status === 'executing')
}

function getConnectionParticleClass(connection: Connection): string {
  // Get the source node type to determine the particle color
  const sourceNode = props.nodes.find(n => n.id === connection.sourceNodeId)
  if (!sourceNode) return ''

  // Get the pin that's the source of this connection
  const nodeType = nodeRegistryStore.getNodeTypeById(sourceNode.type)
  if (!nodeType) return ''

  const pin = nodeType.outputs.find(p => p.id === connection.sourcePinId)
  if (!pin) return ''

  // Return a class based on pin type
  switch (pin.type.id) {
    case 'string': return 'particle-string'
    case 'number': return 'particle-number'
    case 'boolean': return 'particle-boolean'
    case 'object': return 'particle-object'
    case 'array': return 'particle-array'
    default: return 'particle-any'
  }
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

// animation functions
// New function to animate data flowing through connections
function animateDataFlow(connection: Connection, sourceType: string) {
  // Get the SVG path for this connection
  const pathData = getConnectionPath(connection);

  // Create a unique ID for this animation
  const flowId = `flow-${connection.id}-${Date.now()}`;

  // Create animation data
  const flowData = {
    id: flowId,
    path: pathData,
    sourceType,
    progress: 0,
    animationId: 0
  };

  // Add to data flows array
  dataFlows.value.push(flowData);

  // Start animation
  let progress = 0;
  const animate = () => {
    progress += 0.02; // Increment progress (0.0 to 1.0)

    // Update the flow data
    const flow = dataFlows.value.find(f => f.id === flowId);
    if (flow) {
      flow.progress = progress;
    }

    if (progress >= 1) {
      // Animation complete, remove the flow
      dataFlows.value = dataFlows.value.filter(f => f.id !== flowId);
    } else {
      // Continue animation
      flowData.animationId = requestAnimationFrame(animate);
    }
  };

  // Start the animation
  flowData.animationId = requestAnimationFrame(animate);
}

// Function to get path position at a certain percentage
function getPointAlongPath(path: SVGPathElement, percent: number) {
  const length = path.getTotalLength();
  return path.getPointAtLength(length * percent);
}

function updateActivePins(nodeId: string, status: string) {
  const node = props.nodes.find(n => n.id === nodeId);
  if (!node) return;

  const nodeType = nodeRegistryStore.getNodeTypeById(node.type);
  if (!nodeType) return;

  if (status === 'executing') {
    // Mark all input pins as active
    nodeType.inputs.forEach(pin => {
      activePins.value.add(`${nodeId}-${pin.id}`);
    });
  } else {
    // Remove all pins for this node
    const toRemove: string[] = [];
    activePins.value.forEach(id => {
      if (id.startsWith(`${nodeId}-`)) {
        toRemove.push(id);
      }
    });

    toRemove.forEach(id => {
      activePins.value.delete(id);
    });

    if (status === 'completed') {
      // Mark all output pins as briefly active
      nodeType.outputs.forEach(pin => {
        const pinId = `${nodeId}-${pin.id}`;
        activePins.value.add(pinId);

        // Remove after a delay
        setTimeout(() => {
          activePins.value.delete(pinId);
        }, 1000);
      });
    }
  }
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

// Watch for node status changes
watch(() => props.nodeStatuses, async (newStatuses, oldStatuses) => {
  if (!newStatuses) return;

  // Check for new active nodes
  for (const [nodeId, status] of Object.entries(newStatuses)) {
    if (status.status === 'executing') {
      updateActivePins(nodeId, 'executing');
    } else if (status.status === 'completed') {
      updateActivePins(nodeId, 'completed');
      setTimeout(() => {
        props.nodeStatuses[nodeId].status = 'idle'
      }, 2500)
    }
  }
}, { deep: true });

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

.connection-active {
  filter: drop-shadow(0 0 3px rgba(255, 255, 255, 0.7));
  stroke-width: 3px;
  animation: connection-pulse 1s ease-in-out infinite;
}

@keyframes connection-pulse {
  0% { opacity: 0.6; }
  50% { opacity: 1; }
  100% { opacity: 0.6; }
}

.data-particle {
  fill: white;
  filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.8));
}

.particle-string {
  fill: var(--input-pin);
}

.particle-number {
  fill: var(--output-pin);
}

.particle-boolean {
  fill: #dc5050;
}

.particle-object {
  fill: var(--conn-color);
}

.particle-array {
  fill: #bb86fc;
}

.particle-any {
  fill: #aaaaaa;
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

.node-active {
  box-shadow: 0 0 0 2px var(--accent-yellow), 0 0 15px rgba(255, 204, 0, 0.5);
  animation: node-active 1.5s ease-in-out infinite;
}

.node-completed {
  animation: node-completed 0.7s ease-out;
}

.node-error {
  animation: node-error 0.5s ease-in-out 3;
}

@keyframes node-active {
  0% { box-shadow: 0 0 0 2px var(--accent-yellow), 0 0 10px rgba(255, 204, 0, 0.3); }
  50% { box-shadow: 0 0 0 2px var(--accent-yellow), 0 0 20px rgba(255, 204, 0, 0.8); }
  100% { box-shadow: 0 0 0 2px var(--accent-yellow), 0 0 10px rgba(255, 204, 0, 0.3); }
}

@keyframes node-completed {
  0% { box-shadow: 0 0 0 2px var(--accent-green), 0 0 15px rgba(76, 175, 80, 0.5); }
  100% { box-shadow: 0 0 0 2px var(--accent-green), 0 0 5px rgba(76, 175, 80, 0.3); opacity: 1; }
}

@keyframes node-error {
  0% { box-shadow: 0 0 0 2px var(--accent-red), 0 0 10px rgba(244, 67, 54, 0.5); }
  50% { box-shadow: 0 0 0 2px var(--accent-red), 0 0 20px rgba(244, 67, 54, 0.8); }
  100% { box-shadow: 0 0 0 2px var(--accent-red), 0 0 10px rgba(244, 67, 54, 0.5); }
}

/* Improved connection animations */
.connection-active {
  filter: drop-shadow(0 0 3px rgba(255, 255, 255, 0.7));
  stroke-width: 3px;
  animation: connection-pulse 1s ease-in-out infinite;
}

.connection-exec.connection-active {
  stroke-dasharray: none;
  stroke: var(--accent-yellow);
  filter: drop-shadow(0 0 5px rgba(255, 204, 0, 0.7));
}

.connection-data.connection-active {
  stroke: var(--accent-blue);
  filter: drop-shadow(0 0 5px rgba(52, 152, 219, 0.7));
}

@keyframes connection-pulse {
  0% { opacity: 0.7; stroke-width: 2.5px; }
  50% { opacity: 1; stroke-width: 4px; }
  100% { opacity: 0.7; stroke-width: 2.5px; }
}

/* Enhanced data particles */
.data-particle {
  fill: white;
  filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.8));
  r: 4;
}

.particle-active {
  animation: particle-pulse 1s ease-in-out infinite;
}

@keyframes particle-pulse {
  0% { r: 3; filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.5)); }
  50% { r: 5; filter: drop-shadow(0 0 4px rgba(255, 255, 255, 0.9)); }
  100% { r: 3; filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.5)); }
}

/* Pin state indicators */
.pin-input.pin-active::after {
  content: '';
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.3);
  animation: pin-glow 1s ease-in-out infinite;
}

.pin-output.pin-active::after {
  content: '';
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.3);
  animation: pin-glow 1s ease-in-out infinite;
}

@keyframes pin-glow {
  0% { transform: scale(1); opacity: 0.3; }
  50% { transform: scale(1.5); opacity: 0.6; }
  100% { transform: scale(1); opacity: 0.3; }
}

/* Execution path visualization */
.execution-path {
  stroke: var(--accent-yellow);
  stroke-width: 3px;
  stroke-dasharray: 5 3;
  animation: dash-animation 1s linear infinite;
  opacity: 0.8;
  stroke-linecap: round;
}

@keyframes dash-animation {
  to {
    stroke-dashoffset: -8;
  }
}

/* Data flow visualization */
.data-flow {
  position: absolute;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  z-index: 10;
  pointer-events: none;
  transform: translate(-50%, -50%);
  filter: drop-shadow(0 0 3px rgba(255, 255, 255, 0.6));
}

.data-flow-string {
  background-color: var(--input-pin);
}

.data-flow-number {
  background-color: var(--output-pin);
}

.data-flow-boolean {
  background-color: #dc5050;
}

.data-flow-object {
  background-color: var(--conn-color);
}

.data-flow-array {
  background-color: #bb86fc;
}

.data-flow-any {
  background-color: #aaaaaa;
}
</style>
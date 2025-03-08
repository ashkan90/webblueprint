<template>
  <div
      ref="nodeElement"
      class="blueprint-node"
      :class="[status, { 'selected': selected, 'data-node': isDataOnlyNode }]"
      :style="nodeStyle"
      :data-node-id="node.id"
      @mousedown="handleMouseDown"
  >
    <div class="node-header" @mousedown="handleHeaderMouseDown">
      <div class="node-title">{{ nodeTitle }}</div>
      <div v-if="status !== 'idle'" class="node-status-badge">{{ status }}</div>
    </div>

    <div class="node-content">
      <!-- Execution Input Pins - Do not show for data-only nodes -->
      <div v-if="hasExecInputs && !(isInFunction && nodeType.category === 'Function') && !isDataOnlyNode" class="pin-section">
        <div
            v-for="pin in execInputPins"
            :key="pin.id"
            class="pin-row"
        >
          <div
              class="node-pin pin-input pin-exec"
              :class="{ 'pin-active': isPinActive(pin.id) }"
              :data-node-id="node.id"
              :data-pin-id="pin.id"
              :data-pin-type="pin.type.id"
              @mousedown.stop="handlePinMouseDown($event, pin, true, true)"
              @mouseup.stop="handlePinMouseUp($event, pin, true, true)"
          >
            <div class="pin-triangle"></div>
          </div>
          <div class="pin-label">{{ pin.name }}</div>
        </div>
      </div>

      <!-- Data Input Pins -->
      <div v-if="hasDataInputs" class="pin-section">
        <div
            v-for="pin in dataInputPins"
            :key="pin.id"
            class="pin-row"
        >
          <div
              class="node-pin pin-input"
              :class="{ 'pin-active': isPinActive(pin.id) }"
              :data-node-id="node.id"
              :data-pin-id="pin.id"
              :data-pin-type="pin.type.id"
              @mousedown.stop="handlePinMouseDown($event, pin, true, false)"
              @mouseup.stop="handlePinMouseUp($event, pin, true, false)"
          >
            <div class="pin-circle" :style="{ backgroundColor: getPinColor(pin) }"></div>
          </div>
          <div class="pin-label">{{ pin.name }}</div>
        </div>
      </div>

      <!-- Divider if both inputs and outputs and not a data-only node -->
      <div v-if="(hasDataInputs || (hasExecInputs && !isDataOnlyNode)) && (hasDataOutputs || (hasExecOutputs && !isDataOnlyNode)) && (execInputPins.length > 0 && !isInFunction && !isDataOnlyNode)" class="pin-divider"></div>

      <!-- Data Output Pins -->
      <div v-if="hasDataOutputs" class="pin-section">
        <div
            v-for="pin in dataOutputPins"
            :key="pin.id"
            class="pin-row pin-row-output"
        >
          <div class="pin-label">{{ pin.name }}</div>
          <div
              class="node-pin pin-output"
              :data-node-id="node.id"
              :data-pin-id="pin.id"
              :data-pin-type="pin.type.id"
              @mousedown.stop="handlePinMouseDown($event, pin, false, false)"
              @mouseup.stop="handlePinMouseUp($event, pin, false, false)"
          >
            <div class="pin-circle" :style="{ backgroundColor: getPinColor(pin) }"></div>
          </div>
        </div>
      </div>

      <!-- Execution Output Pins - Do not show for data-only nodes -->
      <div v-if="hasExecOutputs && !isDataOnlyNode" class="pin-section">
        <div
            v-for="pin in execOutputPins"
            :key="pin.id"
            class="pin-row pin-row-output"
        >
          <div class="pin-label">{{ pin.name }}</div>
          <div
              class="node-pin pin-output pin-exec"
              :data-node-id="node.id"
              :data-pin-id="pin.id"
              :data-pin-type="pin.type.id"
              @mousedown.stop="handlePinMouseDown($event, pin, false, true)"
              @mouseup.stop="handlePinMouseUp($event, pin, false, true)"
          >
            <div class="pin-triangle"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Node } from '../../types/blueprint'
import type { NodeTypeDefinition, PinDefinition } from '../../types/nodes'
import { useBlueprintStore } from '../../stores/blueprint'

const blueprintStore = useBlueprintStore()

const props = defineProps<{
  node: Node
  nodeType: NodeTypeDefinition | null
  activePins: Set<string>
  selected: boolean
  status: string
}>()

const emit = defineEmits<{
  (e: 'select', nodeId: string): void
  (e: 'deselect'): void
  (e: 'move', nodeId: string, position: { x: number, y: number }): void
  (e: 'pin-mouse-down', data: {
    nodeId: string,
    pinId: string,
    isInput: boolean,
    isExecution: boolean,
    position: { x: number, y: number },
    defaultValue?: any
  }): void
  (e: 'pin-mouse-up', data: {
    nodeId: string,
    pinId: string,
    isInput: boolean,
    isExecution: boolean
  }): void
  (e: 'delete', nodeId: string): void
}>()

// DOM Refs
const nodeElement = ref<HTMLElement | null>(null)

// State
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const initialPosition = ref({ x: 0, y: 0 })

// Computed properties
const isInFunction = computed(() => blueprintStore.isFunctionEditing)

// Check if this is a data-only node (like variables)
const isDataOnlyNode = computed(() => {
  // Check if node was explicitly created as a data-only node
  if (props.node.data?.isDataNode) {
    return true;
  }
  
  // Also consider variable-get and variable-set nodes as data-only
  if (props.node.type === 'variable-get' || props.node.type === 'variable-set') {
    return true;
  }
  
  return false;
})

const nodeTitle = computed(() => {
  return props.nodeType ? props.nodeType.name : props.node.type
})

const nodeStyle = computed(() => {
  return {
    left: `${props.node.position.x}px`,
    top: `${props.node.position.y}px`
  }
})

const execInputPins = computed(() => {
  if (!props.nodeType) return []
  return props.nodeType.inputs.filter(pin => pin.type.id === 'execution')
})

const dataInputPins = computed(() => {
  if (!props.nodeType) return []
  return props.nodeType.inputs.filter(pin => pin.type.id !== 'execution')
})

const execOutputPins = computed(() => {
  if (!props.nodeType) return []
  return props.nodeType.outputs.filter(pin => pin.type.id === 'execution')
})

const dataOutputPins = computed(() => {
  if (!props.nodeType) return []
  return props.nodeType.outputs.filter(pin => pin.type.id !== 'execution')
})

const hasExecInputs = computed(() => execInputPins.value.length > 0)
const hasDataInputs = computed(() => dataInputPins.value.length > 0)
const hasExecOutputs = computed(() => execOutputPins.value.length > 0)
const hasDataOutputs = computed(() => dataOutputPins.value.length > 0)

// Methods

// We need to expose a method to check if a pin is active
function isPinActive(pinId: string): boolean {
  // This would be passed down from BlueprintCanvas
  return props.activePins?.has(`${props.node.id}-${pinId}`) || false;
}

function handleMouseDown(event: MouseEvent) {
  // Prevent default behavior
  event.preventDefault();

  // Stop propagation to prevent canvas from handling the event
  event.stopPropagation();

  // Select this node
  emit('select', props.node.id);
}

function handleHeaderMouseDown(event: MouseEvent) {
  // Only start dragging on left mouse button
  if (event.button !== 0) return;

  // Prevent event propagation
  event.stopPropagation();

  // Start dragging only from header
  isDragging.value = true;
  dragStart.value = { x: event.clientX, y: event.clientY };
  initialPosition.value = { ...props.node.position };

  // Set cursor
  document.body.style.cursor = 'grabbing';

  // Select this node
  emit('select', props.node.id);
}

function handleMouseMove(event: MouseEvent) {
  if (isDragging.value) {
    // Calculate the distance moved
    const dx = event.clientX - dragStart.value.x;
    const dy = event.clientY - dragStart.value.y;

    // Calculate the threshold for starting a drag (prevents accidental drags)
    const dragThreshold = 3; // pixels
    const dragDistance = Math.sqrt(dx * dx + dy * dy);

    // Only move if we've exceeded the threshold
    if (dragDistance > dragThreshold) {
      // Get the canvas scale
      const canvas = nodeElement.value?.closest('.blueprint-canvas') as HTMLElement;
      const transform = window.getComputedStyle(canvas).transform;
      const matrix = new DOMMatrix(transform);
      const scale = matrix.a; // The scale factor is in the 'a' component of the matrix

      // Calculate the new position, accounting for canvas scale
      const newX = initialPosition.value.x + dx / scale;
      const newY = initialPosition.value.y + dy / scale;

      // Emit move event
      emit('move', props.node.id, { x: newX, y: newY });

      // Prevent default to avoid text selection during drag
      event.preventDefault();
    }
  }
}

function handleMouseUp(event: MouseEvent) {
  if (isDragging.value) {
    // Stop dragging
    isDragging.value = false;
    document.body.style.cursor = 'default';

    // Prevent default
    event.preventDefault();
  }
}

function handlePinMouseDown(event: MouseEvent, pin: PinDefinition, isInput: boolean, isExecution: boolean) {
  // Stop propagation to prevent node dragging
  event.stopPropagation()

  // Get pin position
  const pinElement = event.currentTarget as HTMLElement
  const rect = pinElement.getBoundingClientRect()

  // Get canvas element
  const canvas = nodeElement.value?.closest('.blueprint-canvas') as HTMLElement
  const canvasRect = canvas.getBoundingClientRect()

  // Calculate position relative to canvas, accounting for scale
  const transform = window.getComputedStyle(canvas).transform
  const matrix = new DOMMatrix(transform)
  const scale = matrix.a

  const x = (rect.left + rect.width / 2 - canvasRect.left) / scale
  const y = (rect.top + rect.height / 2 - canvasRect.top) / scale

  // For input pins, we'll check if they have default values to use
  let defaultValue = undefined;
  if (isInput && !isExecution && props.node.properties) {
    // Look for a property with the pattern "input_[pinId]"
    const defaultProp = props.node.properties.find(p => p.name === `input_${pin.id}`);
    if (defaultProp) {
      defaultValue = defaultProp.value;
    } else if (pin.default !== undefined) {
      // Fall back to pin's built-in default if available
      defaultValue = pin.default;
    }
  }

  // Emit pin mouse down event
  emit('pin-mouse-down', {
    nodeId: props.node.id,
    pinId: pin.id,
    isInput,
    isExecution,
    position: { x, y },
    defaultValue
  })
}

function handlePinMouseUp(event: MouseEvent, pin: PinDefinition, isInput: boolean, isExecution: boolean) {
  // Stop propagation to prevent node selection
  event.stopPropagation()

  // Emit pin mouse up event
  emit('pin-mouse-up', {
    nodeId: props.node.id,
    pinId: pin.id,
    isInput,
    isExecution
  })
}

function getPinColor(pin: PinDefinition): string {
  // Return a color based on the pin type
  switch (pin.type.id) {
    case 'string':
      return '#f0883e' // Orange
    case 'number':
      return '#6ed69a' // Green
    case 'boolean':
      return '#dc5050' // Red
    case 'object':
      return '#8ab4f8' // Blue
    case 'array':
      return '#bb86fc' // Purple
    case 'any':
      return '#aaaaaa' // Gray
    default:
      return '#aaaaaa' // Default gray
  }
}

// Event listeners
onMounted(() => {
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
})
</script>

<style scoped>
.blueprint-node {
    position: absolute;
    min-width: 180px;
    background-color: var(--node-bg);
    border-radius: 5px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
    user-select: none;
    z-index: 10;
    color: var(--text-color);
    font-family: 'Segoe UI', sans-serif;
    font-size: 13px;
}

/* Special styling for data-only nodes like variables */
.blueprint-node.data-node {
    background-color: var(--node-data-bg, #2d394a); /* Use a different color for data nodes */
    border: 1px solid var(--node-data-border, #44617e);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.4);
}

.blueprint-node.data-node .node-header {
    background-color: var(--node-data-header, #3a4e63);
}

.blueprint-node.data-node .pin-circle {
    box-shadow: 0 0 4px rgba(255, 255, 255, 0.5); /* Add glow effect to pins */
}

.node-header {
    background-color: var(--node-header);
    padding: 8px 10px;
    border-top-left-radius: 5px;
    border-top-right-radius: 5px;
    font-weight: bold;
    cursor: move;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.node-title {
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.node-status-badge {
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 10px;
    text-transform: uppercase;
    font-weight: bold;
    background-color: #444;
}

.node-content {
    padding: 10px;
}

.pin-section {
    margin-bottom: 8px;
}

.pin-section:last-child {
    margin-bottom: 0;
}

.pin-row {
    display: flex;
    align-items: center;
    height: 24px;
    margin: 2px 0;
}

.pin-row-output {
    justify-content: flex-end;
}

.pin-label {
    margin: 0 5px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.node-pin {
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
}

.pin-circle {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background-color: var(--input-pin);
}

.pin-output .pin-circle {
    background-color: var(--output-pin);
}

.pin-triangle {
    width: 0;
    height: 0;
    border-style: solid;
}

.pin-input.pin-exec .pin-triangle {
    border-width: 5px 8px 5px 0;
    border-color: transparent var(--exec-pin) transparent transparent;
}

.pin-output.pin-exec .pin-triangle {
    border-width: 5px 0 5px 8px;
    border-color: transparent transparent transparent var(--exec-pin);
}

.pin-divider {
    height: 1px;
    background-color: #444;
    margin: 8px 0;
}

/* Node status styles */
.blueprint-node.selected {
    box-shadow: 0 0 0 2px var(--node-selected), 0 2px 5px rgba(0, 0, 0, 0.3);
}

.blueprint-node.executing {
  box-shadow: 0 0 0 2px #ffcc00, 0 0 10px rgba(255, 204, 0, 0.5);
  animation: node-executing 1.5s ease-in-out infinite;
}

@keyframes node-executing {
  0% { box-shadow: 0 0 0 2px #ffcc00, 0 0 10px rgba(255, 204, 0, 0.3); }
  50% { box-shadow: 0 0 0 2px #ffcc00, 0 0 15px rgba(255, 204, 0, 0.8); }
  100% { box-shadow: 0 0 0 2px #ffcc00, 0 0 10px rgba(255, 204, 0, 0.3); }
}

.blueprint-node.executing .node-status-badge {
  background-color: #ffcc00;
  color: #333;
}

.blueprint-node.completed {
  box-shadow: 0 0 0 2px #00cc00, 0 0 10px rgba(0, 204, 0, 0.5);
  transition: box-shadow 0.3s ease-out;
}

.blueprint-node.completed .node-status-badge {
  background-color: #00cc00;
  color: #fff;
}

.blueprint-node.error {
  box-shadow: 0 0 0 2px #cc0000, 0 0 10px rgba(204, 0, 0, 0.5);
  animation: node-error 0.5s ease-in-out 3;
}

@keyframes node-error {
  0% { box-shadow: 0 0 0 2px #cc0000, 0 0 10px rgba(204, 0, 0, 0.5); }
  50% { box-shadow: 0 0 0 3px #cc0000, 0 0 20px rgba(204, 0, 0, 0.8); }
  100% { box-shadow: 0 0 0 2px #cc0000, 0 0 10px rgba(204, 0, 0, 0.5); }
}

.blueprint-node.error .node-status-badge {
  background-color: #cc0000;
  color: #fff;
}

.blueprint-node.executing .node-status-badge {
    background-color: #ffcc00;
    color: #333;
}

.blueprint-node.completed {
    box-shadow: 0 0 0 2px #00cc00, 0 0 10px rgba(0, 204, 0, 0.5);
}

.blueprint-node.completed .node-status-badge {
    background-color: #00cc00;
    color: #fff;
}

.blueprint-node.error {
    box-shadow: 0 0 0 2px #cc0000, 0 0 10px rgba(204, 0, 0, 0.5);
}

.blueprint-node.error .node-status-badge {
    background-color: #cc0000;
    color: #fff;
}
</style>

case '
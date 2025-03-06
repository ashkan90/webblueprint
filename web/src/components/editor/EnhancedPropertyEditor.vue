<!-- File: web/src/components/editor/EnhancedPropertyEditor.vue -->
<template>
  <div class="enhanced-property-editor">
    <div class="editor-header">
      <h3 class="node-title">{{ nodeTitle }}</h3>
      <div class="node-type">{{ nodeType?.typeId || 'Unknown Type' }}</div>

      <div class="tab-controls">
        <button
            v-for="tab in tabs"
            :key="tab.id"
            class="tab-button"
            :class="{ active: activeTab === tab.id }"
            @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <div class="editor-content">
      <!-- Details Tab -->
      <div v-if="activeTab === 'details'">
        <div class="section">
          <h4 class="section-title">Node ID</h4>
          <div class="property-row">
            <div class="property-label">ID</div>
            <div class="property-value id-display">{{ node.id }}</div>
          </div>
        </div>

        <div class="section">
          <h4 class="section-title">Position</h4>
          <div class="position-editor">
            <div class="property-row">
              <div class="property-label">X</div>
              <div class="property-input-wrapper">
                <input
                    type="number"
                    v-model.number="position.x"
                    class="property-input"
                    @change="updatePosition"
                />
              </div>
            </div>
            <div class="property-row">
              <div class="property-label">Y</div>
              <div class="property-input-wrapper">
                <input
                    type="number"
                    v-model.number="position.y"
                    class="property-input"
                    @change="updatePosition"
                />
              </div>
            </div>
          </div>
        </div>

        <div v-if="nodeType?.description" class="section">
          <h4 class="section-title">Description</h4>
          <div class="node-description">
            {{ nodeType.description }}
          </div>
        </div>

        <div v-if="properties.length > 0" class="section">
          <h4 class="section-title">Properties</h4>
          <div
              v-for="property in properties"
              :key="property.name"
              class="property-editor"
          >
            <div class="property-row">
              <div class="property-label">{{ formatPropertyName(property.name) }}</div>

              <!-- String property -->
              <div class="property-input-wrapper" v-if="property.type === 'string'">
                <input
                    type="text"
                    v-model="property.value"
                    class="property-input"
                    @change="updateProperty(property.name, property.value)"
                />
              </div>

              <!-- Number property -->
              <div class="property-input-wrapper" v-else-if="property.type === 'number'">
                <input
                    type="number"
                    v-model.number="property.value"
                    class="property-input"
                    @change="updateProperty(property.name, property.value)"
                />
              </div>

              <!-- Boolean property -->
              <div class="property-input-wrapper" v-else-if="property.type === 'boolean'">
                <label class="toggle-switch">
                  <input
                      type="checkbox"
                      v-model="property.value"
                      @change="updateProperty(property.name, property.value)"
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>

              <!-- Select property -->
              <div class="property-input-wrapper" v-else-if="property.type === 'select'">
                <select
                    v-model="property.value"
                    class="property-select"
                    @change="updateProperty(property.name, property.value)"
                >
                  <option
                      v-for="option in property.options"
                      :key="option"
                      :value="option"
                  >
                    {{ option }}
                  </option>
                </select>
              </div>

              <!-- Default for other property types -->
              <div class="property-input-wrapper" v-else>
                <input
                    type="text"
                    v-model="property.value"
                    class="property-input"
                    @change="updateProperty(property.name, property.value)"
                />
              </div>
            </div>

            <div class="property-description" v-if="getPropertyDescription(property.name)">
              {{ getPropertyDescription(property.name) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Inputs Tab -->
      <div v-else-if="activeTab === 'inputs'">
        <div class="section">
          <h4 class="section-title">Input Pins</h4>
          <div
              v-for="pin in nodeType?.inputs"
              :key="pin.id"
              class="pin-info-wrapper"
              :class="{ connected: isPinConnected(pin.id, 'input') }"
          >
            <div class="pin-header">
              <div class="pin-color-indicator" :style="{ backgroundColor: getPinColor(pin.type.id) }"></div>
              <div class="pin-name">{{ pin.name }}</div>
              <div class="pin-type">{{ pin.type.name }}</div>
              <div class="pin-badges">
                <span v-if="pin.optional" class="pin-badge optional">Optional</span>
                <span v-else class="pin-badge required">Required</span>
              </div>
            </div>

            <div class="pin-description" v-if="pin.description">
              {{ pin.description }}
            </div>

            <!-- Default Value Editor (only for non-execution pins that aren't connected) -->
            <div
                v-if="!isPinConnected(pin.id, 'input') && pin.type.id !== 'execution'"
                class="pin-default-editor"
            >
              <div class="default-label">Default Value:</div>

              <!-- String input -->
              <div v-if="pin.type.id === 'string'" class="default-input-wrapper">
                <input
                    type="text"
                    v-model="pinDefaults[pin.id]"
                    class="property-input"
                    @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
                    :placeholder="pin.default !== undefined ? String(pin.default) : ''"
                />
              </div>

              <!-- Number input -->
              <div v-else-if="pin.type.id === 'number'" class="default-input-wrapper">
                <input
                    type="number"
                    v-model.number="pinDefaults[pin.id]"
                    class="property-input"
                    @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
                    :placeholder="pin.default !== undefined ? String(pin.default) : ''"
                />
              </div>

              <!-- Boolean input -->
              <div v-else-if="pin.type.id === 'boolean'" class="default-input-wrapper">
                <label class="toggle-switch">
                  <input
                      type="checkbox"
                      v-model="pinDefaults[pin.id]"
                      @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>

              <!-- Object/Array input -->
              <div v-else-if="pin.type.id === 'object' || pin.type.id === 'array'" class="default-input-wrapper">
                <textarea
                    v-model="pinDefaults[pin.id]"
                    class="property-textarea"
                    @change="updateJsonPinDefault(pin.id, pinDefaults[pin.id])"
                    :placeholder="pin.type.id === 'object' ? '{}' : '[]'"
                ></textarea>
              </div>

              <!-- Any type input -->
              <div v-else class="default-input-wrapper">
                <input
                    type="text"
                    v-model="pinDefaults[pin.id]"
                    class="property-input"
                    @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
                    :placeholder="pin.default !== undefined ? String(pin.default) : ''"
                />
              </div>
            </div>

            <div v-else-if="isPinConnected(pin.id, 'input')" class="pin-connection-info">
              <div class="connection-tag">
                <span class="connection-icon">🔗</span>
                Connected
              </div>
            </div>
          </div>

          <div v-if="!nodeType?.inputs?.length" class="no-pins-message">
            No input pins available.
          </div>
        </div>
      </div>

      <!-- Outputs Tab -->
      <div v-else-if="activeTab === 'outputs'">
        <div class="section">
          <h4 class="section-title">Output Pins</h4>
          <div
              v-for="pin in nodeType?.outputs"
              :key="pin.id"
              class="pin-info-wrapper"
              :class="{ connected: isPinConnected(pin.id, 'output') }"
          >
            <div class="pin-header">
              <div class="pin-color-indicator" :style="{ backgroundColor: getPinColor(pin.type.id) }"></div>
              <div class="pin-name">{{ pin.name }}</div>
              <div class="pin-type">{{ pin.type.name }}</div>
            </div>

            <div class="pin-description" v-if="pin.description">
              {{ pin.description }}
            </div>

            <div v-if="isPinConnected(pin.id, 'output')" class="pin-connection-info">
              <div class="connection-tag">
                <span class="connection-icon">🔗</span>
                Connected
              </div>
              <div class="connection-count">
                {{ getConnectionCount(pin.id, 'output') }} connection(s)
              </div>
            </div>
          </div>

          <div v-if="!nodeType?.outputs?.length" class="no-pins-message">
            No output pins available.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useBlueprintStore } from '../../stores/blueprint'
import type { Node } from '../../types/blueprint'
import type { NodeTypeDefinition, PinDefinition } from '../../types/nodes'

const props = defineProps<{
  node: Node
  nodeType: NodeTypeDefinition | null
  selected: boolean
}>()

const emit = defineEmits<{
  (e: 'property-changed', nodeId: string, propertyName: string, value: any): void
  (e: 'pin-default-changed', nodeId: string, pinId: string, value: any): void
}>()

// Blueprint store for connection info
const blueprintStore = useBlueprintStore()

// State
const position = ref({ x: props.node.position.x, y: props.node.position.y })
const properties = ref<Array<{ name: string, value: any, type: string, options?: string[] }>>([])
const pinDefaults = ref<Record<string, any>>({})
const activeTab = ref('details')

// Available tabs
const tabs = [
  { id: 'details', label: 'Details' },
  { id: 'inputs', label: 'Inputs' },
  { id: 'outputs', label: 'Outputs' },
]

// Computed properties
const nodeTitle = computed(() => {
  return props.nodeType?.name || 'Node Properties'
})

// Methods
function formatPropertyName(name: string): string {
  // Convert from camelCase or snake_case to Title Case
  return name
      // Add space before uppercase letters
      .replace(/([A-Z])/g, ' $1')
      // Replace underscores with spaces
      .replace(/_/g, ' ')
      // Capitalize first letter and trim
      .replace(/^./, str => str.toUpperCase())
      .trim()
}

function updatePosition() {
  emit('property-changed', props.node.id, 'position', {
    x: position.value.x,
    y: position.value.y
  })
}

function updateProperty(name: string, value: any) {
  emit('property-changed', props.node.id, name, value)
}

function updatePinDefault(pinId: string, value: any) {
  emit('pin-default-changed', props.node.id, pinId, value)
}

function updateJsonPinDefault(pinId: string, jsonString: string) {
  try {
    // Try to parse the JSON string
    const value = JSON.parse(jsonString)
    emit('pin-default-changed', props.node.id, pinId, value)
  } catch (e) {
    // If it's not valid JSON, use the string as is
    console.warn(`Invalid JSON for pin ${pinId}:`, e)
    // Optionally show a warning to the user
  }
}

function getPropertyDescription(name: string): string {
  // This could be enhanced to fetch descriptions from a schema or documentation
  return ''
}

function isPinConnected(pinId: string, direction: 'input' | 'output'): boolean {
  return blueprintStore.isNodePinConnected(props.node.id, pinId, direction)
}

function getConnectionCount(pinId: string, direction: 'input' | 'output'): number {
  const connections = blueprintStore.connections.filter(conn => {
    if (direction === 'input') {
      return conn.targetNodeId === props.node.id && conn.targetPinId === pinId
    } else {
      return conn.sourceNodeId === props.node.id && conn.sourcePinId === pinId
    }
  })
  return connections.length
}

function getPinColor(typeId: string): string {
  // Same colors as used in BlueprintNode.vue for consistency
  switch (typeId) {
    case 'execution':
      return '#ffffff' // White
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

// Initialize properties from node data
function initializeProperties() {
  properties.value = []
  pinDefaults.value = {}

  // First load properties from the node
  const nodeProps = props.node.properties || []

  // Then check if the node type has property definitions
  if (props.nodeType) {
    // For now, we'll create simple property editors based on the property values
    nodeProps.forEach(prop => {
      let type = 'string'
      let options: string[] | undefined = undefined

      // Determine property type
      if (typeof prop.value === 'number') {
        type = 'number'
      } else if (typeof prop.value === 'boolean') {
        type = 'boolean'
      } else if (Array.isArray(prop.value)) {
        type = 'select'
        options = prop.value
      }

      // Add to properties if it's not a pin default
      if (!prop.name.startsWith('input_')) {
        properties.value.push({
          name: prop.name,
          value: prop.value,
          type,
          options
        })
      }
    })

    // Initialize pin defaults from node properties
    props.nodeType.inputs.forEach(pin => {
      // Skip execution pins
      if (pin.type.id === 'execution') return

      // Find default value in node properties if exists
      const defaultPropName = `input_${pin.id}`
      const pinProp = nodeProps.find(prop => prop.name === defaultPropName)

      if (pinProp) {
        // Use stored default value
        pinDefaults.value[pin.id] = pinProp.value
      } else if (pin.default !== undefined) {
        // Use pin's default value from the type definition
        pinDefaults.value[pin.id] = pin.default
      } else {
        // Initialize with type-appropriate empty value
        switch (pin.type.id) {
          case 'string':
            pinDefaults.value[pin.id] = ''
            break
          case 'number':
            pinDefaults.value[pin.id] = 0
            break
          case 'boolean':
            pinDefaults.value[pin.id] = false
            break
          case 'object':
            pinDefaults.value[pin.id] = '{}'
            break
          case 'array':
            pinDefaults.value[pin.id] = '[]'
            break
          default:
            pinDefaults.value[pin.id] = ''
        }
      }

      // Format JSON for display
      if (pin.type.id === 'object' || pin.type.id === 'array') {
        if (typeof pinDefaults.value[pin.id] !== 'string') {
          try {
            pinDefaults.value[pin.id] = JSON.stringify(pinDefaults.value[pin.id], null, 2)
          } catch (e) {
            pinDefaults.value[pin.id] = pin.type.id === 'object' ? '{}' : '[]'
          }
        }
      }
    })
  }
}

// Watch for node changes
watch(() => props.node, () => {
  position.value = { x: props.node.position.x, y: props.node.position.y }
  initializeProperties()
}, { deep: true })

// Watch for nodeType changes
watch(() => props.nodeType, () => {
  initializeProperties()
}, { deep: true })

// Initialize
initializeProperties()
</script>

<style scoped>
.enhanced-property-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  color: var(--text-color);
  background-color: var(--node-bg);
}

.editor-header {
  padding: 15px;
  border-bottom: 1px solid #3d3d3d;
  background-color: var(--node-header);
}

.node-title {
  margin: 0 0 5px 0;
  font-size: 1.1rem;
  color: var(--text-color);
}

.node-type {
  font-size: 0.8rem;
  color: #aaa;
  margin-bottom: 15px;
}

.tab-controls {
  display: flex;
  border-bottom: 1px solid #3d3d3d;
  margin: 0 -15px -15px -15px;
}

.tab-button {
  padding: 8px 12px;
  background: none;
  border: none;
  color: #aaa;
  font-size: 0.9rem;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-button:hover {
  color: white;
  background-color: rgba(255, 255, 255, 0.05);
}

.tab-button.active {
  color: var(--accent-blue);
  border-bottom-color: var(--accent-blue);
}

.editor-content {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
}

.section {
  margin-bottom: 20px;
}

.section-title {
  margin: 0 0 10px 0;
  font-size: 0.95rem;
  color: var(--text-color);
  border-bottom: 1px solid #3d3d3d;
  padding-bottom: 5px;
}

.property-row {
  display: flex;
  margin-bottom: 10px;
}

.property-label {
  font-size: 0.9rem;
  color: #bbb;
  padding-top: 6px;
  padding-right: 6px;
}

.property-input-wrapper,
.default-input-wrapper {
  flex: 1;
}

.property-input,
.property-select,
.property-textarea {
  width: 100%;
  background-color: #3d3d3d;
  border: 1px solid #4d4d4d;
  color: white;
  border-radius: 3px;
  padding: 5px 8px;
  font-size: 0.9rem;
}

.property-textarea {
  min-height: 80px;
  font-family: monospace;
  resize: vertical;
}

.property-input:focus,
.property-select:focus,
.property-textarea:focus {
  outline: none;
  border-color: var(--accent-blue);
}

.property-description,
.pin-description {
  margin-left: 100px;
  margin-top: -5px;
  margin-bottom: 10px;
  font-size: 0.8rem;
  color: #888;
}

.id-display {
  font-family: monospace;
  background-color: #3d3d3d;
  padding: 5px 8px;
  border-radius: 3px;
  font-size: 0.9rem;
  word-break: break-all;
  color: #aaa;
}

.node-description {
  color: #bbb;
  line-height: 1.5;
  font-size: 0.9rem;
}

.position-editor {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.position-editor .property-row {
  margin-bottom: 0;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
  margin-top: 5px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #555;
  transition: .3s;
  border-radius: 20px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: var(--accent-blue);
}

input:checked + .toggle-slider:before {
  transform: translateX(20px);
}

.pin-info-wrapper {
  background-color: #333;
  border-radius: 4px;
  padding: 10px;
  margin-bottom: 10px;
}

.pin-info-wrapper.connected {
  border-left: 3px solid var(--accent-blue);
}

.pin-header {
  display: flex;
  align-items: center;
  margin-bottom: 5px;
}

.pin-color-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-right: 8px;
}

.pin-name {
  font-weight: 500;
  margin-right: 8px;
}

.pin-type {
  font-size: 0.8rem;
  color: #aaa;
  background-color: #444;
  padding: 2px 6px;
  border-radius: 10px;
  margin-right: 8px;
}

.pin-badges {
  display: flex;
}

.pin-badge {
  font-size: 0.7rem;
  padding: 2px 6px;
  border-radius: 10px;
  margin-right: 5px;
}

.pin-badge.optional {
  background-color: #666;
  color: white;
}

.pin-badge.required {
  background-color: var(--accent-red);
  color: white;
}

.pin-description {
  margin-left: 18px;
  margin-bottom: 10px;
  color: #aaa;
  font-size: 0.85rem;
}

.pin-default-editor {
  margin-top: 5px;
  padding: 10px;
  background-color: #2d2d2d;
  border-radius: 3px;
}

.default-label {
  font-size: 0.8rem;
  color: #888;
  margin-bottom: 5px;
}

.pin-connection-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 5px;
}

.connection-tag {
  display: inline-flex;
  align-items: center;
  background-color: rgba(52, 152, 219, 0.2);
  color: var(--accent-blue);
  font-size: 0.8rem;
  padding: 3px 8px;
  border-radius: 3px;
}

.connection-icon {
  margin-right: 5px;
}

.connection-count {
  font-size: 0.8rem;
  color: #aaa;
}

.no-pins-message {
  color: #888;
  font-style: italic;
  text-align: center;
  padding: 20px 0;
}

.property-editor {
  margin-bottom: 15px;
}
</style>
<template>
  <div class="node-properties-panel">
    <div class="panel-header">
      <h3>{{ nodeTitle }}</h3>
      <div class="node-type">{{ nodeType?.typeId || 'Unknown Type' }}</div>
    </div>

    <div v-if="nodeType" class="panel-content">
      <div class="section">
        <h4>Node ID</h4>
        <div class="property-row">
          <div class="property-label">ID</div>
          <div class="property-value">{{ node.id }}</div>
        </div>
      </div>

      <div class="section">
        <h4>Position</h4>
        <div class="property-row">
          <div class="property-label">X</div>
          <input
              type="number"
              v-model.number="position.x"
              class="property-input"
              @change="updatePosition"
          />
        </div>
        <div class="property-row">
          <div class="property-label">Y</div>
          <input
              type="number"
              v-model.number="position.y"
              class="property-input"
              @change="updatePosition"
          />
        </div>
      </div>

      <div v-if="properties.length > 0" class="section">
        <h4>Properties</h4>
        <div
            v-for="property in properties"
            :key="property.name"
            class="property-row"
        >
          <div class="property-label">{{ property.name }}</div>

          <!-- String property -->
          <input
              v-if="property.type === 'string'"
              type="text"
              v-model="property.value"
              class="property-input"
              @change="updateProperty(property.name, property.value)"
          />

          <!-- Number property -->
          <input
              v-else-if="property.type === 'number'"
              type="number"
              v-model.number="property.value"
              class="property-input"
              @change="updateProperty(property.name, property.value)"
          />

          <!-- Boolean property -->
          <div v-else-if="property.type === 'boolean'" class="property-boolean">
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
          <select
              v-else-if="property.type === 'select'"
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

          <!-- Default for other property types -->
          <input
              v-else
              type="text"
              v-model="property.value"
              class="property-input"
              @change="updateProperty(property.name, property.value)"
          />
        </div>
      </div>

      <!-- Node description -->
      <div class="section">
        <h4>Description</h4>
        <div class="node-description">
          {{ nodeType.description }}
        </div>
      </div>

      <!-- Input pins and properties -->
      <div v-if="nodeType.inputs.length > 0" class="section">
        <h4>Inputs & Properties</h4>
        <div
            v-for="pin in nodeType.inputs"
            :key="pin.id"
            class="pin-info"
        >
          <div class="pin-info-header">
            <div class="pin-name">{{ pin.name }}</div>
            <div class="pin-type">{{ pin.type.name }}</div>
          </div>
          <div v-if="pin.description" class="pin-description">
            {{ pin.description }}
          </div>

          <!-- Special handling for Constant node properties -->
          <div v-if="!isPinConnected(pin.id) && pin.type.id !== 'execution'" class="pin-property-editor">
            <!-- String input -->
            <div v-if="pin.type.id === 'string'">
              <input
                  v-if="isConstantValuePin(pin)"
                  type="text"
                  v-model="constantValue"
                  class="property-input"
                  @change="updateValue(pin)"
              />
              <!--  -->
              <input
                  v-else-if="!isPinConnected(pin.id)"
                  type="text"
                  v-model="pinDefaults[pin.id]"
                  class="property-input"
                  @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
              />
            </div>

            <!-- Number input -->
            <div v-if="pin.type.id === 'number'">
              <input
                  v-if="isConstantValuePin(pin)"
                  type="number"
                  v-model.number="constantValue"
                  class="property-input"
                  @change="updateValue(pin)"
              />
              <!--  -->
              <input
                  v-else-if="!isPinConnected(pin.id)"
                  type="number"
                  v-model.number="pinDefaults[pin.id]"
                  class="property-input"
                  @change="updateValue(pin)"
              />
            </div>

            <!-- Boolean input -->
            <div v-if="pin.type.id === 'boolean'" class="property-boolean">
              <label class="toggle-switch">
                <input
                    type="checkbox"
                    v-if="isConstantValuePin(pin)"
                    v-model="constantValue"
                    @change="updateValue(pin)"
                />
                <input
                    type="checkbox"
                    v-else-if="!isPinConnected(pin.id)"
                    v-model="pinDefaults[pin.id]"
                    @change="updateValue(pin)"
                />
                <span class="toggle-slider"></span>
              </label>
            </div>

            <div v-if="pin.type.id === 'any'">
              <input
                  v-if="isConstantValuePin(pin)"
                  type="text"
                  class="property-input"
                  v-model="constantValue"
                  @change="updateValue(pin)"
              />
              <input
                  v-else-if="!isPinConnected(pin.id)"
                  type="text"
                  class="property-input"
                  v-model="pinDefaults[pin.id]"
                  @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Output pins -->
      <div v-if="nodeType.outputs.length > 0" class="section">
        <h4>Outputs</h4>
        <div
            v-for="pin in nodeType.outputs"
            :key="pin.id"
            class="pin-info"
        >
          <div class="pin-info-header">
            <div class="pin-name">{{ pin.name }}</div>
            <div class="pin-type">{{ pin.type.name }}</div>
          </div>
          <div v-if="pin.description" class="pin-description">
            {{ pin.description }}
          </div>
          <!-- Special handling for Constant node properties -->
          <div v-if="!isPinConnected(pin.id) && pin.type.id !== 'execution' && !isConstantValuePin(pin)" class="pin-property-editor">
            <!-- String input -->
            <input
                v-if="pin.type.id === 'string'"
                type="text"
                v-model="pinDefaults[pin.id]"
                class="property-input"
                @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
            />

            <!-- Number input -->
            <input
                v-else-if="pin.type.id === 'number'"
                type="number"
                v-model.number="pinDefaults[pin.id]"
                class="property-input"
                @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
            />

            <!-- Boolean input -->
            <div v-else-if="pin.type.id === 'boolean'" class="property-boolean">
              <label class="toggle-switch">
                <input
                    type="checkbox"
                    v-model="pinDefaults[pin.id]"
                    @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
                />
                <span class="toggle-slider"></span>
              </label>
            </div>

            <input
              v-else
              type="text"
              class="property-input"
              v-model="pinDefaults[pin.id]"
              @change="updatePinDefault(pin.id, pinDefaults[pin.id])"
            />
          </div>
        </div>
      </div>
    </div>

    <div v-else class="panel-content empty">
      <div class="empty-message">
        No node type information available.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useBlueprintStore } from '../../stores/blueprint'
import type { Node } from '../../types/blueprint'
import type {NodeTypeDefinition, PinDefinition} from '../../types/nodes'

const props = defineProps<{
  node: Node
  nodeType: NodeTypeDefinition | null
  selected: boolean
}>()

const emit = defineEmits<{
  (e: 'property-changed', nodeId: string, propertyName: string, value: any): void
  (e: 'pin-default-changed', nodeId: string, pinId: string, value: any): void
}>()

// Reactive state
const position = ref({ x: props.node.position.x, y: props.node.position.y })
const properties = ref<Array<{ name: string, value: any, type: string, options?: string[] }>>([])
const pinDefaults = ref<Record<string, any>>({})
const constantValue = ref<any>('')

// Computed
const nodeTitle = computed(() => {
  return props.nodeType?.name || 'Node Properties'
})

// Methods
// Initialize constant value from node properties
function initConstantValue() {
  // Check if this is a constant node
  if (props.nodeType && props.nodeType.typeId.startsWith('constant-')) {
    const property = props.node.properties.find(p => p.name === 'constantValue')

    if (property) {
      constantValue.value = property.value
    } else {
      // Set default value based on type
      if (props.nodeType.typeId === 'constant-string') {
        constantValue.value = ''
      } else if (props.nodeType.typeId === 'constant-number') {
        constantValue.value = 0
      } else if (props.nodeType.typeId === 'constant-boolean') {
        constantValue.value = false
      }
    }
  }
}

// Check if this pin is a constant value input
function isConstantValuePin(pin: any): boolean {
  return pin.id === 'constantValue' || props.nodeType && props.nodeType.typeId.startsWith('constant-')
}

// Update the constant value property
function updateValue(pin: PinDefinition) {
  if (isConstantValuePin(pin)) {
    emit('property-changed', props.node.id, 'constantValue', constantValue.value)
    return
  }
}

function updatePosition() {
  emit('property-changed', props.node.id, 'position', {
    x: position.value.x,
    y: position.value.y
  })
}

function updatePinDefault(pinId: string, value: any) {
  emit('pin-default-changed', props.node.id, pinId, value)
}

function updateProperty(name: string, value: any) {
  emit('property-changed', props.node.id, name, value)
}

function updateJsonPinDefault(pinId: string, jsonString: string) {
  try {
    // Try to parse the JSON string
    const value = JSON.parse(jsonString)
    emit('pin-default-changed', props.node.id, pinId, value)
  } catch (e) {
    // If it's not valid JSON, use the string as is
    console.warn(`Invalid JSON for pin ${pinId}:`, e)
    // Optionally, you could show an error to the user here
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
            pinDefaults.value[pin.id] = '{}'
          }
        }
      }
    })
  }
}

function isPinConnected(pinId: string): boolean {
  // Use blueprint store to check if pin has connections
  const blueprintStore = useBlueprintStore();

  // Check if this input pin has any incoming connections
  return blueprintStore.isNodePinConnected(props.node.id, pinId, 'input');
}

// Watch for node changes
watch(() => props.node, () => {
  position.value = { x: props.node.position.x, y: props.node.position.y }
  initializeProperties()
  initConstantValue()
}, { deep: true })

// Watch for nodeType changes
watch(() => props.nodeType, () => {
  initConstantValue()
}, { deep: true })

// Initialize
initializeProperties()
initConstantValue()
</script>

<style scoped>
.node-properties-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-y: auto;
}

.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid #444;
}

.panel-header h3 {
  margin: 0 0 5px 0;
  font-size: 1.1rem;
}

.node-type {
  font-size: 0.8rem;
  color: #aaa;
}

.panel-content {
  padding: 16px;
  flex: 1;
}

.panel-content.empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-message {
  color: #aaa;
  text-align: center;
}

.section {
  margin-bottom: 20px;
}

.section h4 {
  margin: 0 0 8px 0;
  font-size: 0.9rem;
  color: #ddd;
  border-bottom: 1px solid #444;
  padding-bottom: 5px;
}

.property-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.property-label {
  width: 100px;
  font-size: 0.9rem;
  color: #aaa;
}

.property-input,
.property-select {
  flex: 1;
  background-color: #444;
  border: 1px solid #555;
  border-radius: 4px;
  padding: 4px 8px;
  color: white;
}

.property-input:focus,
.property-select:focus {
  outline: none;
  border-color: var(--accent-blue);
}

.property-boolean {
  flex: 1;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
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

.property-value {
  flex: 1;
  font-family: monospace;
  background-color: #333;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9rem;
  word-break: break-all;
}

.node-description {
  color: #bbb;
  font-size: 0.9rem;
  line-height: 1.4;
}

.pin-info {
  margin-bottom: 10px;
  background-color: #333;
  border-radius: 4px;
  padding: 8px;
}

.pin-info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.pin-name {
  font-weight: 500;
}

.pin-type {
  font-size: 0.8rem;
  color: #aaa;
  background-color: #444;
  padding: 2px 6px;
  border-radius: 10px;
}

.pin-description {
  font-size: 0.8rem;
  color: #bbb;
}

.pin-property-editor {
  margin-top: 8px;
  padding: 8px;
  background-color: #383838;
  border-radius: 4px;
}

.pin-default-editor {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #444;
}

.default-label {
  font-size: 0.8rem;
  color: #aaa;
  margin-bottom: 4px;
}

.property-textarea {
  min-height: 80px;
  font-family: monospace;
  width: 100%;
  resize: vertical;
}
</style>
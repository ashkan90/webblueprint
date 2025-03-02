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

      <!-- Input pins -->
      <div v-if="nodeType.inputs.length > 0" class="section">
        <h4>Inputs</h4>
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
import type { Node } from '../../types/blueprint'
import type { NodeTypeDefinition } from '../../types/nodes'

const props = defineProps<{
  node: Node
  nodeType: NodeTypeDefinition | null
}>()

const emit = defineEmits<{
  (e: 'property-changed', nodeId: string, propertyName: string, value: any): void
}>()

// Reactive state
const position = ref({ x: props.node.position.x, y: props.node.position.y })
const properties = ref<Array<{ name: string, value: any, type: string, options?: string[] }>>([])

// Computed
const nodeTitle = computed(() => {
  return props.nodeType?.name || 'Node Properties'
})

// Methods
function updatePosition() {
  emit('property-changed', props.node.id, 'position', {
    x: position.value.x,
    y: position.value.y
  })
}

function updateProperty(name: string, value: any) {
  emit('property-changed', props.node.id, name, value)
}

// Initialize properties from node data
function initializeProperties() {
  properties.value = []

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

      properties.value.push({
        name: prop.name,
        value: prop.value,
        type,
        options
      })
    })
  }
}

// Watch for node changes
watch(() => props.node, () => {
  position.value = { x: props.node.position.x, y: props.node.position.y }
  initializeProperties()
}, { deep: true })

// Initialize
initializeProperties()
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
</style>
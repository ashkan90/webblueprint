<template>
  <div class="blueprint-left-panel">
    <div class="panel-search">
      <div class="search-container">
        <input
            v-model="searchQuery"
            type="text"
            placeholder="Search..."
            class="search-input"
        />
        <button @click="searchQuery = ''" class="clear-btn" v-if="searchQuery">
          ✕
        </button>
      </div>
      <button class="add-btn" @click="handleAddButtonClick">
        <span class="add-icon">+</span> Add
      </button>
    </div>

    <div class="panel-sections">
      <!-- GRAPHS Section -->
      <div class="section">
        <div class="section-header" @click="toggleSection('graphs')">
          <div class="section-expand" :class="{ expanded: expandedSections.graphs }">
            {{ expandedSections.graphs ? '▼' : '▶' }}
          </div>
          <div class="section-title">GRAPHS</div>
          <div class="section-actions">
            <button class="section-action-btn" title="Add Graph">+</button>
          </div>
        </div>

        <div v-if="expandedSections.graphs" class="section-content">
          <div
              v-for="graph in filteredGraphs"
              :key="graph.id"
              class="section-item"
              :class="{ active: selectedItem === graph.id }"
              @click="selectItem(graph.id)"
          >
            <div class="item-icon">📊</div>
            <div class="item-name">{{ graph.name }}</div>
          </div>

          <div v-if="filteredGraphs.length === 0" class="empty-section">
            No graphs available
          </div>
        </div>
      </div>

      <!-- FUNCTIONS Section -->
      <div class="section">
        <div class="section-header" @click="toggleSection('functions')">
          <div class="section-expand" :class="{ expanded: expandedSections.functions }">
            {{ expandedSections.functions ? '▼' : '▶' }}
          </div>
          <div class="section-title">FUNCTIONS</div>
          <div class="section-actions">
            <button class="section-action-btn" title="Add Function" @click.stop="showCreateFunctionModal = true">+</button>
          </div>
        </div>

        <div v-if="expandedSections.functions" class="section-content">
          <div
              v-for="func in filteredFunctions"
              :key="func.id"
              class="section-item"
              :class="{ active: selectedItem === func.id }"
              @click="selectItem(func.id)"
              draggable="true"
              @dragstart="onFunctionDragStart($event, func)"
          >
            <div class="item-icon">🔧</div>
            <div class="item-name">{{ func.name }}</div>
          </div>

          <div v-if="filteredFunctions.length === 0" class="empty-section">
            No functions available
          </div>
        </div>
      </div>

      <!-- MACROS Section -->
      <div class="section">
        <div class="section-header" @click="toggleSection('macros')">
          <div class="section-expand" :class="{ expanded: expandedSections.macros }">
            {{ expandedSections.macros ? '▼' : '▶' }}
          </div>
          <div class="section-title">MACROS</div>
          <div class="section-actions">
            <button class="section-action-btn" title="Add Macro" @click.stop="showCreateMacroModal = true">+</button>
          </div>
        </div>

        <div v-if="expandedSections.macros" class="section-content">
          <div
              v-for="macro in filteredMacros"
              :key="macro.id"
              class="section-item"
              :class="{ active: selectedItem === macro.id }"
              @click="selectItem(macro.id)"
              draggable="true"
              @dragstart="onMacroDragStart($event, macro)"
          >
            <div class="item-icon">📦</div>
            <div class="item-name">{{ macro.name }}</div>
          </div>

          <div v-if="filteredMacros.length === 0" class="empty-section">
            No macros available
          </div>
        </div>
      </div>

      <!-- VARIABLES Section -->
      <div class="section">
        <div class="section-header" @click="toggleSection('variables')">
          <div class="section-expand" :class="{ expanded: expandedSections.variables }">
            {{ expandedSections.variables ? '▼' : '▶' }}
          </div>
          <div class="section-title">VARIABLES</div>
          <div class="section-actions">
            <button class="section-action-btn" title="Add Variable" @click.stop="showCreateVariableModal = true">+</button>
          </div>
        </div>

        <div v-if="expandedSections.variables" class="section-content">
          <div
              v-for="variable in filteredVariables"
              :key="variable.id"
              class="section-item variable-item"
              :class="{ active: selectedItem === variable.id }"
              @click="selectItem(variable.id)"
              draggable="true"
              @dragstart="onVariableDragStart($event, variable)"
          >
            <div class="variable-type-indicator" :style="{ backgroundColor: getVariableTypeColor(variable.type) }"></div>
            <div class="item-name">{{ variable.name }}</div>
            <div class="item-type">{{ variable.type }}</div>
          </div>

          <div v-if="filteredVariables.length === 0" class="empty-section">
            No variables available
          </div>
        </div>
      </div>

      <!-- EVENT DISPATCHERS Section -->
      <div class="section">
        <div class="section-header" @click="toggleSection('eventDispatchers')">
          <div class="section-expand" :class="{ expanded: expandedSections.eventDispatchers }">
            {{ expandedSections.eventDispatchers ? '▼' : '▶' }}
          </div>
          <div class="section-title">EVENT DISPATCHERS</div>
          <div class="section-actions">
            <button class="section-action-btn" title="Add Event Dispatcher" @click.stop="showCreateEventDispatcherModal = true">+</button>
          </div>
        </div>

        <div v-if="expandedSections.eventDispatchers" class="section-content">
          <div
              v-for="event in filteredEventDispatchers"
              :key="event.id"
              class="section-item"
              :class="{ active: selectedItem === event.id }"
              @click="selectItem(event.id)"
              draggable="true"
              @dragstart="onEventDragStart($event, event)"
          >
            <div class="item-icon">⚡</div>
            <div class="item-name">{{ event.name }}</div>
          </div>

          <div v-if="filteredEventDispatchers.length === 0" class="empty-section">
            No event dispatchers available
          </div>
        </div>
      </div>
    </div>

    <!-- Create Variable Modal -->
    <ModalDialog
        v-if="showCreateVariableModal"
        title="Create Variable"
        @close="showCreateVariableModal = false"
        @confirm="createVariable"
    >
      <div class="modal-form">
        <div class="form-group">
          <label>Name:</label>
          <input v-model="newVariable.name" type="text" class="form-input" />
        </div>
        <div class="form-group">
          <label>Type:</label>
          <select v-model="newVariable.type" class="form-select">
            <option value="string">String</option>
            <option value="number">Number</option>
            <option value="boolean">Boolean</option>
            <option value="object">Object</option>
            <option value="array">Array</option>
          </select>
        </div>
        <div class="form-group">
          <label>Description:</label>
          <textarea v-model="newVariable.description" class="form-textarea"></textarea>
        </div>
      </div>
    </ModalDialog>

    <!-- Create Function Modal -->
    <ModalDialog
        v-if="showCreateFunctionModal"
        title="Create Function"
        @close="showCreateFunctionModal = false"
        @confirm="createFunction"
    >
      <div class="modal-form">
        <div class="form-group">
          <label>Name:</label>
          <input v-model="newFunction.name" type="text" class="form-input" />
        </div>
        <div class="form-group">
          <label>Return Type:</label>
          <select v-model="newFunction.returnType" class="form-select">
            <option value="void">Void</option>
            <option value="string">String</option>
            <option value="number">Number</option>
            <option value="boolean">Boolean</option>
            <option value="object">Object</option>
            <option value="array">Array</option>
          </select>
        </div>
        <div class="form-group">
          <label>Description:</label>
          <textarea v-model="newFunction.description" class="form-textarea"></textarea>
        </div>
      </div>
    </ModalDialog>

    <!-- Create Macro Modal -->
    <ModalDialog
        v-if="showCreateMacroModal"
        title="Create Macro"
        @close="showCreateMacroModal = false"
        @confirm="createMacro"
    >
      <div class="modal-form">
        <div class="form-group">
          <label>Name:</label>
          <input v-model="newMacro.name" type="text" class="form-input" />
        </div>
        <div class="form-group">
          <label>Description:</label>
          <textarea v-model="newMacro.description" class="form-textarea"></textarea>
        </div>
      </div>
    </ModalDialog>

    <!-- Create Event Dispatcher Modal -->
    <ModalDialog
        v-if="showCreateEventDispatcherModal"
        title="Create Event Dispatcher"
        @close="showCreateEventDispatcherModal = false"
        @confirm="createEventDispatcher"
    >
      <div class="modal-form">
        <div class="form-group">
          <label>Name:</label>
          <input v-model="newEventDispatcher.name" type="text" class="form-input" />
        </div>
        <div class="form-group">
          <label>Description:</label>
          <textarea v-model="newEventDispatcher.description" class="form-textarea"></textarea>
        </div>
      </div>
    </ModalDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject } from 'vue'
import { v4 as uuid } from 'uuid'
import ModalDialog from '../common/ModalDialog.vue'
import { useBlueprintStore } from '../../stores/blueprint'
import type { Variable, Blueprint } from '../../types/blueprint'

const emit = defineEmits<{
  (e: 'add-node', data: any): void
  (e: 'select-item', id: string, type: string): void
}>()

// Inject canvas container reference (passed from parent component)
const canvasContainer = inject<HTMLElement>('canvasContainer')

// Blueprint store
const blueprintStore = useBlueprintStore()

// State
const searchQuery = ref('')
const selectedItem = ref<string | null>(null)
const expandedSections = ref({
  graphs: true,
  functions: true,
  macros: true,
  variables: true,
  eventDispatchers: true
})

// Modal states
const showCreateVariableModal = ref(false)
const showCreateFunctionModal = ref(false)
const showCreateMacroModal = ref(false)
const showCreateEventDispatcherModal = ref(false)

// New item forms
const newVariable = ref({
  name: '',
  type: 'string',
  description: ''
})

const newFunction = ref({
  name: '',
  returnType: 'void',
  description: ''
})

const newMacro = ref({
  name: '',
  description: ''
})

const newEventDispatcher = ref({
  name: '',
  description: ''
})

// Filtered lists based on search query
const filteredGraphs = computed(() => {
  const graphs = [{ id: 'eventGraph', name: 'EventGraph' }]
  if (!searchQuery.value) return graphs

  return graphs.filter(g =>
      g.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const filteredFunctions = computed(() => {
  // Replace with actual function data from store
  const functions = [
    { id: 'function1', name: 'CalculateDistance' },
    { id: 'function2', name: 'ProcessInput' }
  ]

  if (!searchQuery.value) return functions

  return functions.filter(f =>
      f.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const filteredMacros = computed(() => {
  // Replace with actual macro data from store
  const macros = [
    { id: 'macro1', name: 'HandleMovement' }
  ]

  if (!searchQuery.value) return macros

  return macros.filter(m =>
      m.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const filteredVariables = computed(() => {
  return blueprintStore.variables.filter(v =>
      !searchQuery.value || v.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const filteredEventDispatchers = computed(() => {
  // Replace with actual event dispatcher data from store
  const eventDispatchers = [
    { id: 'event1', name: 'OnPlayerDeath' },
    { id: 'event2', name: 'OnItemCollected' }
  ]

  if (!searchQuery.value) return eventDispatchers

  return eventDispatchers.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Methods
function toggleSection(section: keyof typeof expandedSections.value) {
  expandedSections.value[section] = !expandedSections.value[section]
}

function selectItem(id: string) {
  selectedItem.value = id
  emit('select-item', id, getItemType(id))
}

function getItemType(id: string): string {
  if (filteredGraphs.value.some(g => g.id === id)) return 'graph'
  if (filteredFunctions.value.some(f => f.id === id)) return 'function'
  if (filteredMacros.value.some(m => m.id === id)) return 'macro'
  if (filteredVariables.value.some(v => v.id === id)) return 'variable'
  if (filteredEventDispatchers.value.some(e => e.id === id)) return 'eventDispatcher'
  return 'unknown'
}

function getVariableTypeColor(type: string): string {
  switch (type) {
    case 'string': return '#f0883e' // Orange
    case 'number': return '#6ed69a' // Green
    case 'boolean': return '#dc5050' // Red
    case 'object': return '#8ab4f8' // Blue
    case 'array': return '#bb86fc' // Purple
    default: return '#aaaaaa' // Gray
  }
}

function handleAddButtonClick() {
  // Show a dropdown or context menu for different "Add" options
  // For now, we'll just show the variable modal as an example
  showCreateVariableModal.value = true
}

// Drag handlers
function onVariableDragStart(event: DragEvent, variable: Variable) {
  if (!event.dataTransfer) return

  // Create a node representation of this variable
  const nodeData = {
    id: uuid(),
    type: `variable-${variable.type}`,
    position: { x: 0, y: 0 },
    properties: [
      { name: 'variableId', value: variable.id },
      { name: 'variableName', value: variable.name }
    ]
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

function onFunctionDragStart(event: DragEvent, func: any) {
  if (!event.dataTransfer) return

  // Create a node representation of this function
  const nodeData = {
    id: uuid(),
    type: 'function-call',
    position: { x: 0, y: 0 },
    properties: [
      { name: 'functionId', value: func.id },
      { name: 'functionName', value: func.name }
    ]
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

function onMacroDragStart(event: DragEvent, macro: any) {
  if (!event.dataTransfer) return

  // Create a node representation of this macro
  const nodeData = {
    id: uuid(),
    type: 'macro-call',
    position: { x: 0, y: 0 },
    properties: [
      { name: 'macroId', value: macro.id },
      { name: 'macroName', value: macro.name }
    ]
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

function onEventDragStart(event: DragEvent, eventDispatcher: any) {
  if (!event.dataTransfer) return

  // Create a node representation of this event
  const nodeData = {
    id: uuid(),
    type: 'event-dispatcher',
    position: { x: 0, y: 0 },
    properties: [
      { name: 'eventId', value: eventDispatcher.id },
      { name: 'eventName', value: eventDispatcher.name }
    ]
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

// Creation methods
function createVariable() {
  const newVar: Variable = {
    id: uuid(),
    name: newVariable.value.name,
    type: newVariable.value.type,
    value: getDefaultValueForType(newVariable.value.type)
  }

  blueprintStore.addVariable(newVar)
  showCreateVariableModal.value = false

  // Reset form
  newVariable.value = {
    name: '',
    type: 'string',
    description: ''
  }
}

function createFunction() {
  // Implement function creation in your store
  console.log('Creating function:', newFunction.value)
  showCreateFunctionModal.value = false

  // Reset form
  newFunction.value = {
    name: '',
    returnType: 'void',
    description: ''
  }
}

function createMacro() {
  // Implement macro creation in your store
  console.log('Creating macro:', newMacro.value)
  showCreateMacroModal.value = false

  // Reset form
  newMacro.value = {
    name: '',
    description: ''
  }
}

function createEventDispatcher() {
  // Implement event dispatcher creation in your store
  console.log('Creating event dispatcher:', newEventDispatcher.value)
  showCreateEventDispatcherModal.value = false

  // Reset form
  newEventDispatcher.value = {
    name: '',
    description: ''
  }
}

function getDefaultValueForType(type: string): any {
  switch (type) {
    case 'string': return ''
    case 'number': return 0
    case 'boolean': return false
    case 'object': return {}
    case 'array': return []
    default: return null
  }
}
</script>

<style scoped>
.blueprint-left-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252525;
  color: #e0e0e0;
  user-select: none;
  width: 100%;
}

.panel-search {
  padding: 10px;
  display: flex;
  gap: 8px;
  border-bottom: 1px solid #3d3d3d;
}

.search-container {
  position: relative;
  flex: 1;
}

.search-input {
  width: 100%;
  background-color: #3d3d3d;
  border: 1px solid #444;
  color: white;
  padding: 6px 8px;
  padding-right: 28px;
  border-radius: 4px;
  font-size: 0.9rem;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent-blue);
}

.clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: #777;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.add-btn {
  background-color: #444;
  border: none;
  color: white;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.9rem;
}

.add-btn:hover {
  background-color: #555;
}

.add-icon {
  font-weight: bold;
}

.panel-sections {
  flex: 1;
  overflow-y: auto;
}

.section {
  margin-bottom: 2px;
}

.section-header {
  display: flex;
  align-items: center;
  padding: 8px;
  background-color: #2d2d2d;
  cursor: pointer;
}

.section-header:hover {
  background-color: #383838;
}

.section-expand {
  width: 16px;
  text-align: center;
  color: #aaa;
  font-size: 10px;
}

.section-title {
  flex: 1;
  font-weight: 500;
  font-size: 0.85rem;
  color: #aaa;
}

.section-actions {
  display: flex;
  gap: 4px;
}

.section-action-btn {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  border: none;
  color: #aaa;
  cursor: pointer;
  font-size: 10px;
  border-radius: 2px;
}

.section-action-btn:hover {
  background-color: #444;
  color: white;
}

.section-content {
  padding: 4px 0;
}

.section-item {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  cursor: pointer;
  border-radius: 2px;
  margin: 1px 4px;
}

.section-item:hover {
  background-color: #3d3d3d;
}

.section-item.active {
  background-color: var(--accent-blue);
}

.item-icon {
  margin-right: 8px;
  font-size: 14px;
  width: 16px;
  text-align: center;
}

.item-name {
  flex: 1;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-type {
  font-size: 0.75rem;
  color: #aaa;
  margin-left: 8px;
}

.empty-section {
  padding: 8px;
  text-align: center;
  font-style: italic;
  color: #777;
  font-size: 0.85rem;
}

/* Variable-specific styles */
.variable-item {
  display: flex;
  align-items: center;
}

.variable-type-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
}

/* Modal form styles */
.modal-form {
  padding: 10px 0;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 8px;
  background-color: #333;
  border: 1px solid #444;
  border-radius: 4px;
  color: white;
  font-size: 0.9rem;
}

.form-textarea {
  min-height: 80px;
  resize: vertical;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--accent-blue);
}
</style>
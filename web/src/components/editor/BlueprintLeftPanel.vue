<!-- File: web/src/components/editor/BlueprintLeftPanel.vue -->
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
              @dblclick="openFunctionEditor(func)"
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
              @dragend="onVariableDragEnd($event, variable)"
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

      <!-- ENTRY POINT EVENTS Section -->
      <div class="section">
        <div class="section-header" @click="toggleSection('entryPointEvents')">
          <div class="section-expand" :class="{ expanded: expandedSections.entryPointEvents }">
            {{ expandedSections.entryPointEvents ? '▼' : '▶' }}
          </div>
          <div class="section-title">ENTRY POINT EVENTS</div>
          <div class="section-actions">
            <!-- No add button for entry point events as they're system-defined -->
          </div>
        </div>

        <div v-if="expandedSections.entryPointEvents" class="section-content">
          <div
              v-for="event in filteredEntryPointEvents"
              :key="event.id"
              class="section-item"
              :class="{ active: selectedItem === event.id }"
              @click="selectItem(event.id)"
              draggable="true"
              @dragstart="onEntryPointEventDragStart($event, event)"
          >
            <div class="item-icon">🔌</div>
            <div class="item-name">{{ event.name }}</div>
          </div>

          <div v-if="filteredEntryPointEvents.length === 0" class="empty-section">
            No entry point events available
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

    <!-- Variable Actions Menu (GET/SET) -->
    <div v-if="showVariableMenu" class="variable-menu" :style="variableMenuStyle">
      <div class="menu-item" @click="createGetVariableNode">
        Get {{ draggedVariable?.name }}
      </div>
      <div class="menu-item" @click="createSetVariableNode">
        Set {{ draggedVariable?.name }}
      </div>
    </div>

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
import { ref, computed, inject, onMounted } from 'vue'
import { v4 as uuid } from 'uuid'
import ModalDialog from '../common/ModalDialog.vue'
import { useBlueprintStore } from '../../stores/blueprint'
import type {Variable, Function, Node} from '../../types/blueprint'
import {useNodeRegistryStore} from "../../stores/nodeRegistry";
import {NodeTypeDefinition} from "../../types/nodes";
import router from "../../router";
import {useRoute} from "vue-router";
import {useWorkspaceStore} from "../../stores/workspace";
import {EventDefinition} from "../../stores/events";

const emit = defineEmits<{
  (e: 'add-node', data: any): void
  (e: 'select-item', id: string, type: string): void
  (e: 'function-double-clicked', functionId: string): void
}>()

const route = useRoute()
// Blueprint store
const blueprintStore = useBlueprintStore()
// Workspace store
const workspaceStore = useWorkspaceStore()
// Node Type store
const nodeTypeStore = useNodeRegistryStore()

// State
const searchQuery = ref('')
const selectedItem = ref<string | null>(null)
const expandedSections = ref({
  graphs: true,
  functions: true,
  macros: true,
  variables: true,
  entryPointEvents: true, // Added for entry point events
  eventDispatchers: true
})

// Modal states
const showCreateVariableModal = ref(false)
const showCreateFunctionModal = ref(false)
const showCreateMacroModal = ref(false)
const showCreateEventDispatcherModal = ref(false)

// Variable menu state
const showVariableMenu = ref(false)
const variableMenuPosition = ref({ x: 0, y: 0 })
const draggedVariable = ref<Variable | null>(null)
const dropPosition = ref({ x: 0, y: 0 })

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

// Add state for custom events from server
const customEvents = computed(() => {
  return blueprintStore.blueprint.events;
})

// Entry point events
const entryPointEvents = computed(() => {
  const _nodeTypes: Record<string, NodeTypeDefinition> = nodeTypeStore.nodeTypes
  const _entryNodes: EventDefinition[] = [];

  Object.values(_nodeTypes).forEach((_node) => {
    if (_node.category === 'Constructive Events') {
      _entryNodes.push({
        id: _node.typeId,
        name: _node.name,
        description: _node.description,
        type: 'entry',
        category: _node.category,
      })
    }
  })

  return _entryNodes;
})

// Computed values for variable menu positioning
const variableMenuStyle = computed(() => {
  return {
    left: `${variableMenuPosition.value.x}px`,
    top: `${variableMenuPosition.value.y}px`
  }
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
  if (!searchQuery.value) return blueprintStore.functions

  return blueprintStore.functions.filter(f =>
      f.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  ) || []
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
  // Ensure variables is an array
  const variables = blueprintStore.variables || [];
  
  // Log for debugging
  console.log('Filtering variables:', variables);
  
  return variables.filter(v =>
      !searchQuery.value || v.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const filteredEntryPointEvents = computed(() => {
  if (!searchQuery.value) return entryPointEvents.value

  return entryPointEvents.value.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const filteredEventDispatchers = computed(() => {
  // Combine default dispatchers with custom events from server
  const allDispatchers = [
    // Default built-in dispatchers
    { id: 'event-with-payload', name: 'Custom Event With Payload' },
    ...customEvents.value
  ]

  if (!searchQuery.value) return allDispatchers

  return allDispatchers.filter(e =>
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
  if (filteredEntryPointEvents.value.some(e => e.id === id)) return 'entryPointEvent'
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

// Handle variable drag start - just mark it as a variable drag
function onVariableDragStart(event: DragEvent, variable: Variable) {
  if (!event.dataTransfer) return;

  // Store variable info for the drag operation
  draggedVariable.value = variable;

  // CRITICAL: Set a completely different format type for variables
  // This prevents canvas from processing it as a regular node
  event.dataTransfer.setData('application/x-blueprint-variable', JSON.stringify(variable));

  // Don't set application/json at all for variables
  // This prevents the canvas from auto-creating nodes

  // Set appearance for the drag operation
  const dragImage = document.createElement('div');
  dragImage.textContent = variable.name;
  dragImage.style.backgroundColor = '#333';
  dragImage.style.color = 'white';
  dragImage.style.padding = '8px';
  dragImage.style.borderRadius = '4px';
  dragImage.style.position = 'absolute';
  dragImage.style.top = '-1000px';
  document.body.appendChild(dragImage);
  event.dataTransfer.setDragImage(dragImage, 0, 0);
  setTimeout(() => document.body.removeChild(dragImage), 0);

  event.dataTransfer.effectAllowed = 'copy';
}

// Handle variable drag end - show the Get/Set menu
function onVariableDragEnd(event: DragEvent, variable: Variable) {
  // Prevent default to avoid any node creation by the canvas drop handlers
  event.preventDefault()

  const canvas = document.querySelector('.blueprint-canvas');
  const canvasRect = canvas.getBoundingClientRect()
  const dropX = event.clientX - canvasRect.left
  const dropY = event.clientY - canvasRect.top

  // Store the canvas position for node creation
  dropPosition.value = { x: dropX, y: dropY }

  // Open the variable menu at the client (screen) position
  variableMenuPosition.value = { x: event.clientX, y: event.clientY }
  showVariableMenu.value = true

  // Add event listener to close the menu when clicking elsewhere
  document.addEventListener('click', closeVariableMenu, { once: true })

  // Prevent immediate propagation to avoid any other handlers
  event.stopPropagation()
}

// Close the variable action menu
function closeVariableMenu() {
  showVariableMenu.value = false
}

// Create a Get Variable node
function createGetVariableNode() {
  if (!draggedVariable.value) return;

  // Create a Get Variable node with the correct position
  const node = {
    id: uuid(),
    type: `get-variable-${draggedVariable.value.name}`,
    position: dropPosition.value,
    properties: [
      { name: 'variableId', value: draggedVariable.value.id },
      { name: 'input_name', value: draggedVariable.value.name},
      // Store variable type to help with type-specific handling
      { name: 'variableType', value: draggedVariable.value.type }
    ],
    // Add data to help with node rendering and connections
    data: {
      variableType: draggedVariable.value.type,
      isDataNode: true, // Flag to indicate this is a pure data node (no execution pins)
    }
  };

  console.log('Creating variable-get node for:', draggedVariable.value);
  // Add the node to the blueprint
  emit('add-node', node);

  // Close the menu and reset
  closeVariableMenu();
  draggedVariable.value = null;
}

// Create a Set Variable node
function createSetVariableNode() {
  if (!draggedVariable.value) return;

  // Create a Set Variable node with the correct position
  const node = {
    id: uuid(),
    type: `set-variable-${draggedVariable.value.name}`,
    position: dropPosition.value,
    properties: [
      { name: 'variableId', value: draggedVariable.value.id },
      { name: 'input_name', value: draggedVariable.value.name},
      // Store variable type to help with type-specific handling
      { name: 'variableType', value: draggedVariable.value.type }
    ],
    // Add data to help with node rendering and connections
    data: {
      variableType: draggedVariable.value.type,
      isDataNode: true, // Flag to indicate this is a pure data node (no execution pins)
    }
  };

  console.log('Creating variable-set node for:', draggedVariable.value);
  // Add the node to the blueprint
  emit('add-node', node);

  // Close the menu and reset
  closeVariableMenu();
  draggedVariable.value = null;
}

function onFunctionDragStart(event: DragEvent, func: Function) {
  if (!event.dataTransfer) return

  console.log('Dragging function:', func)

  // CRITICAL: Ensure the function's node type is properly registered
  // This is necessary when dragging functions after page refresh
  if (func.nodeType) {
    console.log('Ensuring function node type is registered:', func.nodeType);
    nodeTypeStore.registerNodeType(func.nodeType);
  }

  // Create a node representation of this function
  // IMPORTANT: The 'type' must match the nodeType.typeId that was registered
  // with the node registry, which is the function name
  const nodeData = {
    id: uuid(),
    type: func.name, // Use function name (not ID) as the type
    position: { x: 0, y: 0 },
    properties: [
      { name: 'functionId', value: func.id },
      { name: 'functionName', value: func.name }
    ]
  }

  console.log('Creating function node with type:', nodeData.type)
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

// Handle dragging an entry point event node
function onEntryPointEventDragStart(event: DragEvent, entryPoint: any) {
  if (!event.dataTransfer) return

  console.log('Dragging entry point event:', entryPoint)

  // Create a node representation of this entry point event
  const nodeData = {
    id: uuid(),
    type: entryPoint.id, // Use the event ID as the node type
    position: { x: 0, y: 0 },
    properties: [
      { name: 'eventType', value: entryPoint.type },
      { name: 'eventName', value: entryPoint.name }
    ]
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

// Handle dragging an event dispatcher node
function onEventDragStart(event: DragEvent, eventDispatcher: any) {
  if (!event.dataTransfer) return

  console.log('Dragging event dispatcher:', eventDispatcher)

  // Create a node representation of this event
  const nodeData = {
    id: uuid(),
    type: 'event-dispatcher',
    position: { x: 0, y: 0 },
    properties: [
      { name: 'eventId', value: eventDispatcher.id },
      { name: 'eventName', value: eventDispatcher.name },
      { name: 'dynamicParameters', value: true }
    ]
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

// Function to handle double-click on a function in the list
function openFunctionEditor(func: Function) {
  emit('function-double-clicked', func.id);
}

// Creation methods
async function createVariable() {
  const newVar: Variable = {
    id: uuid(),
    name: newVariable.value.name,
    type: newVariable.value.type,
    value: getDefaultValueForType(newVariable.value.type)
  }

  await blueprintStore.addVariable(workspaceStore.currentWorkspace.id, newVar)
  if (route.params.id !== blueprintStore.blueprint.id) {
    await router.push(`/editor/${blueprintStore.blueprint.id}`)
  }

  showCreateVariableModal.value = false

  // Reset form
  newVariable.value = {
    name: '',
    type: 'string',
    description: ''
  }
}

function createFunction() {
  // Create a new function with UUID
  const fnNodeType: NodeTypeDefinition = {
    typeId: newFunction.value.name,
    name: newFunction.value.name,
    description: '',
    category: 'Function',
    version: '1.0.0',
    inputs: [
      {
        description: 'Execution continues',
        id: 'exec',
        name: 'Execute',
        optional: false,
        type: {
          description: 'Controls execution flow',
          id: 'execution',
          name: 'Execution',
        }
      }
    ],
    outputs: [
      {
        description: 'Execution continues',
        id: 'then',
        name: 'Then',
        optional: false,
        type: {
          description: 'Controls execution flow',
          id: 'execution',
          name: 'Execution',
        }
      }
    ],
  }

  const fn: Function = {
    id: uuid(),
    name: newFunction.value.name,
    description: newFunction.value.description,
    nodeType: fnNodeType,
    nodes: [],
    connections: [],
    variables: [],
    metadata: {
      'type': 'function'
    }
  }

  const constructNode: Node = {
    id: uuid(),
    type: newFunction.value.name,
    position: {x: 100, y: 100},
    properties: []
  }

  // Register the node type with the node registry
  console.log('Registering function node type with typeId:', fnNodeType.typeId);
  nodeTypeStore.registerNodeType(fnNodeType)

  // Add the function to the blueprint store
  blueprintStore.addFunction(fn)
  
  // Add a constructor node for the function
  blueprintStore.addNodeToFunction(fn.id, constructNode)
  
  // Ensure the function node can be dragged and dropped properly
  const functionNode = nodeTypeStore.getNodeTypeById(fn.name);
  console.log('Registered function node type:', functionNode);

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

/**
 * Creates an event dispatcher by calling the server API
 */
async function createEventDispatcher() {
  try {
    // Validate input
    if (!newEventDispatcher.value.name) {
      alert("Event name is required");
      return;
    }
    
    // Get current blueprint ID if available
    const blueprintId = blueprintStore.blueprint?.id;
    
    // Prepare request
    const requestData = {
      id: `custom.${newEventDispatcher.value.name}`,
      blueprintId: blueprintId,
      name: newEventDispatcher.value.name,
      description: newEventDispatcher.value.description || "",
      category: "Custom Events",
      parameters: [],
    };
    
    console.log("Creating event dispatcher:", requestData, "for blueprint:", blueprintId);
    
    // // Call API to create event dispatcher
    // let url = '/api/events';
    // // If we have a blueprint ID, associate this event with it
    // if (blueprintId) {
    //   url += `?blueprintId=${blueprintId}`;
    // }
    //
    // const response = await fetch(url, {
    //   method: 'POST',
    //   headers: {
    //     'Content-Type': 'application/json'
    //   },
    //   body: JSON.stringify(requestData)
    // });
    //
    // if (!response.ok) {
    //   throw new Error(`Failed to create event: ${response.statusText}`);
    // }
    //
    // // Parse response
    // const createdEvent = await response.json();
    // console.log('Created event dispatcher:', createdEvent);

    blueprintStore.addEvent(requestData);
    
    // Close modal and reset form
    showCreateEventDispatcherModal.value = false;
    newEventDispatcher.value = {
      name: '',
      description: ''
    };

    
  } catch (error) {
    console.error('Error creating event dispatcher:', error);
    alert(`Failed to create event dispatcher: ${error.message}`);
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

.clear-btn:hover {
  color: white;
  background-color: rgba(255, 255, 255, 0.1);
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

/* Variable menu styles - similar to Unreal's context menu */
.variable-menu {
  position: fixed;
  z-index: 1000;
  background-color: #333333;
  border: 1px solid #555555;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  min-width: 180px;
  animation: menu-appear 0.15s ease-out;
}

@keyframes menu-appear {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}

.menu-item {
  padding: 8px 16px;
  font-size: 0.9rem;
  color: #e0e0e0;
  cursor: pointer;
}

.menu-item:hover {
  background-color: var(--accent-blue);
  color: white;
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

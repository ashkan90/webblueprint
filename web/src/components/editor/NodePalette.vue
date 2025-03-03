<template>
  <div class="node-palette">
    <div class="search-container">
      <input
          v-model="searchQuery"
          type="text"
          placeholder="Search nodes..."
          class="search-input"
      />
    </div>

    <div v-if="isLoading" class="loading">
      Loading node types...
    </div>

    <div v-else-if="error" class="error">
      {{ error }}
    </div>

    <div v-else class="categories">
      <div v-for="category in filteredCategories" :key="category" class="category">
        <div
            class="category-header"
            @click="toggleCategory(category)"
            :class="{ 'collapsed': collapsedCategories.includes(category) }"
        >
          <span class="category-name">{{ category }}</span>
          <span class="toggle-icon">{{ collapsedCategories.includes(category) ? '▶' : '▼' }}</span>
        </div>

        <div v-if="!collapsedCategories.includes(category)" class="category-items">
          <div
              v-for="nodeType in nodeTypesInCategory(category)"
              :key="nodeType.typeId"
              class="node-type-item"
              draggable="true"
              @dragstart="onDragStart($event, nodeType)"
              @dragend="onDragEnd($event)"
          >
            <div class="node-type-name">{{ nodeType.name }}</div>
            <div class="node-type-description">{{ nodeType.description }}</div>
          </div>
        </div>
      </div>

      <div v-if="filteredCategories.length === 0" class="no-results">
        No nodes match your search query.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject } from 'vue'
import { v4 as uuid } from 'uuid'
import { useNodeRegistryStore } from '../../stores/nodeRegistry'
import type { NodeTypeDefinition } from '../../types/nodes'
import type { Node } from '../../types/blueprint'

const emit = defineEmits<{
  (e: 'node-added', node: Node): void
}>()

const nodeRegistryStore = useNodeRegistryStore()

// Inject canvas reference
const canvasContainer = inject<HTMLElement>('canvasContainer')

// State
const searchQuery = ref('')
const collapsedCategories = ref<string[]>([])
const draggedNode = ref<NodeTypeDefinition | null>(null)

// Computed
const isLoading = computed(() => nodeRegistryStore.isLoading)
const error = computed(() => nodeRegistryStore.error)
const allCategories = computed(() => nodeRegistryStore.categories)

const filteredCategories = computed(() => {
  if (!searchQuery.value) {
    return allCategories.value
  }

  const query = searchQuery.value.toLowerCase()

  return allCategories.value.filter(category => {
    // Check if category matches
    if (category.toLowerCase().includes(query)) {
      return true
    }

    // Check if any node in the category matches
    const nodeTypes = nodeRegistryStore.nodeTypesByCategory[category] || []
    return nodeTypes.some(nodeType =>
        nodeType.name.toLowerCase().includes(query) ||
        nodeType.description.toLowerCase().includes(query)
    )
  })
})

// Methods
function nodeTypesInCategory(category: string) {
  const query = searchQuery.value.toLowerCase()
  const nodeTypes = nodeRegistryStore.nodeTypesByCategory[category] || []

  if (!query) {
    return nodeTypes
  }

  return nodeTypes.filter(nodeType =>
      nodeType.name.toLowerCase().includes(query) ||
      nodeType.description.toLowerCase().includes(query)
  )
}

function toggleCategory(category: string) {
  const index = collapsedCategories.value.indexOf(category)
  if (index === -1) {
    collapsedCategories.value.push(category)
  } else {
    collapsedCategories.value.splice(index, 1)
  }
}

function onDragStart(event: DragEvent, nodeType: NodeTypeDefinition) {
  // Create a new node instance with complete data
  const node = {
    id: uuid(),
    type: nodeType.typeId,
    position: { x: 0, y: 0 },
    properties: []  // Ensure properties is initialized
  };

  // Set drag data
  if (event.dataTransfer) {
    // Set the data in proper JSON format
    event.dataTransfer.setData('application/json', JSON.stringify(node));
    event.dataTransfer.effectAllowed = 'copy';

    // Add a visual indicator for what's being dragged
    const dragImage = document.createElement('div');
    dragImage.textContent = nodeType.name;
    dragImage.style.backgroundColor = '#333';
    dragImage.style.color = 'white';
    dragImage.style.padding = '8px';
    dragImage.style.borderRadius = '4px';
    dragImage.style.position = 'absolute';
    dragImage.style.top = '-1000px'; // Hide it initially
    document.body.appendChild(dragImage);

    event.dataTransfer.setDragImage(dragImage, 0, 0);

    // Clean up the element after drag
    setTimeout(() => document.body.removeChild(dragImage), 0);
  }
}

function onDragEnd(event: DragEvent) {
  if (!draggedNode.value || !canvasContainer) return

  // Calculate drop position relative to canvas
  const canvasRect = canvasContainer.getBoundingClientRect()
  const dropX = event.clientX - canvasRect.left
  const dropY = event.clientY - canvasRect.top

  // Create a new node instance
  const node: Node = {
    id: uuid(),
    type: draggedNode.value.typeId,
    position: { x: dropX, y: dropY },
    properties: []
  }

  // Emit node added event
  emit('node-added', node)

  // Reset dragged node
  draggedNode.value = null
}
</script>

<style scoped>
.node-palette {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-y: auto;
}

.search-container {
  padding: 8px;
  border-bottom: 1px solid #3d3d3d;
  background-color: #2d2d2d;
  position: sticky;
  top: 0;
  z-index: 10;
}

.search-input {
  width: 100%;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #444;
  background-color: #333;
  color: white;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent-blue);
}

.categories {
  flex: 1;
}

.category {
  border-bottom: 1px solid #3d3d3d;
}

.category-header {
  padding: 8px 12px;
  background-color: #333;
  cursor: pointer;
  user-select: none;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.category-header:hover {
  background-color: #444;
}

.category-items {
  padding: 8px 0;
}

.node-type-item {
  padding: 8px 12px;
  border-radius: 4px;
  margin: 4px 8px;
  background-color: var(--node-bg);
  cursor: grab;
  transition: background-color 0.2s;
}

.node-type-item:hover {
  background-color: var(--node-header);
}

.node-type-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.node-type-description {
  font-size: 0.8rem;
  color: #aaa;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.loading, .error, .no-results {
  padding: 16px;
  text-align: center;
  color: #aaa;
}

.error {
  color: var(--accent-red);
}

.node-type-item {
  cursor: grab;
}

.node-type-item:active {
  cursor: grabbing;
}
</style>
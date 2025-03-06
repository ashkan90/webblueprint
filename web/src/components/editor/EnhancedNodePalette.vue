<!-- File: web/src/components/editor/EnhancedNodePalette.vue -->
<template>
  <div class="enhanced-node-palette">
    <div class="palette-header">
      <div class="search-container">
        <input
            v-model="searchQuery"
            type="text"
            placeholder="Search nodes..."
            class="search-input"
        />
        <button @click="clearSearch" class="clear-btn" v-if="searchQuery">
          ✕
        </button>
      </div>

      <div class="view-toggles">
        <button
            class="toggle-btn"
            :class="{ active: viewMode === 'categories' }"
            @click="viewMode = 'categories'"
            title="Category View"
        >
          <span class="icon">📁</span>
        </button>
        <button
            class="toggle-btn"
            :class="{ active: viewMode === 'favorites' }"
            @click="viewMode = 'favorites'"
            title="Favorites"
        >
          <span class="icon">⭐</span>
        </button>
        <button
            class="toggle-btn"
            :class="{ active: viewMode === 'recent' }"
            @click="viewMode = 'recent'"
            title="Recent"
        >
          <span class="icon">🕒</span>
        </button>
      </div>
    </div>

    <div class="node-palette-content">
      <!-- Loading state -->
      <div v-if="isLoading" class="loading-container">
        <div class="loading-spinner"></div>
        <div class="loading-text">Loading node types...</div>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="error-container">
        <div class="error-icon">⚠️</div>
        <div class="error-message">{{ error }}</div>
        <button @click="retryFetch" class="retry-btn">Retry</button>
      </div>

      <!-- Categories view -->
      <div v-else-if="viewMode === 'categories'" class="categories-view">
        <div
            v-for="category in filteredCategories"
            :key="category"
            class="category-container"
        >
          <div
              class="category-header"
              @click="toggleCategory(category)"
              :class="{ 'collapsed': collapsedCategories.includes(category) }"
          >
            <span class="toggle-icon">{{ collapsedCategories.includes(category) ? '▶' : '▼' }}</span>
            <span class="category-name">{{ category }}</span>
            <span class="item-count">({{ nodeTypesInCategory(category).length }})</span>
          </div>

          <div v-if="!collapsedCategories.includes(category)" class="category-items">
            <div
                v-for="nodeType in nodeTypesInCategory(category)"
                :key="nodeType.typeId"
                class="node-type-item"
                draggable="true"
                @dragstart="onDragStart($event, nodeType)"
                @dragend="onDragEnd($event)"
                @contextmenu.prevent="showNodeContextMenu($event, nodeType)"
            >
              <div class="item-header">
                <span class="node-type-name">{{ nodeType.name }}</span>
                <button
                    class="favorite-btn"
                    @click.stop="toggleFavorite(nodeType.typeId)"
                    :title="isFavorite(nodeType.typeId) ? 'Remove from favorites' : 'Add to favorites'"
                >
                  <span class="favorite-icon" :class="{ 'active': isFavorite(nodeType.typeId) }">⭐</span>
                </button>
              </div>
              <div class="node-type-description">{{ nodeType.description }}</div>
            </div>
          </div>
        </div>

        <!-- No results -->
        <div v-if="filteredCategories.length === 0" class="no-results">
          <div class="no-results-icon">🔍</div>
          <div class="no-results-text">No nodes match your search query.</div>
          <button @click="clearSearch" class="clear-search-btn">Clear Search</button>
        </div>
      </div>

      <!-- Favorites view -->
      <div v-else-if="viewMode === 'favorites'" class="favorites-view">
        <div v-if="favoriteNodes.length > 0" class="favorites-list">
          <div
              v-for="nodeType in favoriteNodes"
              :key="nodeType.typeId"
              class="node-type-item"
              draggable="true"
              @dragstart="onDragStart($event, nodeType)"
              @dragend="onDragEnd($event)"
          >
            <div class="item-header">
              <span class="node-type-name">{{ nodeType.name }}</span>
              <button
                  class="favorite-btn"
                  @click.stop="toggleFavorite(nodeType.typeId)"
                  title="Remove from favorites"
              >
                <span class="favorite-icon active">⭐</span>
              </button>
            </div>
            <div class="node-type-category">{{ nodeType.category }}</div>
            <div class="node-type-description">{{ nodeType.description }}</div>
          </div>
        </div>

        <div v-else class="empty-favorites">
          <div class="empty-icon">⭐</div>
          <div class="empty-text">No favorite nodes yet.</div>
          <div class="empty-hint">Click the star icon on any node to add it to your favorites.</div>
        </div>
      </div>

      <!-- Recent view -->
      <div v-else-if="viewMode === 'recent'" class="recent-view">
        <div v-if="recentNodes.length > 0" class="recent-list">
          <div
              v-for="nodeType in recentNodes"
              :key="nodeType.typeId"
              class="node-type-item"
              draggable="true"
              @dragstart="onDragStart($event, nodeType)"
              @dragend="onDragEnd($event)"
          >
            <div class="item-header">
              <span class="node-type-name">{{ nodeType.name }}</span>
              <button
                  class="favorite-btn"
                  @click.stop="toggleFavorite(nodeType.typeId)"
                  :title="isFavorite(nodeType.typeId) ? 'Remove from favorites' : 'Add to favorites'"
              >
                <span class="favorite-icon" :class="{ 'active': isFavorite(nodeType.typeId) }">⭐</span>
              </button>
            </div>
            <div class="node-type-category">{{ nodeType.category }}</div>
            <div class="node-type-description">{{ nodeType.description }}</div>
          </div>
        </div>

        <div v-else class="empty-recent">
          <div class="empty-icon">🕒</div>
          <div class="empty-text">No recently used nodes.</div>
          <div class="empty-hint">Nodes you use will appear here for quick access.</div>
        </div>
      </div>
    </div>

    <!-- Context menu for node actions -->
    <div v-if="showContextMenu" class="node-context-menu" :style="contextMenuStyle">
      <div class="context-menu-item" @click="addSelectedNodeToCanvas">
        <span class="context-icon">➕</span> Add to Canvas
      </div>
      <div class="context-menu-item" @click="toggleFavorite(selectedNodeType?.typeId)">
        <span class="context-icon">{{ isFavorite(selectedNodeType?.typeId) ? '★' : '☆' }}</span>
        {{ isFavorite(selectedNodeType?.typeId) ? 'Remove from Favorites' : 'Add to Favorites' }}
      </div>
      <div class="context-menu-divider"></div>
      <div class="context-menu-item" @click="showNodeInfo">
        <span class="context-icon">ℹ️</span> Node Information
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, onMounted } from 'vue'
import { v4 as uuid } from 'uuid'
import { useNodeRegistryStore } from '../../stores/nodeRegistry'
import type { NodeTypeDefinition } from '../../types/nodes'
import type { Node } from '../../types/blueprint'

const emit = defineEmits<{
  (e: 'node-added', node: Node): void
}>()

const nodeRegistryStore = useNodeRegistryStore()

// DOM Refs
const canvasContainer = inject<HTMLElement>('canvasContainer')

// State
const searchQuery = ref('')
const collapsedCategories = ref<string[]>([])
const viewMode = ref<'categories' | 'favorites' | 'recent'>('categories')
const draggedNodeType = ref<NodeTypeDefinition | null>(null)
const showContextMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const selectedNodeType = ref<NodeTypeDefinition | null>(null)

// User preferences
const favoriteNodeIds = ref<string[]>([])
const recentNodeIds = ref<string[]>([])

// Computed properties
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

const contextMenuStyle = computed(() => {
  return {
    left: `${contextMenuPosition.value.x}px`,
    top: `${contextMenuPosition.value.y}px`
  }
})

const favoriteNodes = computed(() => {
  return favoriteNodeIds.value
      .map(id => nodeRegistryStore.getNodeTypeById(id))
      .filter((nodeType): nodeType is NodeTypeDefinition => nodeType !== null)
})

const recentNodes = computed(() => {
  return recentNodeIds.value
      .map(id => nodeRegistryStore.getNodeTypeById(id))
      .filter((nodeType): nodeType is NodeTypeDefinition => nodeType !== null)
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

function clearSearch() {
  searchQuery.value = ''
}

function retryFetch() {
  nodeRegistryStore.fetchNodeTypes()
}

function onDragStart(event: DragEvent, nodeType: NodeTypeDefinition) {
  draggedNodeType.value = nodeType

  // Add to recent nodes
  addToRecentNodes(nodeType.typeId)

  // Create a new node instance with complete data
  const node = {
    id: uuid(),
    type: nodeType.typeId,
    position: { x: 0, y: 0 },
    properties: []
  }

  // Set drag data
  if (event.dataTransfer) {
    // Set the data in proper JSON format
    event.dataTransfer.setData('application/json', JSON.stringify(node))
    event.dataTransfer.effectAllowed = 'copy'

    // Add a visual indicator for what's being dragged
    const dragImage = document.createElement('div')
    dragImage.textContent = nodeType.name
    dragImage.style.backgroundColor = '#333'
    dragImage.style.color = 'white'
    dragImage.style.padding = '8px'
    dragImage.style.borderRadius = '4px'
    dragImage.style.position = 'absolute'
    dragImage.style.top = '-1000px' // Hide it initially
    document.body.appendChild(dragImage)

    event.dataTransfer.setDragImage(dragImage, 0, 0)

    // Clean up the element after drag
    setTimeout(() => document.body.removeChild(dragImage), 0)
  }
}

function onDragEnd(event: DragEvent) {
  draggedNodeType.value = null
}

function toggleFavorite(nodeTypeId?: string) {
  if (!nodeTypeId) return

  const index = favoriteNodeIds.value.indexOf(nodeTypeId)
  if (index === -1) {
    // Add to favorites
    favoriteNodeIds.value.push(nodeTypeId)
  } else {
    // Remove from favorites
    favoriteNodeIds.value.splice(index, 1)
  }

  // Save to localStorage
  localStorage.setItem('favoriteNodeIds', JSON.stringify(favoriteNodeIds.value))

  // Close context menu if open
  showContextMenu.value = false
}

function isFavorite(nodeTypeId?: string): boolean {
  if (!nodeTypeId) return false
  return favoriteNodeIds.value.includes(nodeTypeId)
}

function addToRecentNodes(nodeTypeId: string) {
  // Remove if already exists to move it to the top
  const index = recentNodeIds.value.indexOf(nodeTypeId)
  if (index !== -1) {
    recentNodeIds.value.splice(index, 1)
  }

  // Add to the beginning
  recentNodeIds.value.unshift(nodeTypeId)

  // Keep only the last 10
  if (recentNodeIds.value.length > 10) {
    recentNodeIds.value = recentNodeIds.value.slice(0, 10)
  }

  // Save to localStorage
  localStorage.setItem('recentNodeIds', JSON.stringify(recentNodeIds.value))
}

function showNodeContextMenu(event: MouseEvent, nodeType: NodeTypeDefinition) {
  showContextMenu.value = true
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  selectedNodeType.value = nodeType

  // Add event listener to close menu on click outside
  setTimeout(() => {
    document.addEventListener('click', closeContextMenu, { once: true })
  }, 0)
}

function closeContextMenu() {
  showContextMenu.value = false
}

function addSelectedNodeToCanvas() {
  if (selectedNodeType.value) {
    // Create a new node instance
    const node: Node = {
      id: uuid(),
      type: selectedNodeType.value.typeId,
      position: { x: 100, y: 100 }, // Default position, can be adjusted
      properties: []
    }

    emit('node-added', node)
    addToRecentNodes(selectedNodeType.value.typeId)
  }

  showContextMenu.value = false
}

function showNodeInfo() {
  // This could open a modal with detailed node information
  // For now, just log to console
  console.log('Node info:', selectedNodeType.value)
  showContextMenu.value = false
}

// Initialize
onMounted(() => {
  // Load saved favorites and recent nodes from localStorage
  try {
    const savedFavorites = localStorage.getItem('favoriteNodeIds')
    if (savedFavorites) {
      favoriteNodeIds.value = JSON.parse(savedFavorites)
    }

    const savedRecent = localStorage.getItem('recentNodeIds')
    if (savedRecent) {
      recentNodeIds.value = JSON.parse(savedRecent)
    }
  } catch (e) {
    console.error('Error loading saved node preferences:', e)
  }
})
</script>

<style scoped>
.enhanced-node-palette {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #2d2d2d;
  overflow: hidden;
}

.palette-header {
  padding: 10px;
  background-color: #333;
  border-bottom: 1px solid #444;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
}

.search-input {
  width: 100%;
  padding: 8px 30px 8px 10px;
  border-radius: 4px;
  border: 1px solid #555;
  background-color: #3d3d3d;
  color: white;
  font-size: 0.9rem;
}

.search-input:focus {
  border-color: var(--accent-blue);
  outline: none;
}

.clear-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
}

.clear-btn:hover {
  color: white;
  background-color: rgba(255, 255, 255, 0.1);
}

.view-toggles {
  display: flex;
  justify-content: space-between;
}

.toggle-btn {
  flex: 1;
  background-color: #333;
  border: 1px solid #555;
  color: #aaa;
  padding: 5px 0;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn:first-child {
  border-radius: 4px 0 0 4px;
}

.toggle-btn:last-child {
  border-radius: 0 4px 4px 0;
}

.toggle-btn:hover {
  background-color: #444;
  color: white;
}

.toggle-btn.active {
  background-color: var(--accent-blue);
  color: white;
  border-color: var(--accent-blue);
}

.node-palette-content {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.loading-container, .error-container, .no-results, .empty-favorites, .empty-recent {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  text-align: center;
  height: 200px;
  color: #aaa;
}

.loading-spinner {
  width: 30px;
  height: 30px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--accent-blue);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text, .error-message, .no-results-text, .empty-text {
  margin-bottom: 10px;
}

.error-icon, .no-results-icon, .empty-icon {
  font-size: 2rem;
  margin-bottom: 15px;
}

.error-message {
  color: var(--accent-red);
}

.retry-btn, .clear-search-btn {
  background-color: var(--accent-blue);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 10px;
}

.retry-btn:hover, .clear-search-btn:hover {
  background-color: #2980b9;
}

.empty-hint {
  font-size: 0.8rem;
  color: #777;
  margin-top: 5px;
}

.category-container {
  margin-bottom: 10px;
}

.category-header {
  background-color: #333;
  padding: 8px 10px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background-color 0.2s;
}

.category-header:hover {
  background-color: #444;
}

.toggle-icon {
  width: 16px;
  font-size: 0.7rem;
  color: #aaa;
  transition: transform 0.2s;
}

.category-header.collapsed .toggle-icon {
  transform: rotate(-90deg);
}

.category-name {
  flex: 1;
  margin-left: 5px;
  font-weight: 500;
}

.item-count {
  color: #aaa;
  font-size: 0.8rem;
}

.category-items {
  padding: 5px 0 5px 10px;
}

.node-type-item {
  padding: 8px 10px;
  border-radius: 4px;
  margin-bottom: 5px;
  background-color: #3d3d3d;
  cursor: grab;
  transition: all 0.2s;
}

.node-type-item:hover {
  background-color: #4d4d4d;
  transform: translateY(-1px);
}

.node-type-item:active {
  cursor: grabbing;
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
}

.node-type-name {
  font-weight: 500;
}

.favorite-btn {
  background: none;
  border: none;
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  transition: all 0.2s;
}

.favorite-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.favorite-icon.active {
  color: #f9d71c;
  text-shadow: 0 0 10px rgba(249, 215, 28, 0.6);
}

.node-type-description {
  font-size: 0.8rem;
  color: #aaa;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-type-category {
  font-size: 0.75rem;
  color: #777;
  margin-bottom: 3px;
}

.node-context-menu {
  position: fixed;
  background-color: #333;
  border-radius: 4px;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.3);
  z-index: 1000;
  min-width: 200px;
  animation: menu-appear 0.15s ease-out;
}

@keyframes menu-appear {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}

.context-menu-item {
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
}

.context-menu-item:hover {
  background-color: #444;
}

.context-icon {
  margin-right: 8px;
  width: 16px;
  text-align: center;
}

.context-menu-divider {
  height: 1px;
  background-color: #555;
  margin: 5px 0;
}

/* Prevent Firefox drag ghost image */
.node-type-item:active {
  opacity: 0.7;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .toggle-btn .icon {
    margin-right: 0;
  }

  .toggle-btn span:not(.icon) {
    display: none;
  }
}
</style>

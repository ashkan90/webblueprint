<template>
  <div class="bottom-drawer" :class="{ 'collapsed': isCollapsed }" :style="drawerStyle">
    <div class="drawer-tabs">
      <div 
        v-for="tab in tabs" 
        :key="tab.id"
        class="drawer-tab"
        :class="{ 'active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </div>
      <div class="drawer-controls">
        <button class="drawer-control-btn" @click="toggleCollapse">
          <span class="icon">{{ isCollapsed ? '⬆' : '⬇' }}</span>
        </button>
      </div>
    </div>
    
    <div v-if="!isCollapsed" class="drawer-content" ref="drawerContent">
      <div class="drawer-resize-handle" @mousedown="startResize"></div>
      
      <div v-show="activeTab === 'content'" class="drawer-panel">
        <slot name="content"></slot>
      </div>
      
      <div v-show="activeTab === 'versions'" class="drawer-panel">
        <slot name="versions"></slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';

// Define tabs
const tabs = [
  { id: 'content', label: 'Content' },
  { id: 'versions', label: 'Versions' }
];

// State
const activeTab = ref('content');
const isCollapsed = ref(false);
const drawerHeight = ref(300); // Default height
const minHeight = 150; // Minimum drawer height
const maxHeight = 600; // Maximum drawer height
const drawerContent = ref<HTMLElement | null>(null);
const isResizing = ref(false);
const startY = ref(0);
const startHeight = ref(0);

// Computed
const drawerStyle = computed(() => {
  if (isCollapsed.value) {
    return { height: '40px' }; // Just enough for the tabs
  }
  return { height: `${drawerHeight.value}px` };
});

// Methods
function setActiveTab(tabId: string) {
  if (isCollapsed.value) {
    isCollapsed.value = false;
  } else if (activeTab.value === tabId) {
    isCollapsed.value = true;
  }
  
  activeTab.value = tabId;
}

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value;
}

function startResize(event: MouseEvent) {
  isResizing.value = true;
  startY.value = event.clientY;
  startHeight.value = drawerHeight.value;
  
  // Add move and up listeners to document
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
}

function onMouseMove(event: MouseEvent) {
  if (!isResizing.value) return;
  
  // Calculate new height (remember we're resizing from the top)
  const diff = startY.value - event.clientY;
  let newHeight = startHeight.value + diff;
  
  // Apply limits
  newHeight = Math.max(minHeight, Math.min(maxHeight, newHeight));
  
  // Update height
  drawerHeight.value = newHeight;
}

function onMouseUp() {
  isResizing.value = false;
  
  // Remove move and up listeners
  document.removeEventListener('mousemove', onMouseMove);
  document.removeEventListener('mouseup', onMouseUp);
}

// Lifecycle
onMounted(() => {
  // Save initial height if provided from localStorage
  const savedHeight = localStorage.getItem('drawerHeight');
  if (savedHeight) {
    drawerHeight.value = parseInt(savedHeight, 10);
  }
});

onUnmounted(() => {
  // Save height to localStorage
  localStorage.setItem('drawerHeight', drawerHeight.value.toString());
  
  // Clean up any listeners
  document.removeEventListener('mousemove', onMouseMove);
  document.removeEventListener('mouseup', onMouseUp);
});
</script>

<style scoped>
.bottom-drawer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: #2d2d2d;
  border-top: 1px solid #3d3d3d;
  z-index: 100;
  transition: height 0.2s ease;
  display: flex;
  flex-direction: column;
}

.drawer-tabs {
  display: flex;
  background-color: #333;
  border-bottom: 1px solid #3d3d3d;
  height: 40px;
  flex-shrink: 0;
}

.drawer-tab {
  padding: 0 16px;
  display: flex;
  align-items: center;
  color: #aaa;
  font-size: 0.9rem;
  cursor: pointer;
  border-right: 1px solid #3d3d3d;
  transition: background-color 0.2s;
}

.drawer-tab:hover {
  background-color: #444;
  color: white;
}

.drawer-tab.active {
  background-color: #3a8cd7;
  color: white;
}

.drawer-controls {
  margin-left: auto;
  display: flex;
  align-items: center;
}

.drawer-control-btn {
  background: none;
  border: none;
  color: #aaa;
  font-size: 1rem;
  padding: 0 12px;
  cursor: pointer;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.drawer-control-btn:hover {
  background-color: #444;
  color: white;
}

.drawer-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.drawer-resize-handle {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  cursor: ns-resize;
  background-color: transparent;
  z-index: 10;
}

.drawer-resize-handle:hover,
.drawer-resize-handle:active {
  background-color: rgba(58, 140, 215, 0.3);
}

.drawer-panel {
  height: 100%;
  overflow: auto;
}
</style>

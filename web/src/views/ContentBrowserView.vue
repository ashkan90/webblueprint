<template>
  <div class="content-browser-view">
    <div class="content-browser-header">
      <div class="path-breadcrumb">
        <span
          v-for="(segment, index) in pathSegments"
          :key="index"
          class="breadcrumb-segment"
          @click="navigateToBreadcrumb(index)"
        >
          {{ segment }}
          <span v-if="index < pathSegments.length - 1" class="separator">/</span>
        </span>
      </div>
      <div class="content-browser-toolbar">
        <button @click="navigateUp" :disabled="currentPath === '/'">
          <span class="icon">↑</span>
        </button>
        <button @click="refresh">
          <span class="icon">↻</span>
        </button>
        <input type="text" v-model="searchTerm" placeholder="Search..." @input="performSearch" />
      </div>
    </div>

    <div class="content-browser-layout">
      <div class="content-browser-sidebar">
        <div class="folder-tree">
          <div 
            class="folder-tree-item root" 
            :class="{ active: currentPath === '/' }"
            @click="navigateToFolder('/')"
          >
            <span class="folder-icon">📁</span>
            <span class="folder-name">Content</span>
          </div>
          
          <div 
            v-for="folder in rootFolders" 
            :key="folder.id" 
            class="folder-tree-item"
            :class="{ active: currentPath === folder.path }"
            @click="navigateToFolder(folder.path)"
          >
            <span class="folder-icon">📁</span>
            <span class="folder-name">{{ folder.name }}</span>
          </div>
        </div>
      </div>

      <div class="content-browser-content" @contextmenu.prevent="showFolderContextMenu">
        <div class="folder-grid">
          <div 
            v-for="folder in folders" 
            :key="folder.id" 
            class="folder-item"
            @click="navigateToFolder(folder.path)"
            @contextmenu.prevent="showItemContextMenu($event, folder, 'folder')"
            @dblclick="navigateToFolder(folder.path)"
          >
            <div class="folder-icon">📁</div>
            <div class="folder-name">{{ folder.name }}</div>
          </div>
        </div>

        <div class="asset-grid">
          <div 
            v-for="asset in assets" 
            :key="asset.id" 
            class="asset-item"
            :class="{ selected: selectedAsset?.id === asset.id }"
            @click="selectAsset(asset)"
            @dblclick="openAsset(asset)"
            @contextmenu.prevent="showItemContextMenu($event, asset, 'asset')"
          >
            <div class="asset-icon" :class="asset.type">
              <span v-if="asset.type === 'blueprint'">📄</span>
            </div>
            <div class="asset-name">{{ asset.name }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Context Menu for Folders and Assets -->
    <div 
      v-if="contextMenu.visible" 
      class="context-menu"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
    >
      <div v-if="contextMenu.type === 'folder'">
        <div class="context-menu-item" @click="openContextMenuItem('open')">Open</div>
        <div class="context-menu-item" @click="openContextMenuItem('rename')">Rename</div>
        <div class="context-menu-item" @click="openContextMenuItem('delete')">Delete</div>
      </div>
      
      <div v-else-if="contextMenu.type === 'asset'">
        <div class="context-menu-item" @click="openContextMenuItem('open')">Open</div>
        <div class="context-menu-item" @click="openContextMenuItem('rename')">Rename</div>
        <div class="context-menu-item" @click="openContextMenuItem('duplicate')">Duplicate</div>
        <div class="context-menu-item" @click="openContextMenuItem('delete')">Delete</div>
      </div>
      
      <div v-else-if="contextMenu.type === 'background'">
        <div class="context-menu-item" @click="openContextMenuItem('refresh')">Refresh</div>
        <div class="context-menu-item" @click="openContextMenuItem('newFolder')">New Folder</div>
        <div class="context-menu-item" @click="openContextMenuItem('newBlueprint')">New Blueprint</div>
      </div>
    </div>

    <!-- Rename Dialog -->
    <div v-if="renameDialog.visible" class="modal-backdrop">
      <div class="modal-dialog">
        <h3>Rename {{ renameDialog.type === 'folder' ? 'Folder' : 'Asset' }}</h3>
        <input type="text" v-model="renameDialog.newName" @keyup.enter="confirmRename" />
        <div class="modal-actions">
          <button @click="cancelRename">Cancel</button>
          <button @click="confirmRename">Rename</button>
        </div>
      </div>
    </div>

    <!-- New Folder Dialog -->
    <div v-if="newFolderDialog.visible" class="modal-backdrop">
      <div class="modal-dialog">
        <h3>New Folder</h3>
        <input type="text" v-model="newFolderDialog.name" placeholder="Folder Name" @keyup.enter="confirmNewFolder" />
        <div class="modal-actions">
          <button @click="cancelNewFolder">Cancel</button>
          <button @click="confirmNewFolder">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { MockPersistentStorage } from '../stores/mockPersistent';
import { Asset, Folder, AssetType } from '../types/mockPersistent';

const router = useRouter();
const storage = new MockPersistentStorage();

// State variables
const currentPath = ref('/');
const folders = ref<Folder[]>([]);
const assets = ref<Asset[]>([]);
const rootFolders = ref<Folder[]>([]);
const selectedAsset = ref<Asset | null>(null);
const searchTerm = ref('');

// Context menu state
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  item: null as any,
  type: '' as 'folder' | 'asset' | 'background'
});

// Rename dialog state
const renameDialog = ref({
  visible: false,
  item: null as any,
  type: '' as 'folder' | 'asset',
  newName: ''
});

// New folder dialog state
const newFolderDialog = ref({
  visible: false,
  name: ''
});

// Computed path segments for breadcrumb
const pathSegments = computed(() => {
  const segments = currentPath.value.split('/').filter(Boolean);
  return ['Content', ...segments];
});

// Initialize content browser
onMounted(async () => {
  loadCurrentFolder();
  loadRootFolders();
  
  // Close context menu when clicking outside
  document.addEventListener('click', () => {
    contextMenu.value.visible = false;
  });
});

// Load root-level folders for the sidebar
async function loadRootFolders() {
  rootFolders.value = storage.getFoldersAt('/');
}

// Load current folder contents
async function loadCurrentFolder() {
  try {
    const contents = await storage.loadFolderContents(currentPath.value);
    folders.value = contents.folders;
    assets.value = contents.assets;
  } catch (error) {
    console.error('Error loading folder contents:', error);
  }
}

// Navigate to a folder
function navigateToFolder(path: string) {
  currentPath.value = path;
  selectedAsset.value = null;
  loadCurrentFolder();
}

// Navigate up one level
function navigateUp() {
  if (currentPath.value === '/') return;
  
  const pathParts = currentPath.value.split('/').filter(Boolean);
  pathParts.pop();
  currentPath.value = '/' + pathParts.join('/');
  if (currentPath.value !== '/') currentPath.value += '/';
  
  loadCurrentFolder();
}

// Navigate using breadcrumb
function navigateToBreadcrumb(index: number) {
  if (index === 0) {
    navigateToFolder('/');
    return;
  }
  
  const segments = currentPath.value.split('/').filter(Boolean);
  const newPath = '/' + segments.slice(0, index).join('/') + '/';
  navigateToFolder(newPath);
}

// Select an asset
function selectAsset(asset: Asset) {
  selectedAsset.value = asset;
}

// Open an asset in the editor
function openAsset(asset: Asset) {
  if (asset.type === AssetType.BLUEPRINT) {
    router.push(`/editor/${asset.id}`);
  }
}

// Refresh current folder
function refresh() {
  loadCurrentFolder();
}

// Search functionality
function performSearch() {
  if (!searchTerm.value) {
    loadCurrentFolder();
    return;
  }
  
  storage.searchAssets({
    term: searchTerm.value,
    path: currentPath.value
  }).then(results => {
    assets.value = results;
    folders.value = []; // Hide folders when searching
  });
}

// Context menu handlers
function showItemContextMenu(event: MouseEvent, item: any, type: 'folder' | 'asset') {
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    item: item,
    type: type
  };
}

function showFolderContextMenu(event: MouseEvent) {
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    item: null,
    type: 'background'
  };
}

function openContextMenuItem(action: string) {
  const item = contextMenu.value.item;
  const type = contextMenu.value.type;
  
  switch (action) {
    case 'open':
      if (type === 'folder') {
        navigateToFolder(item.path);
      } else if (type === 'asset') {
        openAsset(item);
      }
      break;
      
    case 'rename':
      renameDialog.value = {
        visible: true,
        item: item,
        type: type,
        newName: item.name
      };
      break;
      
    case 'delete':
      if (type === 'folder') {
        deleteFolder(item);
      } else if (type === 'asset') {
        deleteAsset(item);
      }
      break;
      
    case 'duplicate':
      if (type === 'asset') {
        duplicateAsset(item);
      }
      break;
      
    case 'refresh':
      refresh();
      break;
      
    case 'newFolder':
      newFolderDialog.value = {
        visible: true,
        name: ''
      };
      break;
      
    case 'newBlueprint':
      createNewBlueprint();
      break;
  }
  
  contextMenu.value.visible = false;
}

// Asset operations
async function deleteAsset(asset: Asset) {
  if (confirm(`Are you sure you want to delete "${asset.name}"?`)) {
    try {
      await storage.deleteAsset(asset.id);
      loadCurrentFolder();
    } catch (error) {
      console.error('Error deleting asset:', error);
    }
  }
}

async function duplicateAsset(asset: Asset) {
  try {
    const duplicatedAsset: Asset = {
      ...asset,
      id: '',  // Will be assigned by storage
      name: `${asset.name} (Copy)`,
      created: new Date(),
      modified: new Date()
    };
    
    await storage.saveAsset(duplicatedAsset);
    loadCurrentFolder();
  } catch (error) {
    console.error('Error duplicating asset:', error);
  }
}

async function createNewBlueprint() {
  try {
    const assetName = 'NewBlueprint';
    const newAsset: Asset = {
      id: '',  // Will be assigned by storage
      name: assetName,
      path: `${currentPath.value}${assetName}`,
      type: AssetType.BLUEPRINT,
      created: new Date(),
      modified: new Date(),
      metadata: { author: 'User', version: '1.0' },
      content: { nodes: [], connections: [] }
    };
    
    const savedAsset = await storage.saveAsset(newAsset);
    loadCurrentFolder();
    
    // Select and open the new asset
    selectedAsset.value = savedAsset;
    openAsset(savedAsset);
  } catch (error) {
    console.error('Error creating blueprint:', error);
  }
}

// Folder operations
async function deleteFolder(folder: Folder) {
  if (confirm(`Are you sure you want to delete "${folder.name}" and all its contents?`)) {
    try {
      await storage.deleteFolder(folder.path);
      loadCurrentFolder();
      loadRootFolders();
    } catch (error) {
      console.error('Error deleting folder:', error);
    }
  }
}

// Dialog handlers
function confirmRename() {
  const item = renameDialog.value.item;
  const type = renameDialog.value.type;
  const newName = renameDialog.value.newName;
  
  if (type === 'asset') {
    const asset = item as Asset;
    const pathParts = asset.path.split('/');
    pathParts[pathParts.length - 1] = newName;
    
    storage.saveAsset({
      ...asset,
      name: newName,
      path: pathParts.join('/')
    }).then(() => {
      loadCurrentFolder();
    });
  } else if (type === 'folder') {
    // For folders, we would need proper folder rename functionality
    // This is a simplified version - in a real app, you would handle updating paths of all contained assets
    console.warn('Folder renaming not fully implemented');
  }
  
  renameDialog.value.visible = false;
}

function cancelRename() {
  renameDialog.value.visible = false;
}

async function confirmNewFolder() {
  try {
    await storage.createFolder(
      currentPath.value, 
      newFolderDialog.value.name
    );
    loadCurrentFolder();
    loadRootFolders();
    newFolderDialog.value.visible = false;
  } catch (error) {
    console.error('Error creating folder:', error);
  }
}

function cancelNewFolder() {
  newFolderDialog.value.visible = false;
}
</script>

<style scoped>
.content-browser-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #1e1e1e;
  color: #e0e0e0;
}

.content-browser-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background-color: #252525;
  border-bottom: 1px solid #333;
}

.path-breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.breadcrumb-segment {
  cursor: pointer;
  padding: 4px;
  white-space: nowrap;
}

.breadcrumb-segment:hover {
  color: #4a9af5;
}

.separator {
  margin: 0 4px;
  color: #555;
}

.content-browser-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.content-browser-toolbar button {
  background-color: #333;
  border: 1px solid #444;
  color: #e0e0e0;
  border-radius: 2px;
  padding: 4px 8px;
  cursor: pointer;
}

.content-browser-toolbar button:hover {
  background-color: #444;
}

.content-browser-toolbar button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.content-browser-toolbar input {
  background-color: #333;
  border: 1px solid #444;
  color: #e0e0e0;
  border-radius: 2px;
  padding: 4px 8px;
  min-width: 200px;
}

.content-browser-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.content-browser-sidebar {
  width: 250px;
  background-color: #252525;
  border-right: 1px solid #333;
  overflow-y: auto;
  padding: 8px 0;
}

.folder-tree {
  display: flex;
  flex-direction: column;
}

.folder-tree-item {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  cursor: pointer;
  margin-bottom: 2px;
}

.folder-tree-item:hover {
  background-color: #333;
}

.folder-tree-item.active {
  background-color: #2c5d96;
}

.folder-icon {
  margin-right: 8px;
  font-size: 1em;
}

.content-browser-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.folder-grid, .asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.folder-item, .asset-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  width: 100%;
  height: 100px;
  box-sizing: border-box;
  text-align: center;
}

.folder-item:hover, .asset-item:hover {
  background-color: #333;
}

.asset-item.selected {
  background-color: #2c5d96;
}

.folder-icon, .asset-icon {
  font-size: 2em;
  margin-bottom: 8px;
}

.folder-name, .asset-name {
  font-size: 0.9em;
  word-break: break-word;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
  max-height: 2.8em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* Context Menu Styles */
.context-menu {
  position: fixed;
  background-color: #252525;
  border: 1px solid #444;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
  min-width: 150px;
  z-index: 1000;
}

.context-menu-item {
  padding: 8px 16px;
  cursor: pointer;
}

.context-menu-item:hover {
  background-color: #444;
}

/* Modal Dialog Styles */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-dialog {
  background-color: #252525;
  border-radius: 4px;
  padding: 16px;
  min-width: 300px;
}

.modal-dialog h3 {
  margin-top: 0;
  margin-bottom: 16px;
}

.modal-dialog input {
  width: 100%;
  box-sizing: border-box;
  padding: 8px;
  margin-bottom: 16px;
  background-color: #333;
  border: 1px solid #444;
  color: #e0e0e0;
  border-radius: 2px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.modal-actions button {
  padding: 6px 12px;
  border-radius: 2px;
  cursor: pointer;
  background-color: #333;
  border: 1px solid #444;
  color: #e0e0e0;
}

.modal-actions button:last-child {
  background-color: #1a73e8;
  border-color: #1a73e8;
}

.modal-actions button:hover {
  background-color: #444;
}

.modal-actions button:last-child:hover {
  background-color: #2c5fba;
}
</style>
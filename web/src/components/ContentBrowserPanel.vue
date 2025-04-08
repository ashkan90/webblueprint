<template>
  <div class="content-browser-panel" :class="{ 'collapsed': isCollapsed }" :style="panelStyle">
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
        <button @click="toggleCollapse" class="toggle-btn">
          {{ isCollapsed ? '▲' : '▼' }}
        </button>
      </div>
    </div>

    <!-- Resize handle -->
    <div
      v-if="!isCollapsed"
      class="resize-handle"
      @mousedown="startResizing"
    ></div>

    <div v-if="!isCollapsed" class="content-browser-content">
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

        <div class="browser-main-content" @contextmenu.prevent="showFolderContextMenu">
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
            <!-- Existing Assets -->
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
                <!-- Add other asset type icons here -->
              </div>
              <div class="asset-name">{{ asset.name }}</div>
            </div>
          </div>

          <!-- Schema Components Grid -->
          <div class="asset-grid schema-component-grid">
             <div class="grid-header">Schema Components</div>
             <div
               v-for="schema in schemaComponents"
               :key="schema.id"
               class="asset-item schema-component-item"
               @contextmenu.prevent="showItemContextMenu($event, schema, 'schema_component')"
               @dblclick="openSchemaComponent(schema)"
               @click="selectSchemaComponent(schema)"
               :class="{ selected: selectedSchemaComponent?.id === schema.id }"
               draggable="true" @dragstart="handleDragStart($event, schema, 'schema_component')"
             >
               <div class="asset-icon schema-component">
                 <span>{;}</span> <!-- Placeholder Icon -->
               </div>
               <div class="asset-name">{{ schema.name }}</div>
             </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
      @click.stop
    >
      <!-- Folder Context Menu -->
      <div v-if="contextMenu.type === 'folder'">
        <div class="context-menu-item" @click="openContextMenuItem('open')">Open</div>
        <div class="context-menu-item" @click="openContextMenuItem('rename')">Rename</div>
        <div class="context-menu-item" @click="openContextMenuItem('delete')">Delete</div>
      </div>

      <!-- Asset Context Menu -->
      <div v-else-if="contextMenu.type === 'asset'">
        <div class="context-menu-item" @click="openContextMenuItem('open')">Open</div>
        <div class="context-menu-item" @click="openContextMenuItem('rename')">Rename</div>
        <div class="context-menu-item" @click="openContextMenuItem('duplicate')">Duplicate</div>
        <div class="context-menu-item" @click="openContextMenuItem('delete')">Delete</div>
      </div>

       <!-- Schema Component Context Menu -->
      <div v-else-if="contextMenu.type === 'schema_component'">
        <div class="context-menu-item" @click="openContextMenuItem('open')">Open</div>
        <div class="context-menu-item" @click="openContextMenuItem('rename')">Rename</div>
        <!-- <div class="context-menu-item" @click="openContextMenuItem('duplicate')">Duplicate</div> -->
        <div class="context-menu-item" @click="openContextMenuItem('delete')">Delete</div>
      </div>

      <!-- Background Context Menu -->
      <div v-else-if="contextMenu.type === 'background'">
        <div class="context-menu-item" @click="openContextMenuItem('refresh')">Refresh</div>
        <div class="context-menu-item" @click="openContextMenuItem('newFolder')">New Folder</div>
        <div class="context-menu-item" @click="openContextMenuItem('newBlueprint')">New Blueprint</div>
        <div class="context-menu-item" @click="openContextMenuItem('newSchemaComponent')">New Schema Component</div>
      </div>
    </div>

    <!-- Rename Dialog -->
    <div v-if="renameDialog.visible" class="modal-backdrop">
      <div class="modal-dialog">
        <h3>Rename {{ renameDialog.type === 'folder' ? 'Folder' : (renameDialog.type === 'schema_component' ? 'Schema Component' : 'Asset') }}</h3>
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

    <!-- New Schema Component Dialog -->
    <div v-if="newSchemaComponentDialog.visible" class="modal-backdrop">
       <div class="modal-dialog">
         <h3>New Schema Component</h3>
         <input type="text" v-model="newSchemaComponentDialog.name" placeholder="Component Name" />
         <textarea v-model="newSchemaComponentDialog.definition" placeholder="Enter Schema Definition (e.g., JSON)"></textarea>
         <div class="modal-actions">
           <button @click="cancelNewSchemaComponent">Cancel</button>
           <button @click="confirmNewSchemaComponent">Create</button>
         </div>
       </div>
     </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue';
import { MockPersistentStorage } from '../stores/mockPersistent'; // Keep for now
import { Asset, Folder, AssetType } from '../types/mockPersistent'; // Keep for now
import { schemaComponentService } from '../services/schemaComponentService'; // Use relative path
import type { SchemaComponent } from '../types/schemaComponent'; // Use relative path

const emit = defineEmits(['asset-selected', 'asset-opened']);

const storage = new MockPersistentStorage();

// State variables
const currentPath = ref('/');
const folders = ref<Folder[]>([]);
const assets = ref<Asset[]>([]);
const rootFolders = ref<Folder[]>([]);
const selectedAsset = ref<Asset | null>(null);
const schemaComponents = ref<SchemaComponent[]>([]); // State for schema components
const selectedSchemaComponent = ref<SchemaComponent | null>(null); // State for selected schema component
const searchTerm = ref('');
const isCollapsed = ref(false);
const panelHeight = ref(35); // Initial height as percentage
const isResizing = ref(false);

// Computed panel style
const panelStyle = computed(() => {
  if (isCollapsed.value) {
    return {
      height: '40px',
      width: '95%',
      margin: '0 auto'
    };
  } else {
    return {
      height: `${panelHeight.value}vh`,
      width: '95%',
      margin: '0 auto'
    };
  }
});

// Context menu state
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  item: null as any,
  type: '' as 'folder' | 'asset' | 'schema_component' | 'background' // Added schema_component
});

// Rename dialog state
const renameDialog = ref({
  visible: false,
  item: null as any,
  type: '' as 'folder' | 'asset' | 'schema_component', // Added schema_component
  newName: ''
});

// New folder dialog state
const newFolderDialog = ref({
  visible: false,
  name: ''
});

// New Schema Component dialog state
const newSchemaComponentDialog = ref({
  visible: false,
  name: '',
  definition: '{}' // Default to empty JSON object
});

// Computed path segments for breadcrumb
const pathSegments = computed(() => {
  const segments = currentPath.value.split('/').filter(Boolean);
  return ['Content', ...segments];
});

// Initialize content browser
onMounted(async () => {
  loadCurrentFolder(); // Loads mock folders/assets
  loadRootFolders();   // Loads mock root folders
  loadSchemaComponents(); // Load schema components from API

  // Close context menu when clicking outside
  document.addEventListener('click', handleDocumentClick);

  // Add resize handlers to document
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', stopResizing);
});

// Clean up event listeners on unmount
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResizing);
  document.removeEventListener('click', handleDocumentClick);
});

function handleDocumentClick() {
    if (contextMenu.value.visible) {
        contextMenu.value.visible = false;
    }
}


function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value;
}

// Resizing functionality
function startResizing(event: MouseEvent) {
  isResizing.value = true;
  event.preventDefault();
}

function handleMouseMove(event: MouseEvent) {
  if (!isResizing.value) return;

  // Calculate height based on mouse position
  const windowHeight = window.innerHeight;
  const mouseY = event.clientY;

  // Height percentage based on position from bottom of window
  const heightPercentage = ((windowHeight - mouseY) / windowHeight) * 100;

  // Limit height between 15% and 75%
  panelHeight.value = Math.min(Math.max(heightPercentage, 15), 75);
}

function stopResizing() {
  isResizing.value = false;
}

// Load root-level folders for the sidebar
async function loadRootFolders() {
  rootFolders.value = storage.getFoldersAt('/');
}

// Load current folder contents (Mock data)
async function loadCurrentFolder() {
  try {
    // This still uses the mock storage for folders/assets
    const contents = await storage.loadFolderContents(currentPath.value);
    folders.value = contents.folders;
    assets.value = contents.assets;
  } catch (error) {
    console.error('Error loading mock folder contents:', error);
  }
}

// Load schema components from API
async function loadSchemaComponents() {
  try {
    schemaComponents.value = await schemaComponentService.listSchemaComponents();
  } catch (error) {
    console.error('Error loading schema components:', error);
    // TODO: Show error to user
  }
}

// Navigate to a folder
function navigateToFolder(path: string) {
  currentPath.value = path;
  selectedAsset.value = null;
  selectedSchemaComponent.value = null; // Clear schema selection too
  loadCurrentFolder();
  // Schema components are global, no need to reload per folder
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
  selectedSchemaComponent.value = null; // Deselect schema component
  emit('asset-selected', asset);
}

// Select a schema component
function selectSchemaComponent(schema: SchemaComponent) {
  selectedSchemaComponent.value = schema;
  selectedAsset.value = null; // Deselect asset
  // Optionally emit an event if needed elsewhere
  // emit('schema-component-selected', schema);
}

// Open an asset in the editor
function openAsset(asset: Asset) {
  if (asset.type === AssetType.BLUEPRINT) {
    emit('asset-opened', asset);
  }
  // TODO: Implement opening other asset types if needed
}

// Open a schema component (placeholder)
function openSchemaComponent(schema: SchemaComponent) {
 console.log('Opening schema component editor (not implemented):', schema.name);
 // TODO: Implement opening schema component editor
 // emit('schema-component-opened', schema);
}


// Refresh current folder and schema components
async function refresh() {
  await loadCurrentFolder();
  await loadSchemaComponents();
}

// Search functionality (currently only searches mock assets)
function performSearch() {
  if (!searchTerm.value) {
    loadCurrentFolder();
    // TODO: Potentially filter schemaComponents as well or implement backend search
    return;
  }

  storage.searchAssets({
    term: searchTerm.value,
    path: currentPath.value
  }).then(results => {
    assets.value = results;
    folders.value = []; // Hide folders when searching
    // TODO: Filter schemaComponents based on searchTerm
    // schemaComponents.value = allSchemaComponents.filter(s => s.name.toLowerCase().includes(searchTerm.value.toLowerCase()));
  });
}

// Context menu handlers
function showItemContextMenu(event: MouseEvent, item: any, type: 'folder' | 'asset' | 'schema_component') {
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    item: item,
    type: type
  };
}

function showFolderContextMenu(event: MouseEvent) {
  // Prevent showing if clicking on an item within the content area
  if ((event.target as HTMLElement).closest('.folder-item') || (event.target as HTMLElement).closest('.asset-item')) {
      return;
  }
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
  const type = contextMenu.value.type as 'folder' | 'asset' | 'schema_component' | 'background';

  switch (action) {
    case 'open':
      if (type === 'folder') {
        navigateToFolder(item.path);
      } else if (type === 'asset') {
        openAsset(item);
      } else if (type === 'schema_component') {
        openSchemaComponent(item);
      }
      break;

    case 'rename':
       if (type === 'folder' || type === 'asset' || type === 'schema_component') {
         renameDialog.value = {
           visible: true,
           item: item,
           type: type,
           newName: item.name
         };
       }
      break;

    case 'delete':
      if (type === 'folder') {
        deleteFolder(item);
      } else if (type === 'asset') {
        deleteAsset(item);
      } else if (type === 'schema_component') {
        deleteSchemaComponent(item);
      }
      break;

    case 'duplicate': // Only for assets currently
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

    case 'newSchemaComponent':
      newSchemaComponentDialog.value = { visible: true, name: '', definition: '{}' };
      break;
  }

  // Visibility is handled by the document click listener now
  // contextMenu.value.visible = false;
}

// Asset operations (Mock data)
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
    await loadCurrentFolder();

    // Select and open the new asset
    selectedAsset.value = savedAsset;
    openAsset(savedAsset);
  } catch (error) {
    console.error('Error creating blueprint:', error);
  }
}

// Folder operations (Mock data)
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

// Schema Component operations
async function deleteSchemaComponent(schema: SchemaComponent) {
 if (confirm(`Are you sure you want to delete Schema Component "${schema.name}"?`)) {
   try {
     await schemaComponentService.deleteSchemaComponent(schema.id);
     await loadSchemaComponents(); // Refresh list
     if (selectedSchemaComponent.value?.id === schema.id) {
       selectedSchemaComponent.value = null; // Deselect if deleted
     }
   } catch (error) {
     console.error('Error deleting schema component:', error);
     alert(`Failed to delete schema component: ${error}`); // Replace with better feedback
   }
 }
}


// Dialog handlers

// --- Rename Dialog ---
function cancelRename() {
  renameDialog.value.visible = false;
}

async function confirmRename() {
  const item = renameDialog.value.item;
  const type = renameDialog.value.type;
  const newName = renameDialog.value.newName.trim();

  if (!newName) {
    alert('Name cannot be empty.');
    return;
  }

  if (type === 'asset') {
    const asset = item as Asset;
    const pathParts = asset.path.split('/');
    pathParts[pathParts.length - 1] = newName;
    const newPath = pathParts.join('/');

    try {
      await storage.renameAsset(asset.id, newName, newPath);
      loadCurrentFolder();
    } catch (error) {
      console.error('Error renaming asset:', error);
      alert(`Failed to rename asset: ${error}`);
    }
  } else if (type === 'folder') {
     const folder = item as Folder;
     const pathParts = folder.path.slice(0, -1).split('/'); // Remove trailing slash before split
     pathParts[pathParts.length - 1] = newName;
     const newPath = pathParts.join('/') + '/'; // Add trailing slash back

     try {
       await storage.renameFolder(folder.path, newPath);
       loadCurrentFolder();
       loadRootFolders(); // Update sidebar too
     } catch (error) {
       console.error('Error renaming folder:', error);
       alert(`Failed to rename folder: ${error}`);
     }
  } else if (type === 'schema_component') {
      const schema = item as SchemaComponent;
      try {
          // Assuming update only changes name for now, definition update needs separate UI
          await schemaComponentService.updateSchemaComponent(schema.id, newName, schema.schema_definition);
          await loadSchemaComponents();
      } catch (error) {
          console.error('Error renaming schema component:', error);
          alert(`Failed to rename schema component: ${error}`);
      }
  }

  cancelRename();
}

// --- New Folder Dialog ---
function cancelNewFolder() {
  newFolderDialog.value.visible = false;
}

async function confirmNewFolder() {
  const name = newFolderDialog.value.name.trim();
  if (!name) {
    alert('Folder name cannot be empty.');
    return;
  }

  const newPath = `${currentPath.value}${name}/`;
  try {
    await storage.createFolder(newPath);
    loadCurrentFolder();
    loadRootFolders(); // Update sidebar too
    cancelNewFolder();
  } catch (error) {
    console.error('Error creating folder:', error);
    alert(`Failed to create folder: ${error}`);
  }
}

// --- New Schema Component Dialog ---
function cancelNewSchemaComponent() {
  newSchemaComponentDialog.value.visible = false;
  // Reset fields
  newSchemaComponentDialog.value.name = '';
  newSchemaComponentDialog.value.definition = '{}';
}

async function confirmNewSchemaComponent() {
  if (!newSchemaComponentDialog.value.name.trim()) {
    alert('Schema Component name cannot be empty.'); // Replace with better validation/feedback
    return;
  }
  if (!newSchemaComponentDialog.value.definition.trim()) {
    alert('Schema definition cannot be empty.'); // Replace with better validation/feedback
    return;
  }
  // Basic JSON validation
  try {
    JSON.parse(newSchemaComponentDialog.value.definition);
  } catch (e) {
    alert('Schema definition is not valid JSON.'); // Replace with better validation/feedback
    return;
  }

  try {
    await schemaComponentService.createSchemaComponent(
      newSchemaComponentDialog.value.name,
      newSchemaComponentDialog.value.definition
    );
    await loadSchemaComponents(); // Refresh the list
    cancelNewSchemaComponent(); // Close dialog
  } catch (error) {
    console.error('Error creating schema component:', error);
    alert(`Failed to create schema component: ${error}`); // Replace with better feedback
  }
}

// Drag and Drop
function handleDragStart(event: DragEvent, item: SchemaComponent | Asset, type: 'schema_component' | 'asset') {
  if (event.dataTransfer) {
    const data = {
      type: type,
      id: item.id,
      name: item.name,
      // Add other relevant data if needed for dropping
      ...(type === 'schema_component' && { schemaDefinition: (item as SchemaComponent).schema_definition }),
      ...(type === 'asset' && { assetType: (item as Asset).type })
    };
    event.dataTransfer.setData('application/json', JSON.stringify(data));
    event.dataTransfer.effectAllowed = 'copy'; // Or 'move' if applicable
    console.log(`Dragging ${type}: ${item.name}`);
  }
}

</script>

<style scoped>
.content-browser-panel {
  display: flex;
  flex-direction: column;
  background-color: #f0f0f0;
  border-top: 1px solid #ccc;
  overflow: hidden;
  position: relative; /* Needed for absolute positioning of resize handle */
  height: 35vh; /* Default height */
  width: 95%;
  margin: 0 auto;
  transition: height 0.3s ease;
}

.content-browser-panel.collapsed {
  height: 40px; /* Height when collapsed */
}

.content-browser-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 10px;
  background-color: #e0e0e0;
  border-bottom: 1px solid #ccc;
  flex-shrink: 0; /* Prevent header from shrinking */
  height: 40px; /* Fixed header height */
  box-sizing: border-box;
}

.path-breadcrumb {
  font-size: 0.9em;
  color: #333;
}

.breadcrumb-segment {
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
}
.breadcrumb-segment:hover {
  background-color: #d0d0d0;
}
.breadcrumb-segment .separator {
  margin: 0 3px;
  color: #888;
}

.content-browser-toolbar {
  display: flex;
  align-items: center;
  gap: 5px;
}

.content-browser-toolbar button {
  background: none;
  border: 1px solid transparent;
  padding: 3px 6px;
  cursor: pointer;
  border-radius: 3px;
}
.content-browser-toolbar button:hover:not(:disabled) {
  background-color: #d0d0d0;
  border-color: #bbb;
}
.content-browser-toolbar button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.content-browser-toolbar input[type="text"] {
  padding: 4px 6px;
  border: 1px solid #ccc;
  border-radius: 3px;
  font-size: 0.9em;
}
.toggle-btn {
  font-size: 1.2em;
  padding: 0 5px;
}

.resize-handle {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 5px; /* Height of the draggable area */
  cursor: ns-resize;
  background-color: transparent; /* Make it invisible */
  z-index: 10; /* Ensure it's above content */
}
/* Optional: Add visual indicator on hover */
.resize-handle:hover {
  background-color: rgba(0, 0, 0, 0.1);
}


.content-browser-content {
  flex-grow: 1; /* Allow content to fill remaining space */
  overflow: hidden; /* Prevent content overflow */
  display: flex; /* Needed for layout */
}

.content-browser-layout {
  display: flex;
  width: 100%;
  height: 100%;
}

.content-browser-sidebar {
  width: 200px;
  background-color: #e8e8e8;
  border-right: 1px solid #ccc;
  padding: 10px;
  overflow-y: auto;
  flex-shrink: 0;
}

.folder-tree-item {
  padding: 4px 8px;
  cursor: pointer;
  border-radius: 3px;
  margin-bottom: 2px;
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.9em;
}
.folder-tree-item:hover {
  background-color: #d8d8d8;
}
.folder-tree-item.active {
  background-color: #c8c8c8;
  font-weight: bold;
}
.folder-tree-item.root {
  font-weight: bold;
  margin-bottom: 8px;
}
.folder-icon {
  font-size: 1.1em;
}

.browser-main-content {
  flex-grow: 1;
  padding: 15px;
  overflow-y: auto;
  background-color: #f8f8f8;
}

.folder-grid, .asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 15px;
  margin-bottom: 20px; /* Space between folders and assets */
}

.folder-item, .asset-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 10px;
  border: 1px solid transparent;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease;
  min-height: 80px; /* Ensure items have some height */
}
.folder-item:hover, .asset-item:hover {
  background-color: #eee;
  border-color: #ddd;
}
.asset-item.selected {
  background-color: #d5e5f5;
  border-color: #a5c5e5;
}

.folder-icon, .asset-icon {
  font-size: 2.5em; /* Larger icons */
  margin-bottom: 5px;
}
.folder-name, .asset-name {
  font-size: 0.8em;
  word-break: break-word; /* Prevent long names from overflowing */
  max-height: 2.4em; /* Limit name height */
  overflow: hidden;
}

/* Context Menu Styling */
.context-menu {
  position: fixed;
  background-color: white;
  border: 1px solid #ccc;
  box-shadow: 2px 2px 5px rgba(0,0,0,0.15);
  padding: 5px 0;
  z-index: 1001; /* Ensure it's above other elements */
  min-width: 150px;
}

.context-menu-item {
  padding: 8px 15px;
  cursor: pointer;
  font-size: 0.9em;
}

.context-menu-item:hover {
  background-color: #eee;
}

/* Modal styles */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-dialog {
  background-color: white;
  padding: 20px;
  border-radius: 5px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  min-width: 300px;
  max-width: 500px; /* Limit width */
}

.modal-dialog h3 {
  margin-top: 0;
  margin-bottom: 15px;
}

.modal-dialog input[type="text"],
.modal-dialog textarea {
  width: 100%;
  padding: 8px;
  margin-bottom: 10px;
  box-sizing: border-box;
  border: 1px solid #ccc;
  border-radius: 3px;
}

.modal-dialog textarea {
  min-height: 150px;
  font-family: monospace;
  resize: vertical; /* Allow vertical resize */
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 15px;
}

/* Schema Component Specific Styles */
.schema-component-grid {
  margin-top: 20px;
  padding-top: 10px;
  border-top: 1px solid #ccc; /* Separator */
}

.grid-header {
  font-weight: bold;
  color: #888;
  margin-bottom: 10px;
  font-size: 0.9em;
  text-transform: uppercase;
  grid-column: 1 / -1; /* Span across all columns */
}

.asset-item.schema-component-item {
  /* Add specific styles if needed */
}

.asset-icon.schema-component span {
  font-size: 1.8em; /* Adjust icon size */
  color: #6a1b9a; /* Example color */
}

/* Add styles for selected schema component */
.asset-item.schema-component-item.selected {
 border: 1px solid #6a1b9a;
 background-color: #f3e5f5;
}

</style>
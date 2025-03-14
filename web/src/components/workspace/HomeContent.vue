<template>
  <div class="home-content">
    <div v-if="currentWorkspace" class="workspace-view">
      <div class="workspace-header">
        <div class="workspace-actions">
<!--          <button @click="showWorkspaceSelector = true" class="btn">-->
<!--            Switch Workspace-->
<!--          </button>-->
          <router-link to="/editor" class="btn primary">Create New Blueprint</router-link>
        </div>
      </div>
      
      <div v-if="isLoading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading blueprints...</p>
      </div>
      
      <div v-else-if="error" class="error-state">
        <p class="error">{{ error }}</p>
      </div>
      
      <div v-else class="blueprint-grid">
        <div 
          v-for="blueprint in blueprints" 
          :key="blueprint.id"
          class="blueprint-item"
          @click="openBlueprint(blueprint.id)"
        >
          <div class="blueprint-icon" :title="blueprint.name">
            <span class="blueprint-initials">{{ getBlueprintInitials(blueprint.name) }}</span>
          </div>
          <div class="blueprint-name">{{ blueprint.name }}</div>
        </div>
        
        <div class="blueprint-item add-blueprint" @click="createNewBlueprint">
          <div class="blueprint-icon add">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
            </svg>
          </div>
          <div class="blueprint-name">New Blueprint</div>
        </div>
      </div>
    </div>
    
    <div v-else-if="showWorkspaceSelector || workspaces.length === 0" class="workspace-selector">
      <h1>Select a Workspace</h1>
      
      <div v-if="isLoadingWorkspaces" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading workspaces...</p>
      </div>
      
      <div v-else-if="workspaceError" class="error-state">
        <p class="error">{{ workspaceError }}</p>
      </div>
      
      <div v-else-if="workspaces.length === 0" class="empty-state">
        <p>You don't have any workspaces yet. Create your first workspace to get started.</p>
        <button @click="showCreateModal = true" class="btn primary">Create Workspace</button>
      </div>
      
      <div v-else class="workspace-grid">
        <div 
          v-for="workspace in workspaces" 
          :key="workspace.id"
          class="workspace-item"
          @click="selectWorkspace(workspace)"
        >
          <div class="workspace-icon" :title="workspace.name">
            {{ getWorkspaceInitials(workspace.name) }}
          </div>
          <div class="workspace-name">{{ workspace.name }}</div>
        </div>
        
        <div class="workspace-item add-workspace" @click="showCreateModal = true">
          <div class="workspace-icon add">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
            </svg>
          </div>
          <div class="workspace-name">New Workspace</div>
        </div>
      </div>
    </div>
    
    <div v-else class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading workspace...</p>
    </div>
    
    <!-- Create Workspace Modal -->
    <div v-if="showCreateModal" class="modal-backdrop">
      <div class="modal">
        <div class="modal-header">
          <h3>Create New Workspace</h3>
          <button class="close-btn" @click="showCreateModal = false">×</button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label for="workspace-name">Workspace Name</label>
            <input 
              id="workspace-name" 
              v-model="newWorkspace.name" 
              type="text" 
              class="form-input"
              placeholder="Enter workspace name"
            >
          </div>
          
          <div class="form-group">
            <label for="workspace-description">Description</label>
            <textarea 
              id="workspace-description" 
              v-model="newWorkspace.description" 
              class="form-input"
              placeholder="Enter workspace description"
            ></textarea>
          </div>
          
          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="newWorkspace.isPublic">
              <span>Public Workspace</span>
            </label>
          </div>
        </div>
        
        <div class="modal-footer">
          <button class="btn" @click="showCreateModal = false">Cancel</button>
          <button 
            class="btn primary" 
            @click="createNewWorkspace"
            :disabled="isCreating || !newWorkspace.name"
          >
            <span v-if="isCreating" class="button-spinner"></span>
            <span v-else>Create</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useWorkspaceStore } from '../../stores/workspace';
import { Workspace } from '../../types/workspace';

const router = useRouter();
const workspaceStore = useWorkspaceStore();

// States
const showWorkspaceSelector = ref(false);
const showCreateModal = ref(false);
const isCreating = ref(false);
const newWorkspace = ref({
  name: '',
  description: '',
  isPublic: false
});
const isLoading = ref(false);
const isLoadingWorkspaces = ref(false);
const error = ref<string | null>(null);
const workspaceError = ref<string | null>(null);
const blueprints = ref<any[]>([]);

// Computed properties
const workspaces = computed(() => workspaceStore.workspaces);
const currentWorkspace = computed(() => workspaceStore.currentWorkspace);

// Methods
function getWorkspaceInitials(name: string): string {
  if (!name) return '?';
  
  const parts = name.split(' ');
  if (parts.length === 1) {
    return parts[0].substring(0, 2).toUpperCase();
  }
  return (parts[0].charAt(0) + parts[1].charAt(0)).toUpperCase();
}

function getBlueprintInitials(name: string): string {
  if (!name) return '?';
  
  const parts = name.split(' ');
  if (parts.length === 1) {
    if (parts[0].length === 1) return parts[0].toUpperCase();
    return parts[0].substring(0, 2).toUpperCase();
  }
  return (parts[0].charAt(0) + parts[1].charAt(0)).toUpperCase();
}

function selectWorkspace(workspace: Workspace) {
  workspaceStore.setCurrentWorkspace(workspace);
  showWorkspaceSelector.value = false;
  loadBlueprintsForWorkspace(workspace.id);
}

async function createNewWorkspace() {
  if (!newWorkspace.value.name) return;
  
  isCreating.value = true;
  
  try {
    const workspace = await workspaceStore.createWorkspace(
      newWorkspace.value.name,
      newWorkspace.value.description,
      newWorkspace.value.isPublic
    );
    
    // Reset form and close modal
    newWorkspace.value = { name: '', description: '', isPublic: false };
    showCreateModal.value = false;
    
    // Select the new workspace
    workspaceStore.setCurrentWorkspace(workspace);
    showWorkspaceSelector.value = false;
    
    // Load blueprints for the new workspace
    loadBlueprintsForWorkspace(workspace.id);
    
  } catch (error) {
    console.error('Failed to create workspace:', error);
  } finally {
    isCreating.value = false;
  }
}

async function loadBlueprintsForWorkspace(workspaceId: string) {
  isLoading.value = true;
  error.value = null;
  blueprints.value = [];
  
  try {
    // Fetch blueprints for the current workspace
    const loadedBlueprints = await workspaceStore.fetchWorkspaceBlueprints(workspaceId);
    blueprints.value = loadedBlueprints;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
    console.error('Error loading blueprints:', err);
  } finally {
    isLoading.value = false;
  }
}

function openBlueprint(id: string) {
  router.push(`/editor/${id}`);
}

function createNewBlueprint() {
  router.push('/editor');
}

// Load workspaces and initialize
onMounted(async () => {
  isLoadingWorkspaces.value = true;
  workspaceError.value = null;
  
  try {
    await workspaceStore.fetchWorkspaces();
    
    // If workspaces exist and no current workspace is selected, select the first one
    if (workspaces.value.length > 0 && !currentWorkspace.value) {
      workspaceStore.setCurrentWorkspace(workspaces.value[0]);
    }
    
    // If we have a current workspace, load its blueprints
    if (currentWorkspace.value) {
      loadBlueprintsForWorkspace(currentWorkspace.value.id);
    }
  } catch (error) {
    workspaceError.value = error instanceof Error ? error.message : String(error);
    console.error('Failed to load workspaces:', error);
  } finally {
    isLoadingWorkspaces.value = false;
  }
});

// Watch for changes to current workspace
watch(() => currentWorkspace.value, (newWorkspace) => {
  if (newWorkspace) {
    loadBlueprintsForWorkspace(newWorkspace.id);
  }
});
</script>

<style scoped>
.home-content {
  height: 100%;
  width: 100%;
}

.workspace-view, .workspace-selector {
  padding: 30px;
  max-width: 1200px;
  margin: 0 auto;
}

.workspace-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.workspace-header h1 {
  font-size: 2.5rem;
  margin: 0;
  color: var(--accent-blue);
}

.workspace-actions {
  display: flex;
  gap: 12px;
}

.workspace-selector h1 {
  font-size: 2.5rem;
  margin-bottom: 30px;
  text-align: center;
  color: var(--accent-blue);
}

.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  text-align: center;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--accent-blue);
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error {
  color: var(--accent-red);
}

.workspace-grid, .blueprint-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.workspace-item, .blueprint-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 15px;
  border-radius: 8px;
  background-color: #2d2d2d;
  transition: transform 0.2s, background-color 0.2s;
  cursor: pointer;
  text-align: center;
}

.workspace-item:hover, .blueprint-item:hover {
  background-color: #3d3d3d;
  transform: translateY(-5px);
}

.workspace-icon, .blueprint-icon {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 24px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.workspace-icon {
  background-color: var(--accent-blue);
  border-radius: 50%; /* Make workspace icons circular */
}

.blueprint-icon {
  background-color: #555;
  border-radius: 12px;
}

.blueprint-initials {
  font-size: 20px;
}

.workspace-name, .blueprint-name {
  font-weight: 500;
  word-break: break-word;
  max-width: 100%;
}

.add-workspace .workspace-icon, .add-blueprint .blueprint-icon {
  background-color: #444;
}

.add-workspace .workspace-icon svg, .add-blueprint .blueprint-icon svg {
  width: 28px;
  height: 28px;
  fill: currentColor;
}

.btn {
  background-color: #444;
  border: none;
  color: white;
  padding: 10px 16px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  transition: background-color 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.btn:hover {
  background-color: #555;
}

.btn.primary {
  background-color: var(--accent-blue);
}

.btn.primary:hover {
  background-color: #0086e8;
}

/* Modal styles */
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
  z-index: 1000;
}

.modal {
  background-color: #2d2d2d;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #3d3d3d;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
}

.close-btn {
  background: none;
  border: none;
  color: #aaa;
  font-size: 1.5rem;
  cursor: pointer;
}

.close-btn:hover {
  color: white;
}

.modal-body {
  padding: 16px;
  max-height: 70vh;
  overflow-y: auto;
}

.modal-footer {
  padding: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  border-top: 1px solid #3d3d3d;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #ddd;
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  background-color: #333;
  border: 1px solid #444;
  border-radius: 4px;
  color: white;
  font-size: 1rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent-blue);
}

textarea.form-input {
  min-height: 100px;
  resize: vertical;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.button-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
}
</style>
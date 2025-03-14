<template>
  <div class="workspace-panel">
    <!-- Editor Button -->
    <div class="panel-section top">
      <router-link to="/editor" class="open-editor-btn" title="Open Editor">
        <div class="btn-icon">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
          </svg>
        </div>
        <span class="btn-text">Open Editor</span>
      </router-link>
    </div>

    <!-- Workspaces List -->
    <div class="panel-section workspaces-list">
      <div class="section-title">Workspaces</div>
      
      <div v-if="isLoading" class="loading-indicator">
        <div class="loading-spinner"></div>
      </div>
      
      <div v-else-if="error" class="error-message">
        Failed to load workspaces
      </div>
      
      <div v-else-if="workspaces.length === 0" class="empty-state">
        No workspaces found
      </div>
      
      <template v-else>
        <div 
          v-for="workspace in workspaces" 
          :key="workspace.id"
          class="workspace-item"
          :class="{ active: currentWorkspace?.id === workspace.id }"
          @click="selectWorkspace(workspace)"
        >
          <div class="workspace-icon" :title="workspace.name">
            {{ getWorkspaceInitials(workspace.name) }}
          </div>
        </div>
      </template>
      
      <button class="add-workspace-btn" @click="showCreateModal = true" title="Create New Workspace">
        <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
        </svg>
      </button>
    </div>

    <!-- User Settings -->
    <div class="panel-section bottom">
      <router-link to="/user/settings" class="user-settings">
        <div class="user-avatar">
          {{ userInitials }}
        </div>
      </router-link>
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
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useWorkspaceStore } from '../../stores/workspace';
import { useUserStore } from '../../stores/user';
import { Workspace } from '../../types/workspace';

const router = useRouter();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();

// States
const showCreateModal = ref(false);
const isCreating = ref(false);
const newWorkspace = ref({
  name: '',
  description: '',
  isPublic: false
});

// Computed properties
const workspaces = computed(() => workspaceStore.workspaces);
const currentWorkspace = computed(() => workspaceStore.currentWorkspace);
const isLoading = computed(() => workspaceStore.isLoading);
const error = computed(() => workspaceStore.error);

const userInitials = computed(() => {
  if (!userStore.currentUser) return '?';
  
  const name = userStore.currentUser.name;
  if (!name) return '?';
  
  const parts = name.split(' ');
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
});

// Methods
function getWorkspaceInitials(name: string): string {
  if (!name) return '?';
  
  const parts = name.split(' ');
  if (parts.length === 1) {
    return parts[0].substring(0, 2).toUpperCase();
  }
  return (parts[0].charAt(0) + parts[1].charAt(0)).toUpperCase();
}

function selectWorkspace(workspace: Workspace) {
  workspaceStore.setCurrentWorkspace(workspace);
  router.push('/');
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
    
  } catch (error) {
    console.error('Failed to create workspace:', error);
  } finally {
    isCreating.value = false;
  }
}

// Lifecycle
onMounted(async () => {
  try {
    await workspaceStore.fetchWorkspaces();
    if (workspaces.value.length > 0 && !currentWorkspace.value) {
      workspaceStore.setCurrentWorkspace(workspaces.value[0]);
    }
    
    // Fetch user data
    await userStore.fetchCurrentUser();
  } catch (error) {
    console.error('Failed to initialize workspace panel:', error);
  }
});
</script>

<style scoped>
.workspace-panel {
  width: 72px;
  min-width: 72px;
  background-color: #1e1e1e;
  border-right: 1px solid #333;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  transition: width 0.3s ease;
}

.workspace-panel:hover {
  width: 240px;
}

.panel-section {
  padding: 12px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.panel-section.top {
  border-bottom: 1px solid #333;
}

.panel-section.workspaces-list {
  flex: 1;
  overflow-y: auto;
  padding-top: 8px;
}

.panel-section.bottom {
  border-top: 1px solid #333;
}

.section-title {
  color: #888;
  font-size: 0.8rem;
  margin-bottom: 8px;
  padding: 0 12px;
  width: 100%;
  text-align: left;
  opacity: 0;
  transition: opacity 0.3s ease;
  white-space: nowrap;
}

.workspace-panel:hover .section-title {
  opacity: 1;
}

.open-editor-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background-color: var(--accent-blue);
  color: white;
  text-decoration: none;
  margin: 4px 8px;
  transition: background-color 0.2s;
  width: calc(100% - 16px);
  overflow: hidden;
}

.open-editor-btn:hover {
  background-color: #0086e8;
}

.btn-icon {
  width: 24px;
  height: 24px;
  min-width: 24px;
  fill: currentColor;
}

.btn-icon svg {
  width: 100%;
  height: 100%;
}

.btn-text {
  white-space: nowrap;
  overflow: hidden;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.workspace-panel:hover .btn-text {
  opacity: 1;
}

.workspace-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  margin: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  overflow: hidden;
}

.workspace-item:hover {
  background-color: #333;
}

.workspace-item.active {
  background-color: rgba(0, 120, 212, 0.2);
}

.workspace-icon {
  width: 40px;
  height: 40px;
  min-width: 40px;
  background-color: var(--accent-blue);
  color: white;
  border-radius: 50%; /* Changed from 8px to 50% to make circular */
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.add-workspace-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%; /* Changed from 8px to 50% to match workspace icons */
  background-color: #444;
  color: #ddd;
  border: none;
  margin: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.add-workspace-btn:hover {
  background-color: #555;
}

.add-workspace-btn svg {
  width: 24px;
  height: 24px;
  fill: currentColor;
}

.user-settings {
  display: flex;
  gap: 12px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  margin: 4px 8px;
  transition: background-color 0.2s;
  text-decoration: none;
}

.user-settings:hover {
  background-color: #333;
}

.user-avatar {
  width: 40px;
  height: 40px;
  min-width: 40px;
  background-color: #444;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 16px;
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
  max-height: 400px;
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
  padding: 8px 12px;
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
  min-height: 80px;
  resize: vertical;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.btn {
  background-color: #444;
  border: none;
  color: white;
  padding: 8px 16px;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn:hover {
  background-color: #555;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn.primary {
  background-color: var(--accent-blue);
}

.btn.primary:hover:not(:disabled) {
  background-color: #0086e8;
}

.loading-indicator {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

.loading-spinner, .button-spinner {
  display: inline-block;
  width: 24px;
  height: 24px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--accent-blue);
  animation: spin 1s ease-in-out infinite;
}

.button-spinner {
  width: 16px;
  height: 16px;
  border-width: 2px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-message {
  color: var(--accent-red);
  text-align: center;
  padding: 12px;
}

.empty-state {
  color: #888;
  text-align: center;
  padding: 20px 0;
}
</style>
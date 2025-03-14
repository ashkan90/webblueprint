<template>
  <div class="versions-panel">
    <div class="panel-header">
      <h3>Blueprint Versions</h3>
      <div class="header-actions">
        <button @click="refreshVersions" class="icon-button" title="Refresh versions">
          <span class="icon">🔄</span>
        </button>
      </div>
    </div>
    
    <div v-if="blueprintStore.isLoading" class="loading-indicator">
      Loading versions...
    </div>
    
    <div v-else-if="blueprintStore.availableVersions.length === 0" class="empty-state">
      <p>No versions available</p>
    </div>
    
    <div v-else class="versions-list">
      <div 
        v-for="version in blueprintStore.availableVersions" 
        :key="version.versionNumber"
        class="version-item"
        :class="{ 'current': isCurrentVersion(version.versionNumber) }"
        @click="loadVersion(version.versionNumber)"
      >
        <div class="version-number">v{{ version.versionNumber }}</div>
        <div class="version-info">
          <div class="version-date">{{ formatDate(version.createdAt) }}</div>
          <div class="version-comment">{{ version.comment || 'No comment' }}</div>
        </div>
      </div>
    </div>
    
    <div class="panel-footer">
      <button 
        @click="createNewVersion" 
        class="btn primary"
        :disabled="blueprintStore.isLoading || !blueprintStore.hasUnsavedChanges"
      >
        <span class="icon">💾</span> Save Version
      </button>
    </div>
    
    <!-- New Version Modal -->
    <div v-if="showSaveModal" class="modal-backdrop">
      <div class="modal">
        <div class="modal-header">
          <h3>Save New Version</h3>
          <button class="close-btn" @click="showSaveModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>Add an optional comment to describe this version:</p>
          <textarea 
            v-model="versionComment" 
            class="version-comment-input" 
            placeholder="What changed in this version?"
            rows="3"
          ></textarea>
          
          <div v-if="blueprintStore.availableVersions.length >= 20" class="warning-message">
            <p>⚠️ You have reached the maximum of 20 versions. The oldest version will be removed.</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showSaveModal = false">Cancel</button>
          <button 
            class="btn primary" 
            @click="saveVersionWithComment"
            :disabled="blueprintStore.isLoading"
          >
            {{ blueprintStore.isLoading ? 'Saving...' : 'Save Version' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useBlueprintStore } from '../stores/blueprint';

// State
const blueprintStore = useBlueprintStore();
const showSaveModal = ref(false);
const versionComment = ref('');
const currentVersionNumber = ref<number | null>(null);

// Methods
function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('en-US', { 
    month: 'short', 
    day: 'numeric', 
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date);
}

function isCurrentVersion(versionNumber: number): boolean {
  return versionNumber === currentVersionNumber.value;
}

async function loadVersion(versionNumber: number) {
  if (blueprintStore.isLoading) return;
  
  if (blueprintStore.hasUnsavedChanges) {
    if (!confirm('You have unsaved changes. Loading another version will discard these changes. Continue?')) {
      return;
    }
  }
  
  try {
    await blueprintStore.loadBlueprintVersion(blueprintStore.blueprint.id, versionNumber);
    currentVersionNumber.value = versionNumber;
  } catch (err) {
    console.error('Failed to load version:', err);
    alert('Failed to load version. Please try again.');
  }
}

function createNewVersion() {
  versionComment.value = '';
  showSaveModal.value = true;
}

async function saveVersionWithComment() {
  try {
    // First, save the blueprint to ensure we're creating a version of the current state
    await blueprintStore.saveBlueprint();
    
    // Then create a new version with the comment
    const versionNumber = await blueprintStore.createNewVersion(versionComment.value);
    currentVersionNumber.value = versionNumber;
    
    showSaveModal.value = false;
  } catch (err) {
    console.error('Failed to save version:', err);
    alert('Failed to save version. Please try again.');
  }
}

async function refreshVersions() {
  if (!blueprintStore.blueprint.id) return;
  
  try {
    await blueprintStore.loadBlueprintVersions(blueprintStore.blueprint.id);
  } catch (err) {
    console.error('Failed to refresh versions:', err);
    alert('Failed to refresh versions. Please try again.');
  }
}
</script>

<style scoped>
.versions-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #2d2d2d;
  color: #e0e0e0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #3d3d3d;
}

.panel-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 500;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.icon-button {
  background: none;
  border: none;
  color: #aaa;
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border-radius: 4px;
}

.icon-button:hover {
  background-color: #444;
  color: white;
}

.loading-indicator,
.empty-state {
  padding: 16px;
  text-align: center;
  color: #999;
  font-style: italic;
}

.versions-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.version-item {
  display: flex;
  align-items: flex-start;
  padding: 10px 12px;
  margin-bottom: 6px;
  border-radius: 4px;
  background-color: #333;
  cursor: pointer;
  transition: background-color 0.2s;
  border-left: 3px solid transparent;
}

.version-item:hover {
  background-color: #404040;
}

.version-item.current {
  background-color: #37474f;
  border-left-color: #3a8cd7;
}

.version-number {
  font-weight: 600;
  margin-right: 12px;
  color: #3a8cd7;
}

.version-info {
  flex: 1;
}

.version-date {
  font-size: 0.8rem;
  color: #999;
  margin-bottom: 4px;
}

.version-comment {
  font-size: 0.9rem;
  word-break: break-word;
}

.panel-footer {
  padding: 12px 16px;
  border-top: 1px solid #3d3d3d;
}

.btn {
  background-color: #444;
  border: none;
  color: white;
  padding: 8px 12px;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.2s;
  width: 100%;
  justify-content: center;
}

.btn:hover {
  background-color: #555;
}

.btn.primary {
  background-color: #3a8cd7;
}

.btn.primary:hover {
  background-color: #4a9de7;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  width: 400px;
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
}

.modal-footer {
  padding: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  border-top: 1px solid #3d3d3d;
}

.version-comment-input {
  width: 100%;
  background-color: #333;
  border: 1px solid #444;
  border-radius: 4px;
  color: white;
  padding: 8px;
  font-family: inherit;
  resize: vertical;
  margin-top: 8px;
}

.version-comment-input:focus {
  outline: none;
  border-color: #3a8cd7;
}

.warning-message {
  margin-top: 12px;
  padding: 10px;
  background-color: rgba(255, 193, 7, 0.1);
  border-left: 3px solid #ffc107;
  border-radius: 4px;
}

.warning-message p {
  margin: 0;
  color: #ffc107;
  font-size: 0.9rem;
}
</style>
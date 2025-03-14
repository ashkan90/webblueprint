<template>
  <div class="user-settings-view">
    <div class="settings-container">
      <div class="settings-header">
        <h2>User Settings</h2>
        <router-link to="/" class="btn secondary">Back to Workspace</router-link>
      </div>

      <div v-if="isLoading" class="loading-container">
        <div class="loading-spinner"></div>
        <p>Loading user data...</p>
      </div>

      <div v-else-if="error" class="error-container">
        <div class="error-icon">!</div>
        <p>{{ error }}</p>
        <button class="btn primary" @click="fetchUserData">Retry</button>
      </div>

      <div v-else-if="user" class="settings-content">
        <div class="user-profile">
          <div class="profile-header">
            <div class="user-avatar">
              {{ userInitials }}
            </div>
            <div class="user-info">
              <h3>{{ user.name }}</h3>
              <p>{{ user.email }}</p>
              <p class="user-since">User since {{ formatDate(user.createdAt) }}</p>
            </div>
          </div>
        </div>

        <div class="settings-section">
          <h4>Account Settings</h4>
          
          <div class="form-group">
            <label for="username">Username</label>
            <input 
              id="username" 
              type="text" 
              v-model="userForm.username" 
              class="form-input"
            >
          </div>
          
          <div class="form-group">
            <label for="name">Display Name</label>
            <input 
              id="name" 
              type="text" 
              v-model="userForm.name" 
              class="form-input"
            >
          </div>
          
          <div class="form-group">
            <label for="email">Email Address</label>
            <input 
              id="email" 
              type="email" 
              v-model="userForm.email" 
              class="form-input"
            >
          </div>
        </div>

        <div class="settings-section">
          <h4>Preferences</h4>
          
          <div class="form-group">
            <label for="theme">Theme</label>
            <select id="theme" v-model="userForm.preferences.theme" class="form-input">
              <option value="dark">Dark (Default)</option>
              <option value="light">Light</option>
              <option value="system">System</option>
            </select>
          </div>
          
          <div class="form-group">
            <label class="checkbox-label">
              <input 
                type="checkbox" 
                v-model="userForm.preferences.enableAutosave"
              >
              <span>Enable auto-save</span>
            </label>
          </div>
          
          <div class="form-group">
            <label class="checkbox-label">
              <input 
                type="checkbox" 
                v-model="userForm.preferences.showGridInEditor"
              >
              <span>Show grid in editor</span>
            </label>
          </div>
        </div>

        <div class="form-actions">
          <button 
            class="btn secondary" 
            @click="resetForm"
          >
            Cancel
          </button>
          <button 
            class="btn primary" 
            @click="saveUserSettings"
            :disabled="isSaving"
          >
            <span v-if="isSaving" class="button-spinner"></span>
            <span v-else>Save Changes</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useUserStore } from '../stores/user';

const userStore = useUserStore();

// State
const isLoading = ref(false);
const isSaving = ref(false);
const error = ref<string | null>(null);
const userForm = ref<any>({
  username: '',
  name: '',
  email: '',
  preferences: {
    theme: 'dark',
    enableAutosave: true,
    showGridInEditor: true
  }
});

// Computed properties
const user = computed(() => userStore.currentUser);
const userInitials = computed(() => {
  if (!user.value) return '?';
  
  const name = user.value.name;
  if (!name) return '?';
  
  const parts = name.split(' ');
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
});

// Methods
function formatDate(date: Date | string | null): string {
  if (!date) return 'Unknown';
  
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return dateObj.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
}

function resetForm() {
  if (!user.value) return;
  
  userForm.value = {
    username: user.value.username,
    name: user.value.name,
    email: user.value.email,
    preferences: { ...user.value.preferences }
  };
}

async function fetchUserData() {
  isLoading.value = true;
  error.value = null;
  
  try {
    await userStore.fetchCurrentUser();
    resetForm();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    isLoading.value = false;
  }
}

async function saveUserSettings() {
  if (!user.value) return;
  
  isSaving.value = true;
  error.value = null;
  
  try {
    // Update user preferences
    await userStore.updateUserPreferences(userForm.value.preferences);
    
    // TODO: Add API call to update user profile
    // This would require a new method in the user store
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    isSaving.value = false;
  }
}

// Lifecycle
onMounted(() => {
  fetchUserData();
});
</script>

<style scoped>
.user-settings-view {
  flex: 1;
  display: flex;
  justify-content: center;
  padding: 40px 20px;
  height: calc(100vh - 50px);
  overflow: auto;
}

.settings-container {
  width: 800px;
  max-width: 100%;
  color: var(--text-color);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.settings-header h2 {
  font-size: 2rem;
  margin: 0;
  color: var(--accent-blue);
}

.loading-container, .error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background-color: #252526;
  border-radius: 8px;
  text-align: center;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--accent-blue);
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon {
  width: 40px;
  height: 40px;
  background-color: var(--accent-red);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 16px;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 24px;
  background-color: #252526;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}

.user-avatar {
  width: 80px;
  height: 80px;
  background-color: var(--accent-blue);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: bold;
}

.user-info h3 {
  margin: 0 0 4px 0;
  font-size: 1.5rem;
}

.user-info p {
  margin: 0;
  color: #bbb;
}

.user-since {
  margin-top: 8px !important;
  font-size: 0.9rem;
}

.settings-section {
  background-color: #252526;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}

.settings-section h4 {
  margin: 0 0 20px 0;
  font-size: 1.2rem;
  color: #ddd;
  border-bottom: 1px solid #333;
  padding-bottom: 12px;
}

.form-group {
  margin-bottom: 20px;
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
  box-shadow: 0 0 0 2px rgba(0, 120, 212, 0.3);
}

select.form-input {
  appearance: none;
  background-image: url("data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%22292.4%22%20height%3D%22292.4%22%3E%3Cpath%20fill%3D%22%23FFFFFF%22%20d%3D%22M287%2069.4a17.6%2017.6%200%200%200-13-5.4H18.4c-5%200-9.3%201.8-12.9%205.4A17.6%2017.6%200%200%200%200%2082.2c0%205%201.8%209.3%205.4%2012.9l128%20127.9c3.6%203.6%207.8%205.4%2012.8%205.4s9.2-1.8%2012.8-5.4L287%2095c3.5-3.5%205.4-7.8%205.4-12.8%200-5-1.9-9.2-5.5-12.8z%22%2F%3E%3C%2Fsvg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  background-size: 12px;
  padding-right: 30px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 12px;
}

.btn {
  padding: 10px 20px;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
  font-size: 1rem;
}

.btn.primary {
  background-color: var(--accent-blue);
  color: white;
}

.btn.primary:hover:not(:disabled) {
  background-color: #0086e8;
}

.btn.secondary {
  background-color: #333;
  color: white;
}

.btn.secondary:hover {
  background-color: #444;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.button-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
}
</style>
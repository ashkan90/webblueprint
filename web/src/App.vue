<!-- File: web/src/App.vue -->
<template>
  <div class="app">
    <header class="app-header">
      <div class="logo">
        <router-link to="/" class="logo-link">
          <h1>{{ appTitle }}</h1>
        </router-link>
      </div>
      <nav>
        <RouterLink to="/content">Content Browser</RouterLink>
        <RouterLink to="/about">About</RouterLink>
      </nav>
      <div class="connection-status" :class="connectionStatus">
        <span class="status-indicator"></span>
        <span class="status-text">{{ connectionStatusText }}</span>
      </div>
    </header>

    <div v-if="isInitializing" class="initializing-overlay">
      <div class="initializing-content">
        <div class="loading-spinner"></div>
        <div class="loading-text">Initializing WebBlueprint...</div>
      </div>
    </div>

    <main class="app-content">
      <router-view v-if="!isInitializing" />
    </main>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, onBeforeUnmount, computed} from 'vue'
import { executionManager, initializeBlueprintSystem, cleanupBlueprintSystem } from './bootstrap/blueprintSystem'
import { WebSocketExecutionBridge } from './services/websocketBridge'
import {useWebSocketStore} from "./stores/websocket";
import {useWorkspaceStore} from "./stores/workspace";

const websocketStore = useWebSocketStore()
const workspaceStore = useWorkspaceStore()

// State
const isInitializing = ref(true)

// Computed properties
const appTitle = computed(() => {
  return workspaceStore.currentWorkspace ? workspaceStore.currentWorkspace.name : 'WebBlueprint';
})

// Initialize the system on component mount
onMounted(async () => {
  try {
    // Initialize blueprint system
    await initializeBlueprintSystem()

    // Initialize WebSocket bridge
    const websocketBridge = new WebSocketExecutionBridge(executionManager)
    websocketBridge.initialize()

    // Mark initialization as complete
    isInitializing.value = false
  } catch (error) {
    console.error('Failed to initialize WebBlueprint:', error)

    // Could show an error message here
    // For now, we'll still mark initialization as complete to allow UI to load
    isInitializing.value = false
  }
})

// Computed properties for connection status
const connectionStatus = computed(() => websocketStore.connectionStatus)
const connectionStatusText = computed(() => {
  switch (websocketStore.connectionStatus) {
    case 'connected':
      return 'Connected'
    case 'connecting':
      return 'Connecting...'
    case 'disconnected':
      return 'Disconnected'
    default:
      return 'Unknown'
  }
})

// Clean up on component unmount
onBeforeUnmount(() => {
  cleanupBlueprintSystem()
})
</script>

<style>
:root {
  /* Colors - Updated for Unreal Engine look */
  --bg-color: #1e1e1e;
  --text-color: #e0e0e0;
  --node-bg: #252526;
  --node-header: #333333;
  --node-selected: #0078d4;
  --grid-color: rgba(60, 60, 60, 0.2);
  --conn-color: #00a8ff;
  --conn-exec: #fff;
  --exec-pin: #fff;
  --input-pin: #ff9f43;
  --output-pin: #2ed573;
  --context-menu-bg: #252526;
  --context-menu-hover: #333333;

  /* Accent colors */
  --accent-blue: #0078d4;
  --accent-green: #2ed573;
  --accent-red: #ff4757;
  --accent-yellow: #ffa502;

  /* Spacing */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;
}

html, body {
  margin: 0;
  padding: 0;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-family: 'Segoe UI', 'Roboto', 'Helvetica Neue', sans-serif;
  height: 100%;
  overflow: hidden;
}

.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  height: 50px;
  background-color: #1a1a1a;
  border-bottom: 1px solid #333;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.logo .logo-link {
  text-decoration: none;
}

.logo h1 {
  font-size: 1.2rem;
  font-weight: bold;
  color: var(--accent-blue);
  margin: 0;
  cursor: pointer;
  transition: color 0.2s;
}

.logo h1:hover {
  color: #0090ff;
}

nav {
  display: flex;
  gap: 20px;
}

nav a {
  color: #bbb;
  text-decoration: none;
  transition: color 0.2s, border-bottom 0.2s;
  padding: 8px 0;
  border-bottom: 2px solid transparent;
}

nav a:hover {
  color: white;
  border-bottom: 2px solid rgba(0, 120, 212, 0.5);
}

nav a.router-link-active {
  color: var(--accent-blue);
  border-bottom: 2px solid var(--accent-blue);
}

.app-content {
  flex: 1;
  overflow: hidden;
}

.initializing-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.initializing-content {
  text-align: center;
}

.loading-spinner {
  display: inline-block;
  width: 50px;
  height: 50px;
  border: 5px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--accent-blue);
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 1.2rem;
  color: #fff;
}

.connection-status {
  display: flex;
  align-items: center;
  font-size: 0.8rem;
  padding: var(--space-xs) var(--space-sm);
  border-radius: 4px;
  background-color: rgba(0, 0, 0, 0.2);
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: var(--space-xs);
}

.connection-status.connected .status-indicator {
  background-color: var(--accent-green);
  box-shadow: 0 0 5px rgba(46, 213, 115, 0.8);
}

.connection-status.connecting .status-indicator {
  background-color: var(--accent-yellow);
  box-shadow: 0 0 5px rgba(255, 165, 2, 0.8);
  animation: pulse 1s infinite;
}

.connection-status.disconnected .status-indicator {
  background-color: var(--accent-red);
  box-shadow: 0 0 5px rgba(255, 71, 87, 0.8);
}

@keyframes pulse {
  0% { opacity: 0.4; }
  50% { opacity: 1; }
  100% { opacity: 0.4; }
}

/* Scrollbars - Unreal Engine style */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #1a1a1a;
}

::-webkit-scrollbar-thumb {
  background: #333;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #444;
}

/* Buttons and interactive elements */
button, 
.btn {
  font-family: 'Segoe UI', 'Roboto', 'Helvetica Neue', sans-serif;
  font-weight: 500;
  letter-spacing: 0.3px;
}

/* Form elements */
input, textarea, select {
  font-family: 'Segoe UI', 'Roboto', 'Helvetica Neue', sans-serif;
}
</style>
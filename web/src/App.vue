<!-- File: web/src/App.vue -->
<template>
  <div class="app">
    <header class="app-header">
      <div class="logo">
        <h1>WebBlueprint</h1>
      </div>
      <nav>
        <RouterLink to="/">Home</RouterLink>
        <RouterLink to="/content">Content Browser</RouterLink>
        <RouterLink to="/editor">Editor</RouterLink>
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


const websocketStore = useWebSocketStore()

// State
const isInitializing = ref(true)

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
  /* Colors */
  --bg-color: #1e1e1e;
  --text-color: #e0e0e0;
  --node-bg: #333333;
  --node-header: #444444;
  --node-selected: #1a73e8;
  --grid-color: rgba(80, 80, 80, 0.2);
  --conn-color: #8ab4f8;
  --conn-exec: #fff;
  --exec-pin: #fff;
  --input-pin: #f0883e;
  --output-pin: #6ed69a;
  --context-menu-bg: #333333;
  --context-menu-hover: #444444;

  /* Accent colors */
  --accent-blue: #3498db;
  --accent-green: #2ecc71;
  --accent-red: #e74c3c;
  --accent-yellow: #f1c40f;

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
  background-color: #242424;
  border-bottom: 1px solid #333;
}

.logo a {
  font-size: 1.2rem;
  font-weight: bold;
  color: var(--accent-blue);
  text-decoration: none;
}

nav {
  display: flex;
  gap: 20px;
}

nav a {
  color: #bbb;
  text-decoration: none;
  transition: color 0.2s;
}

nav a:hover {
  color: white;
}

nav a.router-link-active {
  color: var(--accent-blue);
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
  box-shadow: 0 0 5px rgba(46, 204, 113, 0.8);
}

.connection-status.connecting .status-indicator {
  background-color: var(--accent-yellow);
  box-shadow: 0 0 5px rgba(241, 196, 15, 0.8);
  animation: pulse 1s infinite;
}

.connection-status.disconnected .status-indicator {
  background-color: var(--accent-red);
  box-shadow: 0 0 5px rgba(231, 76, 60, 0.8);
}

@keyframes pulse {
  0% { opacity: 0.4; }
  50% { opacity: 1; }
  100% { opacity: 0.4; }
}
</style>
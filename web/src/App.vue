<template>
  <header class="app-header">
    <div class="logo">
      <h1>WebBlueprint</h1>
    </div>
    <nav>
      <RouterLink to="/">Home</RouterLink>
      <RouterLink to="/editor">Editor</RouterLink>
      <RouterLink to="/about">About</RouterLink>
    </nav>
    <div class="connection-status" :class="connectionStatus">
      <span class="status-indicator"></span>
      <span class="status-text">{{ connectionStatusText }}</span>
    </div>
  </header>

  <RouterView />
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { computed, onMounted } from 'vue'
import { useWebSocketStore } from './stores/websocket'

const websocketStore = useWebSocketStore()

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

// Connect to WebSocket when the app starts
onMounted(() => {
  websocketStore.connect()
})
</script>

<style>
/* Global styles */
:root {
  /* Colors based on Unreal Engine Blueprint theme */
  --bg-color: #1e1e1e;
  --grid-color: #2a2a2a;
  --node-bg: #2d2d2d;
  --node-header: #3a3a3a;
  --node-selected: #4a4a7a;
  --exec-pin: #ffffff;
  --conn-exec: #ffffff;
  --conn-color: #8ab4f8;
  --input-pin: #f0883e;
  --output-pin: #6ed69a;
  --text-color: #e0e0e0;
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

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background-color: var(--bg-color);
  color: var(--text-color);
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  line-height: 1.6;
  overflow: hidden;
  height: 100vh;
}

.app-header {
  background-color: #222;
  padding: 0 var(--space-md);
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  position: relative;
  z-index: 100;
}

.logo h1 {
  font-size: 1.4rem;
  color: var(--accent-blue);
  text-shadow: 0 0 5px rgba(52, 152, 219, 0.5);
}

nav {
  display: flex;
  gap: var(--space-lg);
}

nav a {
  color: #ccc;
  text-decoration: none;
  font-weight: 500;
  padding: var(--space-xs) var(--space-sm);
  border-radius: 4px;
  transition: all 0.2s ease;
}

nav a:hover {
  color: white;
  background-color: rgba(255, 255, 255, 0.1);
}

nav a.router-link-active {
  color: var(--accent-blue);
  background-color: rgba(52, 152, 219, 0.2);
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
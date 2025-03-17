<template>
  <div class="error-testing-view">
    <h1>Error Management Testing</h1>
    
    <div class="content-container">
      <div class="testing-panel">
        <ErrorTestingPanel />
      </div>
      
      <div class="error-panel">
        <ErrorPanel />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {defineComponent, onMounted, onUnmounted} from 'vue';
import ErrorTestingPanel from '../components/debug/ErrorTestingPanel.vue';
import ErrorPanel from '../components/debug/ErrorPanel.vue';
import {useErrorStore} from "../stores/errorStore";
import {useErrorViewStore} from "../stores/errorViewStore";
import {RealWebSocketConnection} from "../error-handling-setup";
import {useErrorWebSocketHandler} from "../composables/useErrorWebSocketHandler";

const errorStore = useErrorStore();
const errorViewStore = useErrorViewStore();

// Create real WebSocket connection
const wsConnection = new RealWebSocketConnection();

// Initialize WebSocket handler
const errorWsHandler = useErrorWebSocketHandler(wsConnection);

onMounted(() => {
  errorWsHandler.init();
  wsConnection.connect();

  errorStore.$subscribe((mutation, state) => {
    const criticalErrorCount = state.errors.filter(
        err => err.severity === 'critical' || err.severity === 'high'
    ).length;

    if (criticalErrorCount > 0 && errorViewStore.autoShowErrors) {
      errorViewStore.toggleErrorPanel();
    }
  });
})

onUnmounted(() => {
  errorWsHandler.cleanup()
  wsConnection.disconnect()
})

</script>

<style scoped>
.error-testing-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 20px;
  background-color: #f0f3f6;
}

h1 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #2c3e50;
  font-size: 1.8rem;
}

.content-container {
  display: flex;
  flex: 1;
  gap: 20px;
  overflow: hidden;
}

.testing-panel {
  flex: 1;
  min-width: 400px;
  max-width: 600px;
  overflow-y: auto;
}

.error-panel {
  flex: 1;
  min-width: 400px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  overflow: hidden;
  background-color: #252525;
}
</style>
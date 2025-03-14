import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

// Import error handling components
import { useErrorStore } from './stores/errorStore';
import { useErrorViewStore } from './stores/errorViewStore';
import { useErrorWebSocketHandler, WebSocketConnection } from './composables/useErrorWebSocketHandler';

// Create app instance
const app = createApp(App);
const pinia = createPinia();

// Register Pinia
app.use(pinia);
app.use(router);

// Initialize WebSocket connection
// This is a stub - replace with your actual WebSocket implementation
class AppWebSocketConnection implements WebSocketConnection {
  private handlers: Record<string, ((data: any) => void)[]> = {};
  private connected = false;
  
  addMessageHandler(type: string, callback: (data: any) => void): void {
    if (!this.handlers[type]) {
      this.handlers[type] = [];
    }
    this.handlers[type].push(callback);
  }
  
  removeMessageHandler(type: string): void {
    delete this.handlers[type];
  }
  
  connect(): void {
    this.connected = true;
    console.log('WebSocket connected');
  }
  
  disconnect(): void {
    this.connected = false;
    console.log('WebSocket disconnected');
  }
  
  isConnected(): boolean {
    return this.connected;
  }
  
  // Call this when a message is received from the server
  handleMessage(type: string, data: any): void {
    if (this.handlers[type]) {
      this.handlers[type].forEach(handler => handler(data));
    }
  }
}

// Create WebSocket connection
const wsConnection = new AppWebSocketConnection();

// Setup error handling
const setupErrorHandling = () => {
  // Initialize stores
  const errorStore = useErrorStore();
  const errorViewStore = useErrorViewStore();
  
  // Initialize WebSocket handler
  const errorWsHandler = useErrorWebSocketHandler(wsConnection);
  errorWsHandler.init();
  
  // Auto-show error panel for critical errors
  errorStore.$subscribe((mutation, state) => {
    const criticalErrorCount = state.errors.filter(
      err => err.severity === 'critical' || err.severity === 'high'
    ).length;
    
    if (criticalErrorCount > 0 && errorViewStore.autoShowErrors) {
      errorViewStore.toggleErrorPanel();
    }
  });
  
  // Connect WebSocket
  wsConnection.connect();
  
  // Return cleanup function
  return () => {
    errorWsHandler.cleanup();
    wsConnection.disconnect();
  };
};

// Initialize error handling
const cleanupErrorHandling = setupErrorHandling();

// Mount app
app.mount('#app');

// Handle cleanup on app unmount
app.unmount = () => {
  cleanupErrorHandling();
};

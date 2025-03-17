import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

// Import error handling components
import { useErrorStore } from './stores/errorStore';
import { useErrorViewStore } from './stores/errorViewStore';
import { useErrorWebSocketHandler, WebSocketConnection } from './composables/useErrorWebSocketHandler';
import { useWebSocketStore, WebSocketEvents } from './stores/websocket';

// Implement a proper WebSocket connection adapter that works with existing websocket store
export class RealWebSocketConnection implements WebSocketConnection {
  private wsStore: ReturnType<typeof useWebSocketStore>;
  private handlers: Record<string, ((data: any) => void)[]> = {};
  private unsubscribeCallbacks: (() => void)[] = [];

  constructor() {
    this.wsStore = useWebSocketStore();
  }
  
  addMessageHandler(type: string, callback: (data: any) => void): void {
    if (!this.handlers[type]) {
      this.handlers[type] = [];
    }
    this.handlers[type].push(callback);

    // Map our handler types to WebSocketEvents types
    let wsEventType: string;
    
    // Map error-specific message types to WebSocket event types
    switch (type) {
      case 'error':
        wsEventType = WebSocketEvents.NODE_ERROR;
        break;
      case 'error_analysis':
        wsEventType = WebSocketEvents.DEBUG_DATA;
        break;
      case 'recovery_attempt':
        wsEventType = WebSocketEvents.EXEC_STATUS;
        break;
      default:
        wsEventType = type;
    }

    // Register with the WebSocket store and unwrap the nested payload structure
    const unsubscribe = this.wsStore.on(wsEventType, (data: any) => {
      console.log(`Received ${wsEventType} message:`, data);
      
      // Handle the nested structure: data is the payload which contains a 'type' field
      if (data && typeof data === 'object') {
        // For error messages, the structure is:
        // { type: "error", error: {...}, executionId: "..." }
        if (data.type === type) {
          this.handlers[type].forEach(handler => handler(data));
        }
      }
    });
    
    this.unsubscribeCallbacks.push(unsubscribe);
  }
  
  removeMessageHandler(type: string): void {
    delete this.handlers[type];
    // Note: we're not removing the underlying store handlers here 
    // as they'll be cleaned up when we disconnect
  }
  
  connect(): void {
    this.wsStore.connect();
    console.log('WebSocket connection requested');
  }
  
  disconnect(): void {
    // Clean up all registered handlers
    this.unsubscribeCallbacks.forEach(unsubscribe => unsubscribe());
    this.unsubscribeCallbacks = [];
    
    // Not actually disconnecting the WebSocket since it might be used by other parts of the app
    console.log('Error handling WebSocket handlers removed');
  }
  
  isConnected(): boolean {
    return this.wsStore.connectionStatus === 'connected';
  }
}

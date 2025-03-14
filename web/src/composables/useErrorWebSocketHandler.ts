import { useErrorStore } from './errorStore';
import { ErrorNotificationType } from '../types/errors';

// Interface for WebSocket connection and message handling
export interface WebSocketConnection {
  addMessageHandler(type: string, callback: (data: any) => void): void;
  removeMessageHandler(type: string): void;
  connect(): void;
  disconnect(): void;
  isConnected(): boolean;
}

export class ErrorWebSocketHandler {
  private errorStore: ReturnType<typeof useErrorStore>;
  private wsConnection: WebSocketConnection;
  
  constructor(errorStore: ReturnType<typeof useErrorStore>, wsConnection: WebSocketConnection) {
    this.errorStore = errorStore;
    this.wsConnection = wsConnection;
  }
  
  // Initialize and register handlers
  public init(): void {
    this.registerHandlers();
  }
  
  // Register WebSocket message handlers
  private registerHandlers(): void {
    // Handle error notifications
    this.wsConnection.addMessageHandler('error', (data: any) => {
      this.handleErrorNotification(data);
    });
    
    // Handle error analysis notifications
    this.wsConnection.addMessageHandler('error_analysis', (data: any) => {
      this.handleErrorAnalysisNotification(data);
    });
    
    // Handle recovery attempt notifications
    this.wsConnection.addMessageHandler('recovery_attempt', (data: any) => {
      this.handleRecoveryNotification(data);
    });
  }
  
  // Cleanup and unregister handlers
  public cleanup(): void {
    this.wsConnection.removeMessageHandler('error');
    this.wsConnection.removeMessageHandler('error_analysis');
    this.wsConnection.removeMessageHandler('recovery_attempt');
  }
  
  // Handle error notification
  private handleErrorNotification(notification: ErrorNotificationType): void {
    if (notification.type === 'error' && notification.error) {
      this.errorStore.addError(notification.error);
    }
  }
  
  // Handle error analysis notification
  private handleErrorAnalysisNotification(notification: ErrorNotificationType): void {
    if (notification.type === 'error_analysis' && notification.analysis) {
      this.errorStore.updateErrorAnalysis(notification.analysis);
    }
  }
  
  // Handle recovery attempt notification
  private handleRecoveryNotification(notification: ErrorNotificationType): void {
    if (notification.type === 'recovery_attempt') {
      this.errorStore.addRecoveryAttempt({
        strategy: notification.strategy,
        successful: notification.successful,
        nodeId: notification.nodeId,
        errorCode: notification.errorCode,
        details: notification.details,
        executionId: notification.executionId,
        timestamp: new Date().toISOString()
      });
    }
  }
}

// Create and export a composable function to use the handler
export function useErrorWebSocketHandler(wsConnection: WebSocketConnection) {
  const errorStore = useErrorStore();
  const handler = new ErrorWebSocketHandler(errorStore, wsConnection);
  
  return {
    init: () => handler.init(),
    cleanup: () => handler.cleanup()
  };
}

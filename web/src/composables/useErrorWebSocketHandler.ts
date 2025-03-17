import { useErrorStore } from '../stores/errorStore';
import { ErrorNotificationType, RecoveryStrategy } from '../types/errors';

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
    console.log('[ErrorWebSocketHandler] Initializing...');
    this.registerHandlers();
  }
  
  // Register WebSocket message handlers
  private registerHandlers(): void {
    // Handle error notifications
    this.wsConnection.addMessageHandler('error', (data: any) => {
      console.log('[ErrorWebSocketHandler] Received error notification:', data);
      this.handleErrorNotification(data);
    });
    
    // Handle error analysis notifications
    this.wsConnection.addMessageHandler('error_analysis', (data: any) => {
      console.log('[ErrorWebSocketHandler] Received error analysis:', data);
      this.handleErrorAnalysisNotification(data);
    });
    
    // Handle recovery attempt notifications
    this.wsConnection.addMessageHandler('recovery_attempt', (data: any) => {
      console.log('[ErrorWebSocketHandler] Received recovery attempt:', data);
      this.handleRecoveryNotification(data);
    });
  }
  
  // Cleanup and unregister handlers
  public cleanup(): void {
    console.log('[ErrorWebSocketHandler] Cleaning up...');
    this.wsConnection.removeMessageHandler('error');
    this.wsConnection.removeMessageHandler('error_analysis');
    this.wsConnection.removeMessageHandler('recovery_attempt');
  }
  
  // Handle error notification
  private handleErrorNotification(notification: ErrorNotificationType): void {
    console.log('[ErrorWebSocketHandler] Processing error notification:', notification);
    
    // Validation to make sure we have a properly structured notification
    if (!notification) {
      console.error('[ErrorWebSocketHandler] Null notification received');
      return;
    }
    
    // Check for the correct type and that it contains an error object
    if (notification.type === 'error' && notification.error) {
      console.log('[ErrorWebSocketHandler] Adding error to store:', notification.error);
      this.errorStore.addError(notification.error);
    } else {
      console.warn('[ErrorWebSocketHandler] Invalid error notification format:', notification);
    }
  }
  
  // Handle error analysis notification
  private handleErrorAnalysisNotification(notification: ErrorNotificationType): void {
    console.log('[ErrorWebSocketHandler] Processing error analysis notification:', notification);
    
    // Check if we have a proper analysis object
    if (notification.type === 'error_analysis' && notification.analysis) {
      console.log('[ErrorWebSocketHandler] Updating error analysis in store');
      this.errorStore.updateErrorAnalysis(notification.analysis);
    } else {
      console.warn('[ErrorWebSocketHandler] Invalid error analysis notification format:', notification);
    }
  }
  
  // Handle recovery attempt notification
  private handleRecoveryNotification(notification: ErrorNotificationType): void {
    console.log('[ErrorWebSocketHandler] Processing recovery notification:', notification);
    
    // Check if we have a recovery attempt notification with all required fields
    if (notification.type === 'recovery_attempt' && 
        notification.nodeId && 
        notification.errorCode && 
        notification.executionId) {
      
      const recoveryAttempt = {
        strategy: notification.strategy as RecoveryStrategy,
        successful: notification.successful,
        nodeId: notification.nodeId,
        errorCode: notification.errorCode,
        details: notification.details,
        executionId: notification.executionId,
        timestamp: new Date().toISOString()
      };
      
      console.log('[ErrorWebSocketHandler] Adding recovery attempt to store:', recoveryAttempt);
      this.errorStore.addRecoveryAttempt(recoveryAttempt);
    } else {
      console.warn('[ErrorWebSocketHandler] Invalid recovery notification format:', notification);
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

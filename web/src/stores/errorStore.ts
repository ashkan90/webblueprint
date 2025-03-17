import {defineStore} from 'pinia';
import {BlueprintError, ErrorAnalysis, ErrorSeverity, RecoveryAttempt, RecoveryStrategy} from '../types/errors';

export const useErrorStore = defineStore('errorHandler', {
  state: () => ({
    errors: [] as BlueprintError[],
    errorAnalysis: null as ErrorAnalysis | null,
    recoveryAttempts: [] as RecoveryAttempt[],
    selectedError: null as BlueprintError | null,
  }),
  
  getters: {
    errorsByNode: (state) => {
      const result: Record<string, BlueprintError[]> = {};
      
      state.errors.forEach(error => {
        if (error.nodeId) {
          if (!result[error.nodeId]) {
            result[error.nodeId] = [];
          }
          result[error.nodeId].push(error);
        }
      });
      
      return result;
    },
    
    errorsByType: (state) => {
      const result: Record<string, BlueprintError[]> = {};
      
      state.errors.forEach(error => {
        if (!result[error.type]) {
          result[error.type] = [];
        }
        result[error.type].push(error);
      });
      
      return result;
    },
    
    recoverableErrors: (state) => {
      return state.errors.filter(error => error.recoverable);
    },
    
    hasErrors: (state) => state.errors.length > 0,
    
    hasCriticalErrors: (state) => {
      return state.errors.some(error => 
        error.severity === ErrorSeverity.Critical || 
        error.severity === ErrorSeverity.High
      );
    }
  },
  
  actions: {
    addError(error: BlueprintError) {
      // Add expanded property for UI toggling
      error.expanded = false;
      
      // Ensure we don't add duplicates
      const existingErrorIndex = this.errors.findIndex(e => 
        e.code === error.code && e.nodeId === error.nodeId && e.executionId === error.executionId
      );
      
      if (existingErrorIndex >= 0) {
        // Update existing error
        this.errors[existingErrorIndex] = error;
      } else {
        // Add new error
        this.errors.push(error);
      }
      
      // Add debug log for monitoring
      console.log('Error added to store:', error, 'Total errors:', this.errors.length);
    },
    
    updateError(error: BlueprintError) {
      // Find the error and update it
      const index = this.errors.findIndex(e => 
        e.code === error.code && e.nodeId === error.nodeId && e.executionId === error.executionId
      );
      
      if (index >= 0) {
        // Replace the error at the found index
        this.errors[index] = error;
      }
    },
    
    updateErrorAnalysis(analysis: ErrorAnalysis) {
      this.errorAnalysis = analysis;
    },
    
    addRecoveryAttempt(attempt: RecoveryAttempt) {
      this.recoveryAttempts.push(attempt);
    },
    
    clearErrors() {
      this.errors = [];
      this.errorAnalysis = null;
    },
    
    clearRecoveryAttempts() {
      this.recoveryAttempts = [];
    },
    
    selectError(error: BlueprintError) {
      this.selectedError = error;
    },
    
    clearSelectedError() {
      this.selectedError = null;
    },
    
    getErrorsForNode(nodeId: string): BlueprintError[] {
      return this.errors.filter(error => error.nodeId === nodeId);
    },
    
    getRecoveryAttemptsForNode(nodeId: string): RecoveryAttempt[] {
      return this.recoveryAttempts.filter(attempt => attempt.nodeId === nodeId);
    },
    
    async recoverFromError(error: BlueprintError, strategy?: RecoveryStrategy) {
      if (!error.recoverable) return false;
      
      // Use the first strategy if none specified
      if (!strategy && error.recoveryOptions && error.recoveryOptions.length > 0) {
        strategy = error.recoveryOptions[0];
      }
      
      try {
        const response = await fetch('/api/errors/recover', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            executionId: error.executionId,
            nodeId: error.nodeId,
            errorCode: error.code,
            strategy
          })
        });
        
        if (!response.ok) {
          throw new Error('Failed to recover from error');
        }

        return await response.json();
      } catch (err) {
        console.error('Error recovery failed:', err);
        return { success: false, error: err };
      }
    }
  }
});

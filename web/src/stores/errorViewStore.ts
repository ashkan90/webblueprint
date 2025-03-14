import { defineStore } from 'pinia';
import { useErrorStore } from './errorStore';
import { BlueprintError, RecoveryStrategy } from '../types/errors';

// Store to manage error view settings and interactions
export const useErrorViewStore = defineStore('errorView', {
  state: () => ({
    showErrorPanel: false,
    errorFilterText: '',
    selectedNodeId: '',
    autoShowErrors: true,
    errorSeverityFilters: {
      low: true,
      medium: true,
      high: true,
      critical: true
    }
  }),
  
  getters: {
    // Get filtered errors based on current settings
    filteredErrors: (state) => {
      const errorStore = useErrorStore();
      let errors = errorStore.errors;
      
      // Filter by node if selected
      if (state.selectedNodeId) {
        errors = errors.filter(err => err.nodeId === state.selectedNodeId);
      }
      
      // Filter by text
      if (state.errorFilterText) {
        const searchText = state.errorFilterText.toLowerCase();
        errors = errors.filter(err => 
          err.message.toLowerCase().includes(searchText) ||
          err.type.toLowerCase().includes(searchText) ||
          err.code.toLowerCase().includes(searchText) ||
          (err.nodeId && err.nodeId.toLowerCase().includes(searchText))
        );
      }
      
      // Filter by severity
      errors = errors.filter(err => {
        const severity = err.severity.toLowerCase();
        return state.errorSeverityFilters[severity];
      });
      
      return errors;
    },
    
    // Get error count
    errorCount: () => {
      const errorStore = useErrorStore();
      return errorStore.errors.length;
    },
    
    // Get count of critical errors
    criticalErrorCount: () => {
      const errorStore = useErrorStore();
      return errorStore.errors.filter(
        err => err.severity === 'critical' || err.severity === 'high'
      ).length;
    },
    
    // Get count of recoverable errors
    recoverableErrorCount: () => {
      const errorStore = useErrorStore();
      return errorStore.errors.filter(err => err.recoverable).length;
    }
  },
  
  actions: {
    // Toggle error panel visibility
    toggleErrorPanel() {
      this.showErrorPanel = !this.showErrorPanel;
    },
    
    // Select a node to filter errors
    selectNode(nodeId: string) {
      this.selectedNodeId = nodeId;
      if (nodeId && !this.showErrorPanel) {
        this.showErrorPanel = true;
      }
    },
    
    // Clear node selection
    clearNodeSelection() {
      this.selectedNodeId = '';
    },
    
    // Set error filter text
    setErrorFilter(text: string) {
      this.errorFilterText = text;
    },
    
    // Toggle a severity filter
    toggleSeverityFilter(severity: string) {
      this.errorSeverityFilters[severity] = !this.errorSeverityFilters[severity];
    },
    
    // Clear all error filters
    clearErrorFilters() {
      this.errorFilterText = '';
      this.errorSeverityFilters = {
        low: true,
        medium: true,
        high: true,
        critical: true
      };
      this.selectedNodeId = '';
    },
    
    // Toggle auto-show errors setting
    toggleAutoShowErrors() {
      this.autoShowErrors = !this.autoShowErrors;
    },
    
    // Attempt to recover from an error
    async recoverFromError(error: BlueprintError, strategy?: RecoveryStrategy) {
      const errorStore = useErrorStore();
      return await errorStore.recoverFromError(error, strategy);
    },
    
    // Clear all errors
    clearAllErrors() {
      const errorStore = useErrorStore();
      errorStore.clearErrors();
      errorStore.clearRecoveryAttempts();
    }
  }
});

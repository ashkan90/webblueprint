import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useErrorStore } from './errorHandler'
import { useWebSocketStore } from './websocket'

export const useErrorViewStore = defineStore('errorView', () => {
  // Dependencies
  const errorStore = useErrorStore()
  const wsStore = useWebSocketStore()
  
  // State
  const showErrorPanel = ref(false)
  const errorFilterText = ref('')
  const selectedNodeId = ref('')
  const autoShowErrors = ref(true)
  const errorSeverityFilters = ref({
    low: true,
    medium: true,
    high: true,
    critical: true
  })
  
  // Computed
  const visibleErrors = computed(() => {
    let errors = errorStore.errors
    
    // Filter by node if selected
    if (selectedNodeId.value) {
      errors = errors.filter(err => err.nodeId === selectedNodeId.value)
    }
    
    // Filter by text
    if (errorFilterText.value) {
      const searchText = errorFilterText.value.toLowerCase()
      errors = errors.filter(err => 
        err.message.toLowerCase().includes(searchText) ||
        err.type.toLowerCase().includes(searchText) ||
        err.code.toLowerCase().includes(searchText) ||
        (err.nodeId && err.nodeId.toLowerCase().includes(searchText))
      )
    }
    
    // Filter by severity
    errors = errors.filter(err => {
      const severity = err.severity.toLowerCase()
      return errorSeverityFilters.value[severity]
    })
    
    return errors
  })
  
  const errorCount = computed(() => errorStore.errors.length)
  
  const criticalErrorCount = computed(() => 
    errorStore.errors.filter(
      err => err.severity === 'critical' || err.severity === 'high'
    ).length
  )
  
  const recoverableErrorCount = computed(() => 
    errorStore.errors.filter(err => err.recoverable).length
  )
  
  // Actions
  function toggleErrorPanel() {
    showErrorPanel.value = !showErrorPanel.value
  }
  
  function selectNode(nodeId) {
    selectedNodeId.value = nodeId
    if (nodeId && !showErrorPanel.value) {
      showErrorPanel.value = true
    }
  }
  
  function clearNodeSelection() {
    selectedNodeId.value = ''
  }
  
  function setErrorFilter(text) {
    errorFilterText.value = text
  }
  
  function toggleSeverityFilter(severity) {
    errorSeverityFilters.value[severity] = !errorSeverityFilters.value[severity]
  }
  
  function clearErrorFilters() {
    errorFilterText.value = ''
    errorSeverityFilters.value = {
      low: true,
      medium: true,
      high: true,
      critical: true
    }
    selectedNodeId.value = ''
  }
  
  function toggleAutoShowErrors() {
    autoShowErrors.value = !autoShowErrors.value
  }
  
  // WebSocket message handler for error notifications
  function handleErrorNotification(notification) {
    if (notification.type === 'error') {
      errorStore.addError(notification.error)
      
      // Auto-show the error panel for critical errors
      if (autoShowErrors.value && 
          (notification.error.severity === 'critical' || 
           notification.error.severity === 'high')) {
        showErrorPanel.value = true
      }
    } else if (notification.type === 'error_analysis') {
      errorStore.updateErrorAnalysis(notification.analysis)
    } else if (notification.type === 'recovery_attempt') {
      errorStore.addRecoveryAttempt({
        errorCode: notification.errorCode,
        nodeId: notification.nodeId,
        strategy: notification.strategy,
        successful: notification.successful,
        details: notification.details,
        timestamp: new Date().toISOString()
      })
    }
  }
  
  // Recovery
  async function recoverFromError(error, strategy) {
    if (!error.recoverable) return
    
    // Use the first strategy if none specified
    if (!strategy && error.recoveryOptions.length > 0) {
      strategy = error.recoveryOptions[0]
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
      })
      
      if (!response.ok) {
        throw new Error('Failed to recover from error')
      }
      
      const result = await response.json()
      return result
    } catch (err) {
      console.error('Error recovery failed:', err)
      return { success: false, error: err.message }
    }
  }
  
  // Set up WebSocket listeners
  function setupWebSocketListeners() {
    wsStore.on('debug.data', handleErrorNotification)
    wsStore.on('node.error', handleErrorNotification)
    wsStore.on('recovery_attempt', handleErrorNotification)
  }
  
  // Clean up
  function clear() {
    errorStore.clearErrors()
    clearErrorFilters()
  }
  
  return {
    // State
    showErrorPanel,
    errorFilterText,
    selectedNodeId,
    autoShowErrors,
    errorSeverityFilters,
    
    // Computed
    visibleErrors,
    errorCount,
    criticalErrorCount,
    recoverableErrorCount,
    
    // Actions
    toggleErrorPanel,
    selectNode,
    clearNodeSelection,
    setErrorFilter,
    toggleSeverityFilter,
    clearErrorFilters,
    toggleAutoShowErrors,
    handleErrorNotification,
    recoverFromError,
    setupWebSocketListeners,
    clear
  }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// Types
export interface BlueprintError {
  type: string
  code: string
  message: string
  details?: Record<string, any>
  severity: string
  recoverable: boolean
  recoveryOptions?: string[]
  nodeId?: string
  pinId?: string
  blueprintId?: string
  executionId?: string
  timestamp: string
  stackTrace?: string[]
  expanded?: boolean
}

export interface ErrorAnalysis {
  totalErrors: number
  recoverableErrors: number
  typeBreakdown: Record<string, number>
  severityBreakdown: Record<string, number>
  topProblemNodes: Array<{nodeId: string, count: number}>
  mostCommonCodes: Record<string, number>
  timestamp: string
}

export interface RecoveryAttempt {
  errorId: string
  errorCode: string
  nodeId: string
  strategy: string
  successful: boolean
  timestamp: string
  details?: Record<string, any>
}

// Store
export const useErrorStore = defineStore('errorHandler', () => {
  // State
  const errors = ref<BlueprintError[]>([])
  const errorAnalysis = ref<ErrorAnalysis | null>(null)
  const recoveryAttempts = ref<RecoveryAttempt[]>([])
  const selectedError = ref<BlueprintError | null>(null)
  
  // Computed
  const errorsByNode = computed(() => {
    const result: Record<string, BlueprintError[]> = {}
    
    errors.value.forEach(error => {
      if (error.nodeId) {
        if (!result[error.nodeId]) {
          result[error.nodeId] = []
        }
        result[error.nodeId].push(error)
      }
    })
    
    return result
  })
  
  const errorsByType = computed(() => {
    const result: Record<string, BlueprintError[]> = {}
    
    errors.value.forEach(error => {
      if (!result[error.type]) {
        result[error.type] = []
      }
      result[error.type].push(error)
    })
    
    return result
  })
  
  const recoverableErrors = computed(() => {
    return errors.value.filter(error => error.recoverable)
  })
  
  const hasErrors = computed(() => errors.value.length > 0)
  
  const hasCriticalErrors = computed(() => {
    return errors.value.some(error => error.severity === 'critical' || error.severity === 'high')
  })
  
  // Actions
  function addError(error: BlueprintError) {
    // Add expanded property for UI toggling
    error.expanded = false
    errors.value.push(error)
  }
  
  function updateErrorAnalysis(analysis: ErrorAnalysis) {
    errorAnalysis.value = analysis
  }
  
  function addRecoveryAttempt(attempt: RecoveryAttempt) {
    recoveryAttempts.value.push(attempt)
  }
  
  function clearErrors() {
    errors.value = []
    errorAnalysis.value = null
  }
  
  function clearRecoveryAttempts() {
    recoveryAttempts.value = []
  }
  
  function selectError(error: BlueprintError) {
    selectedError.value = error
  }
  
  function clearSelectedError() {
    selectedError.value = null
  }
  
  function getErrorsForNode(nodeId: string): BlueprintError[] {
    return errors.value.filter(error => error.nodeId === nodeId)
  }
  
  function getRecoveryAttemptsForNode(nodeId: string): RecoveryAttempt[] {
    return recoveryAttempts.value.filter(attempt => attempt.nodeId === nodeId)
  }
  
  // Return store
  return {
    // State
    errors,
    errorAnalysis,
    recoveryAttempts,
    selectedError,
    
    // Computed
    errorsByNode,
    errorsByType,
    recoverableErrors,
    hasErrors,
    hasCriticalErrors,
    
    // Actions
    addError,
    updateErrorAnalysis,
    addRecoveryAttempt,
    clearErrors,
    clearRecoveryAttempts,
    selectError,
    clearSelectedError,
    getErrorsForNode,
    getRecoveryAttemptsForNode
  }
})

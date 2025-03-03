// Node execution status
export interface NodeExecutionStatus {
    nodeId: string
    status: 'idle' | 'executing' | 'completed' | 'error'
    timestamp: Date
    message?: string
    errorDetails?: any
}

// Debug snapshot
export interface DebugSnapshot {
    timestamp: Date
    description: string
    data: any
}

// Node debug data
export interface NodeDebugData {
    nodeId: string
    executionId: string
    timestamp: Date
    inputs?: Record<string, any>
    outputs?: Record<string, any>
    internalState?: Record<string, any>
    snapshots: DebugSnapshot[]
}

// Data flow between nodes
export interface DataFlow {
    sourceNodeId: string
    sourcePinId: string
    targetNodeId: string
    targetPinId: string
    value: any
    timestamp: Date
}

// Execution result
export interface ExecutionResult {
    executionId: string
    success: boolean
    startTime: Date
    endTime: Date
    error?: string
}

export interface LogEntry {
    level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'
    timestamp: Date
    nodeId: string
    message: string
    details?: any
    expanded?: boolean
}
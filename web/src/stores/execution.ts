import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWebSocketStore, WebSocketEvents } from './websocket'
import type {NodeExecutionStatus, NodeDebugData, DataFlow, LogEntry} from '../types/execution'
import { useBlueprintStore } from './blueprint'

type NodeStatus = 'idle' | 'executing' | 'completed' | 'error'

function isNodeStatusData(data: unknown): data is {
    nodeId: string,
    status?: NodeStatus,
    timestamp?: string | Date,
    message?: string,
    errorDetails?: any
} {
    return typeof data === 'object' && data !== null && 'nodeId' in data
}

function normalizeNodeStatus(status: unknown): NodeStatus {
    const validStatuses: NodeStatus[] = ['idle', 'executing', 'completed', 'error']
    return validStatuses.includes(status as NodeStatus)
        ? status as NodeStatus
        : 'idle'
}

export const useExecutionStore = defineStore('execution', () => {
    const currentExecutionId = ref<string | null>(null)
    const blueprintId = ref<string | null>(null)
    const executionStatus = ref<'idle' | 'running' | 'completed' | 'error'>('idle')
    const nodeStatuses = ref<Record<string, NodeExecutionStatus>>({})
    const debugData = ref<Record<string, NodeDebugData>>({})
    const dataFlows = ref<DataFlow[]>([])
    const executionStartTime = ref<Date | null>(null)
    const executionEndTime = ref<Date | null>(null)
    const errorMessage = ref<string | null>(null)
    const logs = ref<LogEntry[]>([])


    const websocketStore = useWebSocketStore()
    const blueprintStore = useBlueprintStore()

    function safeDate(dateInput: string | Date | undefined): Date {
        if (!dateInput) return new Date()
        return dateInput instanceof Date ? dateInput : new Date(dateInput)
    }

    function safeString(value: unknown): string {
        return typeof value === 'string' ? value : ''
    }

    function safeBoolean(status: unknown): boolean {
        return typeof status === 'boolean' ? status : false
    }

    const isExecuting = computed(() => executionStatus.value === 'running')

    const getNodeStatus = computed(() =>
        (nodeId: string) => nodeStatuses.value[nodeId] || null
    )

    const getNodeDebugData = computed(() =>
        (nodeId: string) => debugData.value[nodeId] || null
    )

    const executionDuration = computed(() => {
        if (!executionStartTime.value) return null

        const endTime = executionEndTime.value || new Date()
        return endTime.getTime() - executionStartTime.value.getTime()
    })

    function startExecution(executionId: string) {
        currentExecutionId.value = executionId
        executionStatus.value = 'running'
        executionStartTime.value = new Date()
        executionEndTime.value = null
        errorMessage.value = null

        nodeStatuses.value = {}
        debugData.value = {}
        dataFlows.value = []
    }

    function endExecution(status: 'completed' | 'error', error?: string) {
        executionStatus.value = status
        executionEndTime.value = new Date()
        errorMessage.value = error || null
    }

    function updateNodeStatus(data: unknown) {
        if (!isNodeStatusData(data)) return

        const nodeStatus: NodeExecutionStatus = {
            nodeId: data.nodeId,
            status: normalizeNodeStatus(data.status),
            timestamp: safeDate(data.timestamp),
            message: safeString(data.message),
            errorDetails: data.errorDetails
        }

        nodeStatuses.value[nodeStatus.nodeId] = nodeStatus
    }

    function updateNodeDebugData(rawData: unknown) {
        const data = rawData as Partial<NodeDebugData>
        if (!data || !data.nodeId) return

        const nodeData: NodeDebugData = {
            nodeId: data.nodeId || '',
            executionId: data.executionId || '',
            timestamp: safeDate(data.timestamp),
            inputs: data.inputs || {},
            outputs: data.outputs || {},
            internalState: data.internalState || {},
            snapshots: data.snapshots || []
        }

        debugData.value[nodeData.nodeId] = nodeData
    }

    function addLogEntry(log: LogEntry) {
        logs.value.push(log)
    }

    function clearLogs() {
        logs.value = []
    }

    function toggleLogExpanded(index: number) {
        if (logs.value[index]) {
            logs.value[index].expanded = !logs.value[index].expanded
        }
    }

    function clearDebugData() {
        nodeStatuses.value = {}
        debugData.value = {}
        dataFlows.value = []
        currentExecutionId.value = null
        executionStatus.value = 'idle'
        executionStartTime.value = null
        executionEndTime.value = null
    }

    // Fetch debug data for a node
    async function fetchNodeDebugData(executionId: string, nodeId: string) {
        try {
            const response = await fetch(`/api/executions/${executionId}/nodes/${nodeId}`)

            if (!response.ok) {
                throw new Error(`Failed to fetch debug data`)
            }

            const data = await response.json()

            updateNodeDebugData({
                nodeId: data.nodeId,
                executionId: data.executionId,
                timestamp: new Date(),
                inputs: data.debug?.inputs,
                outputs: data.debug?.outputs,
                internalState: data.debug?.internalState,
                snapshots: data.debug?.snapshots || []
            })

            return data
        } catch (error) {
            console.error('Error fetching node debug data:', error)
            throw error
        }
    }

    function recordDataFlow(flow: DataFlow) {
        // Store the flow
        dataFlows.value.push(flow)

        // Update node statuses to reflect data flow
        // This is important for visualizing active connections

        // Ensure we have statuses for both source and target nodes
        if (!nodeStatuses.value[flow.sourceNodeId]) {
            nodeStatuses.value[flow.sourceNodeId] = {
                nodeId: flow.sourceNodeId,
                status: 'completed', // Assume completed if we're getting data from it
                timestamp: flow.timestamp
            }
        }

        if (!nodeStatuses.value[flow.targetNodeId]) {
            nodeStatuses.value[flow.targetNodeId] = {
                nodeId: flow.targetNodeId,
                status: 'executing', // Assume executing if we're sending data to it
                timestamp: flow.timestamp
            }
        }
    }

    async function loadExecution(executionId: string) {
        try {
            const response = await fetch(`/api/executions/${executionId}`)

            if (!response.ok) {
                throw new Error(`Failed to load execution`)
            }

            const data = await response.json()

            currentExecutionId.value = executionId
            blueprintId.value = data.blueprintId
            executionStatus.value = data.status
            executionStartTime.value = new Date(data.startTime)
            executionEndTime.value = data.endTime ? new Date(data.endTime) : null
            errorMessage.value = data.error || null

            if (typeof data.nodes === 'object' && data.nodes !== null) {
                Object.entries(data.nodes as Record<string, unknown>).forEach(([nodeId, nodeData]) => {
                    updateNodeStatus({
                        nodeId,
                        ...(typeof nodeData === 'object' && nodeData !== null ? nodeData : {})
                    })
                })
            }

            if (data.blueprintId && (!blueprintStore.blueprint.id || blueprintStore.blueprint.id !== data.blueprintId)) {
                await blueprintStore.loadBlueprint(data.blueprintId)
            }

            return data
        } catch (error) {
            console.error('Error loading execution:', error)
            throw error
        }
    }

    function setupWebSocketListeners() {
        websocketStore.on(WebSocketEvents.EXEC_START, (data: unknown) => {
            const execData = data as { executionId?: string }
            if (execData.executionId) {
                startExecution(execData.executionId)
            }
        })

        websocketStore.on(WebSocketEvents.EXEC_END, (data: unknown) => {
            const execData = data as { success?: boolean, error?: string }
            endExecution(
                safeBoolean(execData.success) ? 'completed' : 'error',
                safeString(execData.error)
            )
        })

        websocketStore.on(WebSocketEvents.NODE_START, updateNodeStatus)
        websocketStore.on(WebSocketEvents.NODE_COMPLETE, updateNodeStatus)
        websocketStore.on(WebSocketEvents.NODE_ERROR, updateNodeStatus)
        websocketStore.on(WebSocketEvents.DEBUG_DATA, updateNodeDebugData)
        websocketStore.on(WebSocketEvents.DATA_FLOW, recordDataFlow)
        websocketStore.on(WebSocketEvents.LOG, (data: any) => {
            addLogEntry({
                level: data.level.toUpperCase(),
                timestamp: new Date(data.timestamp),
                nodeId: data.nodeId,
                message: data.message,
                details: data.fields,
                expanded: false
            })
        })
    }

    setupWebSocketListeners()

    return {
        currentExecutionId,
        blueprintId,
        executionStatus,
        nodeStatuses,
        debugData,
        dataFlows,
        executionStartTime,
        executionEndTime,
        errorMessage,
        isExecuting,
        getNodeStatus,
        getNodeDebugData,
        updateNodeStatus,
        updateNodeDebugData,
        recordDataFlow,
        clearDebugData,
        fetchNodeDebugData,
        logs,
        addLogEntry,
        clearLogs,
        toggleLogExpanded,
        executionDuration,
        startExecution,
        endExecution,
        loadExecution,
        executeBlueprint: async (blueprintId: string, initialData?: Record<string, any>) => {
            try {
                const response = await fetch(`/api/blueprints/${blueprintId}/execute`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: initialData ? JSON.stringify({ initialVariables: initialData }) : undefined
                })

                if (!response.ok) {
                    const errorText = await response.text()
                    throw new Error(`Failed to execute blueprint: ${errorText}`)
                }

                return await response.json()
            } catch (error) {
                console.error('Error executing blueprint:', error)
                throw error
            }
        }
    }
})
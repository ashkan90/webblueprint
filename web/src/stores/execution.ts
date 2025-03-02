import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWebSocketStore, WebSocketEvents } from './websocket'
import type { NodeExecutionStatus, NodeDebugData, DataFlow } from '../types/execution'

export const useExecutionStore = defineStore('execution', () => {
    // State
    const currentExecutionId = ref<string | null>(null)
    const blueprintId = ref<string | null>(null)
    const executionStatus = ref<'idle' | 'running' | 'completed' | 'error'>('idle')
    const nodeStatuses = ref<Record<string, NodeExecutionStatus>>({})
    const debugData = ref<Record<string, NodeDebugData>>({})
    const dataFlows = ref<DataFlow[]>([])
    const executionStartTime = ref<Date | null>(null)
    const executionEndTime = ref<Date | null>(null)

    // WebSocket store
    const websocketStore = useWebSocketStore()

    // Getters
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

    // Actions
    function startExecution(executionId: string) {
        currentExecutionId.value = executionId
        executionStatus.value = 'running'
        executionStartTime.value = new Date()
        executionEndTime.value = null

        // Clear previous state
        nodeStatuses.value = {}
        debugData.value = {}
        dataFlows.value = []
    }

    function endExecution(status: 'completed' | 'error') {
        executionStatus.value = status
        executionEndTime.value = new Date()
    }

    function updateNodeStatus(status: NodeExecutionStatus) {
        nodeStatuses.value[status.nodeId] = status
    }

    function updateNodeDebugData(data: NodeDebugData) {
        const existingData = debugData.value[data.nodeId]

        if (existingData) {
            // Merge with existing data
            existingData.snapshots.push(...data.snapshots)
            existingData.inputs = { ...existingData.inputs, ...data.inputs }
            existingData.outputs = { ...existingData.outputs, ...data.outputs }
            existingData.internalState = { ...existingData.internalState, ...data.internalState }
            existingData.timestamp = data.timestamp
        } else {
            // Add new data
            debugData.value[data.nodeId] = data
        }
    }

    function recordDataFlow(flow: DataFlow) {
        dataFlows.value.push(flow)
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

    // Initialize WebSocket listeners
    function setupWebSocketListeners() {
        websocketStore.on(WebSocketEvents.EXEC_START, (data) => {
            startExecution(data.executionId)
        })

        websocketStore.on(WebSocketEvents.EXEC_END, (data) => {
            endExecution(data.success ? 'completed' : 'error')
        })

        websocketStore.on(WebSocketEvents.NODE_START, (data) => {
            updateNodeStatus({
                nodeId: data.nodeId,
                status: 'executing',
                timestamp: new Date(data.timestamp),
                message: data.message
            })
        })

        websocketStore.on(WebSocketEvents.NODE_COMPLETE, (data) => {
            updateNodeStatus({
                nodeId: data.nodeId,
                status: 'completed',
                timestamp: new Date(data.timestamp),
                message: data.message
            })
        })

        websocketStore.on(WebSocketEvents.NODE_ERROR, (data) => {
            updateNodeStatus({
                nodeId: data.nodeId,
                status: 'error',
                timestamp: new Date(data.timestamp),
                message: data.message,
                errorDetails: data.error
            })
        })

        websocketStore.on(WebSocketEvents.DEBUG_DATA, (data) => {
            updateNodeDebugData({
                nodeId: data.nodeId,
                executionId: data.executionId,
                timestamp: new Date(data.timestamp),
                inputs: data.debugData?.inputs,
                outputs: data.debugData?.outputs,
                internalState: data.debugData?.internalState,
                snapshots: [
                    {
                        timestamp: new Date(data.timestamp),
                        description: data.debugData?.description || 'Debug snapshot',
                        data: data.debugData
                    }
                ]
            })
        })

        websocketStore.on(WebSocketEvents.DATA_FLOW, (data) => {
            recordDataFlow({
                sourceNodeId: data.sourceNodeId,
                sourcePinId: data.sourcePinId,
                targetNodeId: data.targetNodeId,
                targetPinId: data.targetPinId,
                value: data.value,
                timestamp: new Date(data.timestamp)
            })
        })
    }

    // Function to execute a blueprint
    async function executeBlueprint(blueprintId: string, initialData?: Record<string, any>) {
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

    // Call setup when store is created
    setupWebSocketListeners()

    return {
        currentExecutionId,
        executionStatus,
        nodeStatuses,
        debugData,
        dataFlows,
        executionStartTime,
        executionEndTime,
        isExecuting,
        getNodeStatus,
        getNodeDebugData,
        executionDuration,
        clearDebugData,
        executeBlueprint,
        fetchNodeDebugData
    }
})
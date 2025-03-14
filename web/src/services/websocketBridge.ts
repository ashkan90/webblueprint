/**
 * WebSocket Execution Bridge
 *
 * This service bridges between the client-side execution and the WebSocket communication,
 * ensuring that execution events are properly synchronized across the system.
 */

import { useWebSocketStore, WebSocketEvents } from '../stores/websocket';
import { useExecutionStore } from '../stores/execution';
import { ExecutionManager } from '../utils/executionManager';
import type { Node, Connection } from '../types/blueprint';
import type { DataFlow } from '../types/execution';

/**
 * WebSocketExecutionBridge connects local execution with the WebSocket-based events
 */
export class WebSocketExecutionBridge {
    private executionManager: ExecutionManager;

    private websocketStore = useWebSocketStore();
    private isInitialized = false;

    constructor(executionManager: ExecutionManager) {
        this.executionManager = executionManager
    }
    /**
     * Initialize the WebSocket bridge
     */
    initialize(): void {
        if (this.isInitialized) return;

        this.setupEventListeners();
        this.isInitialized = true;
        console.log('WebSocket execution bridge initialized');
    }

    /**
     * Set up event listeners for WebSocket events
     */
    private setupEventListeners(): void {
        // Node execution status events
        this.websocketStore.on(WebSocketEvents.NODE_START, this.handleNodeStart.bind(this));
        this.websocketStore.on(WebSocketEvents.NODE_COMPLETE, this.handleNodeComplete.bind(this));
        this.websocketStore.on(WebSocketEvents.NODE_ERROR, this.handleNodeError.bind(this));

        // Data flow events
        this.websocketStore.on(WebSocketEvents.DATA_FLOW, this.handleDataFlow.bind(this));

        // Execution lifecycle events
        this.websocketStore.on(WebSocketEvents.EXEC_START, this.handleExecutionStart.bind(this));
        this.websocketStore.on(WebSocketEvents.EXEC_END, this.handleExecutionEnd.bind(this));

        // Debug data
        this.websocketStore.on(WebSocketEvents.DEBUG_DATA, this.handleDebugData.bind(this));
    }

    /**
     * Handle node start event
     */
    private handleNodeStart(data: any): void {
        const nodeId = data?.nodeId;
        if (!nodeId) return;

        // Update execution store
        this.executionManager.executionStore.updateNodeStatus({
            nodeId,
            status: 'executing',
            timestamp: new Date(data.timestamp || Date.now()),
            message: data.message
        });

        // Update execution manager for visualization
        this.executionManager.trackNodeExecution(nodeId, 'executing');
    }

    /**
     * Handle node complete event
     */
    private handleNodeComplete(data: any): void {
        const nodeId = data?.nodeId;
        if (!nodeId) return;

        // Update execution store
        this.executionManager.getExecutionStore().updateNodeStatus({
            nodeId,
            status: 'completed',
            timestamp: new Date(data.timestamp || Date.now()),
            message: data.message
        });

        // Update execution manager for visualization
        this.executionManager.trackNodeExecution(nodeId, 'completed');
    }

    /**
     * Handle node error event
     */
    private handleNodeError(data: any): void {
        const nodeId = data?.nodeId;
        if (!nodeId) return;

        // Update execution store
        this.executionManager.getExecutionStore().updateNodeStatus({
            nodeId,
            status: 'error',
            timestamp: new Date(data.timestamp || Date.now()),
            message: data.error || data.message,
            errorDetails: data.error || data.errorDetails
        });

        // Update execution manager for visualization
        this.executionManager.trackNodeExecution(nodeId, 'error');
    }

    /**
     * Handle data flow event
     */
    private handleDataFlow(data: any): void {
        // Ensure we have all required fields
        if (!data.sourceNodeId || !data.sourcePinId || !data.targetNodeId || !data.targetPinId) {
            return;
        }

        // Create data flow object
        const dataFlow: DataFlow = {
            sourceNodeId: data.sourceNodeId,
            sourcePinId: data.sourcePinId,
            targetNodeId: data.targetNodeId,
            targetPinId: data.targetPinId,
            value: data.value,
            timestamp: new Date(data.timestamp || Date.now())
        };

        // Add to execution store
        this.executionManager.getExecutionStore().recordDataFlow(dataFlow);

        // Find the corresponding connection
        // We need to identify it to animate the connection
        const connectionData = {
            sourceNodeId: data.sourceNodeId,
            sourcePinId: data.sourcePinId,
            targetNodeId: data.targetNodeId,
            targetPinId: data.targetPinId
        };

        // Let the execution manager know about this data flow
        // It will handle the animation
        this.activateConnectionForDataFlow(connectionData);
    }

    /**
     * Handle execution start event
     */
    private handleExecutionStart(data: any): void {
        const executionId = data?.executionID;
        if (!executionId) return;

        // Start execution in the store
        this.executionManager.getExecutionStore().startExecution(executionId);

        // Clear any previous execution visualization state
        this.executionManager.clearActiveStates();
    }

    /**
     * Handle execution end event
     */
    private handleExecutionEnd(data: any): void {
        // End execution in the store
        this.executionManager.getExecutionStore().endExecution(
            data.success ? 'completed' : 'error',
            data.error || data.errorMessage
        );
    }

    /**
     * Handle debug data event
     */
    private handleDebugData(data: any): void {
        if (!data.nodeId) return;

        // Update node debug data in the store
        this.executionManager.getExecutionStore().updateNodeDebugData({
            nodeId: data.nodeId,
            executionId: data.executionId,
            timestamp: new Date(data.timestamp || Date.now()),
            inputs: data.inputs,
            outputs: data.outputs,
            snapshots: data.snapshots || []
        });
    }

    /**
     * Activate a connection to show data flow
     */
    private activateConnectionForDataFlow(connectionData: {
        sourceNodeId: string;
        sourcePinId: string;
        targetNodeId: string;
        targetPinId: string;
    }): void {
        // Find the actual connection object from the blueprint store
        // This would usually come from a store, but for this example we'll mock it
        const mockConnection: Connection = {
            id: `${connectionData.sourceNodeId}-${connectionData.sourcePinId}-${connectionData.targetNodeId}-${connectionData.targetPinId}`,
            ...connectionData,
            connectionType: 'data'
        };

        // Activate the connection in the execution manager
        this.executionManager.activateConnection(mockConnection);

        // Animate data flow along this connection
        // We'll use "string" as the source type, but this could be determined from the actual pin type
        this.executionManager.animateDataFlow(mockConnection, 'string');
    }

    /**
     * Send execution event to the server
     */
    sendExecutionEvent(eventType: string, data: any): void {
        this.websocketStore.send(eventType, data);
    }
}

export let websocketBridge: WebSocketExecutionBridge;
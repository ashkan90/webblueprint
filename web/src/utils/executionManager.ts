// This utility helps with managing execution flow and visualization

import { useExecutionStore } from '../stores/execution';
import { useBlueprintStore } from '../stores/blueprint';
import { useNodeRegistryStore } from '../stores/nodeRegistry';
import type { Connection } from '../types/blueprint';

/**
 * ExecutionManager helps track and visualize execution flow
 */
export class ExecutionManager {
    executionStore = useExecutionStore();
    blueprintStore = useBlueprintStore();
    nodeRegistryStore = useNodeRegistryStore();

    private activeConnections = new Set<string>();
    private activeNodes = new Set<string>();
    private activePins = new Set<string>();

    /**
     * Tracks a node's execution state
     */
    trackNodeExecution(nodeId: string, status: 'executing' | 'completed' | 'error'): void {
        if (status === 'executing') {
            this.activeNodes.add(nodeId);
            this.updatePinVisualization(nodeId, 'input', true);
        } else {
            // For completed or error, we briefly highlight output pins
            this.updatePinVisualization(nodeId, 'output', true);

            // Then remove inputs from active set after a delay
            setTimeout(() => {
                this.updatePinVisualization(nodeId, 'input', false);
            }, 1000);

            // And finally remove outputs after another delay
            setTimeout(() => {
                this.updatePinVisualization(nodeId, 'output', false);
                this.activeNodes.delete(nodeId);
            }, 2000);
        }
    }

    /**
     * Updates pin visualization based on active state
     */
    private updatePinVisualization(nodeId: string, pinType: 'input' | 'output', isActive: boolean): void {
        const node = this.blueprintStore.getNodeById(nodeId);
        if (!node) return;

        const nodeType = this.nodeRegistryStore.getNodeTypeById(node.type);
        if (!nodeType) return;

        const pins = pinType === 'input' ? nodeType.inputs : nodeType.outputs;

        pins.forEach(pin => {
            const pinId = `${nodeId}-${pin.id}`;
            if (isActive) {
                this.activePins.add(pinId);
            } else {
                this.activePins.delete(pinId);
            }
        });
    }

    /**
     * Activates a connection to visualize data flow
     */
    activateConnection(connection: Connection, duration: number = 1000): void {
        this.activeConnections.add(connection.id);

        // Auto-deactivate after duration
        setTimeout(() => {
            this.activeConnections.delete(connection.id);
        }, duration);
    }

    /**
     * Creates a data flow animation along a connection
     */
    animateDataFlow(connection: Connection, sourceType: string): void {
        // First activate the connection
        this.activateConnection(connection, 1500);

        // The actual animation will be handled by the BlueprintCanvas component
        // We just need to trigger it via the execution store
        this.executionStore.recordDataFlow({
            sourceNodeId: connection.sourceNodeId,
            sourcePinId: connection.sourcePinId,
            targetNodeId: connection.targetNodeId,
            targetPinId: connection.targetPinId,
            value: null, // This could be enhanced to show actual value
            timestamp: new Date()
        });
    }

    /**
     * Gets all currently active connections
     */
    getActiveConnections(): Set<string> {
        return this.activeConnections;
    }

    /**
     * Gets all currently active nodes
     */
    getActiveNodes(): Set<string> {
        return this.activeNodes;
    }

    /**
     * Gets all currently active pins
     */
    getActivePins(): Set<string> {
        return this.activePins;
    }

    /**
     * Clears all active states
     */
    clearActiveStates(): void {
        this.activeConnections.clear();
        this.activeNodes.clear();
        this.activePins.clear();
    }
}
import { ExecutionManager } from '../utils/executionManager';
import { domHandler } from '../utils/domHandler';
import { storageHandler } from '../utils/storageHandler';
import type { Node, Connection } from '../types/blueprint';

/**
 * ExecutionMode determines how blueprints are executed
 */
export enum ExecutionMode {
    /**
     * Direct mode executes nodes in the browser where possible,
     * falling back to server-side execution when necessary
     */
    DIRECT = 'direct',

    /**
     * Actor mode uses the actor system on the server for all execution,
     * which provides better isolation and reliability
     */
    ACTOR = 'actor'
}

/**
 * ExecutionService handles starting and monitoring blueprint executions
 */
export class ExecutionService {
    private executionManager: ExecutionManager;
    private executionMode: ExecutionMode = ExecutionMode.DIRECT; // Default to direct mode

    constructor(executionManager: ExecutionManager) {
        this.executionManager = executionManager
    }

    // Map of node types that can be executed client-side
    private clientSideNodeTypes: Set<string> = new Set([
        'dom-element',
        'dom-event',
        'storage',
        'constant-string',
        'constant-number',
        'constant-boolean',
        // Could add more client-side executable node types here
    ]);

    /**
     * Set the execution mode (direct or actor-based)
     */
    setExecutionMode(mode: ExecutionMode): void {
        this.executionMode = mode;
        console.log(`Execution mode set to: ${mode}`);
    }

    /**
     * Get the current execution mode
     */
    getExecutionMode(): ExecutionMode {
        return this.executionMode;
    }

    /**
     * Execute a blueprint
     */
    async executeBlueprint(blueprintId: string, initialData?: Record<string, any>): Promise<any> {
        // Clear any previous execution state
        this.executionManager.clearActiveStates();

        // Start execution based on the mode
        if (this.executionMode === ExecutionMode.ACTOR) {
            return this.executeWithActorSystem(blueprintId, initialData);
        } else {
            return this.executeWithDirectMode(blueprintId, initialData);
        }
    }

    /**
     * Execute a blueprint using the actor system (server-side)
     */
    private async executeWithActorSystem(blueprintId: string, initialData?: Record<string, any>): Promise<any> {
        // This just calls the server-side execution endpoint
        return this.executionManager.executionStore.executeBlueprint(blueprintId, initialData);
    }

    /**
     * Execute a blueprint in direct mode (client-side where possible)
     */
    private async executeWithDirectMode(blueprintId: string, initialData?: Record<string, any>): Promise<any> {
        // First, register that we're starting execution
        const executionId = `client-${Date.now()}`;
        this.executionManager.executionStore.startExecution(executionId);

        try {
            // Find entry points (nodes with no execution inputs)
            const entryPoints = this.findEntryPoints();
            if (entryPoints.length === 0) {
                throw new Error('No entry points found in blueprint');
            }

            // Process each entry point sequentially
            for (const nodeId of entryPoints) {
                await this.executeNode(nodeId);
            }

            // Mark execution as completed
            this.executionManager.executionStore.endExecution('completed');

            return {
                executionId,
                success: true,
                message: 'Blueprint executed successfully'
            };
        } catch (error) {
            // Mark execution as failed
            this.executionManager.executionStore.endExecution(
                'error',
                error instanceof Error ? error.message : String(error)
            );

            return {
                executionId,
                success: false,
                error: error instanceof Error ? error.message : String(error)
            };
        }
    }

    /**
     * Find entry points in the blueprint
     */
    private findEntryPoints(): string[] {
        return this.executionManager.blueprintStore.findEntryPoints;
    }

    /**
     * Execute a single node
     */
    private async executeNode(nodeId: string): Promise<void> {
        const node = this.executionManager.blueprintStore.getNodeById(nodeId);
        if (!node) {
            throw new Error(`Node not found: ${nodeId}`);
        }

        // Mark node as executing
        this.executionManager.executionStore.updateNodeStatus({
            nodeId,
            status: 'executing',
            timestamp: new Date()
        });

        // Track in execution manager for visualization
        this.executionManager.trackNodeExecution(nodeId, 'executing');

        try {
            // Check if this node type can be executed client-side
            if (this.canExecuteClientSide(node)) {
                await this.executeClientSideNode(node);
            } else {
                // For nodes that need server-side execution, we'd send them to the server
                // and wait for a response, but for now let's assume it succeeds
                await this.mockServerExecution(node);
            }

            // Node completed successfully
            this.executionManager.executionStore.updateNodeStatus({
                nodeId,
                status: 'completed',
                timestamp: new Date()
            });

            // Track in execution manager
            this.executionManager.trackNodeExecution(nodeId, 'completed');

            // Find and execute outgoing execution connections
            await this.followExecutionConnections(nodeId);

        } catch (error) {
            // Node execution failed
            this.executionManager.executionStore.updateNodeStatus({
                nodeId,
                status: 'error',
                timestamp: new Date(),
                message: error instanceof Error ? error.message : String(error)
            });

            // Track in execution manager
            this.executionManager.trackNodeExecution(nodeId, 'error');

            throw error; // Re-throw to stop execution
        }
    }

    /**
     * Check if a node can be executed client-side
     */
    private canExecuteClientSide(node: Node): boolean {
        return this.clientSideNodeTypes.has(node.type);
    }

    /**
     * Execute a node on the client side
     */
    private async executeClientSideNode(node: Node): Promise<void> {
        // Handle different node types
        switch (node.type) {
            case 'dom-element':
                return this.executeDomElementNode(node);

            case 'dom-event':
                return this.executeDomEventNode(node);

            case 'storage':
                return this.executeStorageNode(node);

            case 'constant-string':
            case 'constant-number':
            case 'constant-boolean':
                return this.executeConstantNode(node);

            default:
                throw new Error(`Unsupported client-side node type: ${node.type}`);
        }
    }

    /**
     * Execute a DOM Element node
     */
    private async executeDomElementNode(node: Node): Promise<void> {
        // Extract operation details from node properties
        const operation = this.buildDomOperation(node);

        // Execute the operation using DOM handler
        const result = domHandler.processOperation(operation);

        // Store the result for debug and data flow visualization
        this.executionManager.executionStore.updateNodeDebugData({
            nodeId: node.id,
            outputs: { element: result },
            timestamp: new Date()
        });

        return Promise.resolve();
    }

    /**
     * Execute a DOM Event node
     */
    private async executeDomEventNode(node: Node): Promise<void> {
        // Extract operation details from node properties
        const operation = this.buildDomEventOperation(node);

        // Register the event handler
        domHandler.processEventOperation(operation);

        // For event nodes, we're just setting up the listener, so success is immediate
        return Promise.resolve();
    }

    /**
     * Execute a Storage node
     */
    private async executeStorageNode(node: Node): Promise<void> {
        // Extract operation details from node properties
        const operation = this.buildStorageOperation(node);

        // Execute the operation
        const result = storageHandler.executeOperation(operation);

        // Store the result for debug and data flow visualization
        this.executionManager.executionStore.updateNodeDebugData({
            nodeId: node.id,
            outputs: { result },
            timestamp: new Date()
        });

        return Promise.resolve();
    }

    /**
     * Execute a Constant node (string, number, boolean)
     */
    private async executeConstantNode(node: Node): Promise<void> {
        // Find the constant value property
        const constantProp = node.properties.find(p => p.name === 'constantValue');
        let value = constantProp ? constantProp.value : null;

        // If no explicit value is set, use a type-appropriate default
        if (value === undefined || value === null) {
            switch (node.type) {
                case 'constant-string': value = ''; break;
                case 'constant-number': value = 0; break;
                case 'constant-boolean': value = false; break;
            }
        }

        // Store the value for data flow
        this.executionManager.executionStore.updateNodeDebugData({
            nodeId: node.id,
            outputs: { value },
            timestamp: new Date()
        });

        return Promise.resolve();
    }

    /**
     * Mock server-side execution with a delay
     */
    private async mockServerExecution(node: Node): Promise<void> {
        // Simulate server processing time
        await new Promise(resolve => setTimeout(resolve, 300 + Math.random() * 700));

        // TODO: In a real implementation, this would communicate with the server
        // and wait for the execution result

        // For now, just record some fake output
        this.executionManager.executionStore.updateNodeDebugData({
            nodeId: node.id,
            outputs: { result: `Mock output for ${node.type}` },
            timestamp: new Date()
        });

        return Promise.resolve();
    }

    /**
     * Follow and execute any execution connections from a node
     */
    private async followExecutionConnections(nodeId: string): Promise<void> {
        // Find outgoing execution connections
        const connections = this.executionManager.blueprintStore.connections.filter(
            conn => conn.sourceNodeId === nodeId && conn.connectionType === 'execution'
        );

        // Execute each target node sequentially
        for (const connection of connections) {
            // Highlight the connection as active
            this.executionManager.activateConnection(connection);

            // Execute the target node
            await this.executeNode(connection.targetNodeId);
        }
    }

    /**
     * Build DOM operation from node properties
     */
    private buildDomOperation(node: Node): any {
        // This would extract operation details from node properties
        // and input connections
        // For simplicity, just returning a mock operation
        return {
            mode: 'create',
            tagName: 'div',
            innerHTML: 'Created by WebBlueprint',
            parentSelector: 'body',
            attributes: {
                id: `wb-${node.id}`,
                class: 'wb-element'
            },
            styles: {
                backgroundColor: '#333',
                color: 'white',
                padding: '10px',
                margin: '10px',
                borderRadius: '4px'
            }
        };
    }

    /**
     * Build DOM event operation from node properties
     */
    private buildDomEventOperation(node: Node): any {
        // Mock event operation
        return {
            selector: 'body',
            eventType: 'click',
            useCapture: false,
            preventDefault: false,
            stopPropagation: false,
            nodeId: node.id,
            executionId: this.executionManager.executionStore.currentExecutionId
        };
    }

    /**
     * Build storage operation from node properties
     */
    private buildStorageOperation(node: Node): any {
        // Mock storage operation
        return {
            operation: 'get',
            storageType: 'local',
            key: 'test-key',
            nodeId: node.id,
            executionId: this.executionManager.executionStore.currentExecutionId,
            timestamp: Date.now()
        };
    }
}
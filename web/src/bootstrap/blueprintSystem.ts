// File: web/src/bootstrap/blueprintSystem.ts
/**
 * Blueprint System Bootstrap
 *
 * This module initializes and wires together all the components of the
 * WebBlueprint visual programming system.
 */

import { useNodeRegistryStore } from '../stores/nodeRegistry';
import { useWebSocketStore } from '../stores/websocket';
import { ExecutionService, ExecutionMode } from '../services/executionService';
import { ExecutionManager } from '../utils/executionManager';
import { domHandler } from '../utils/domHandler';
import { storageHandler } from '../utils/storageHandler';


export let executionManager: ExecutionManager;
export let executionService: ExecutionService;

/**
 * Initialize the WebBlueprint system
 */
export async function initializeBlueprintSystem() {
    console.log('Initializing WebBlueprint system...');

    executionManager = new ExecutionManager();
    executionService = new ExecutionService(executionManager);

    // Initialize stores
    const nodeRegistryStore = useNodeRegistryStore();
    const websocketStore = useWebSocketStore();

    try {
        // Connect to WebSocket for real-time updates
        websocketStore.connect();

        // Load node types
        await nodeRegistryStore.fetchNodeTypes();
        console.log(`Loaded ${Object.keys(nodeRegistryStore.nodeTypes).length} node types`);

        // Set execution mode based on environment or user preference
        // For now, default to DIRECT mode
        const preferredMode = localStorage.getItem('execution-mode') as ExecutionMode;
        if (preferredMode && Object.values(ExecutionMode).includes(preferredMode)) {
            executionService.setExecutionMode(preferredMode);
        }

        // Check for client-side capabilities
        checkClientCapabilities();

        console.log('WebBlueprint system initialization complete');
        return true;
    } catch (error) {
        console.error('Error initializing WebBlueprint system:', error);
        return false;
    }
}

/**
 * Check for client-side capabilities and adjust execution accordingly
 */
function checkClientCapabilities() {
    // Check for storage capability
    const localStorageAvailable = storageHandler.isStorageAvailable('localStorage');
    const sessionStorageAvailable = storageHandler.isStorageAvailable('sessionStorage');

    console.log(`Client capabilities: localStorage=${localStorageAvailable}, sessionStorage=${sessionStorageAvailable}`);

    // If key browser features aren't available, switch to actor mode
    if (!localStorageAvailable || !sessionStorageAvailable) {
        console.warn('Some browser features are not available, switching to actor-based execution mode');
        executionService.setExecutionMode(ExecutionMode.ACTOR);
    }
}

/**
 * Set the execution mode
 */
export function setExecutionMode(mode: ExecutionMode) {
    executionService.setExecutionMode(mode);
    localStorage.setItem('execution-mode', mode);
}

/**
 * Get current execution mode
 */
export function getExecutionMode(): ExecutionMode {
    return executionService.getExecutionMode();
}

/**
 * Execute a blueprint
 */
export async function executeBlueprint(blueprintId: string, initialData?: Record<string, any>) {
    return executionService.executeBlueprint(blueprintId, initialData);
}

/**
 * Clean up the WebBlueprint system
 * Call this when the app is being unmounted
 */
export function cleanupBlueprintSystem() {
    // Disconnect WebSocket
    const websocketStore = useWebSocketStore();
    websocketStore.disconnect();

    // Clean up DOM handler (remove event listeners, etc.)
    domHandler.cleanup();

    // Clear any active visualizations
    executionManager.clearActiveStates();

    console.log('WebBlueprint system cleanup complete');
}
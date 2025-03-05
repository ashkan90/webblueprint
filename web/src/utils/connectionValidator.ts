import type { NodeTypeDefinition, PinDefinition } from '../types/nodes';
import { useBlueprintStore } from '../stores/blueprint';
import { useNodeRegistryStore } from '../stores/nodeRegistry';

interface ValidationResult {
    valid: boolean;
    reason?: string;
    suggestedFix?: string;
}

/**
 * Validates a potential connection between two pins and provides detailed feedback
 */
export function validateConnection(
    sourceNodeId: string,
    sourcePinId: string,
    targetNodeId: string,
    targetPinId: string
): ValidationResult {
    const blueprintStore = useBlueprintStore();
    const nodeRegistryStore = useNodeRegistryStore();

    // Get the source and target nodes
    const sourceNode = blueprintStore.getNodeById(sourceNodeId);
    const targetNode = blueprintStore.getNodeById(targetNodeId);

    if (!sourceNode || !targetNode) {
        return {
            valid: false,
            reason: 'Source or target node not found',
        };
    }

    // Get the node type definitions
    const sourceNodeType = nodeRegistryStore.getNodeTypeById(sourceNode.type);
    const targetNodeType = nodeRegistryStore.getNodeTypeById(targetNode.type);

    if (!sourceNodeType || !targetNodeType) {
        return {
            valid: false,
            reason: 'Source or target node type not found',
        };
    }

    // Find the pin definitions
    const sourcePin = sourceNodeType.outputs.find(pin => pin.id === sourcePinId);
    const targetPin = targetNodeType.inputs.find(pin => pin.id === targetPinId);

    if (!sourcePin || !targetPin) {
        return {
            valid: false,
            reason: 'Source or target pin not found',
        };
    }

    // Check if connecting an output to an input
    if (sourceNodeType.inputs.some(pin => pin.id === sourcePinId) ||
        targetNodeType.outputs.some(pin => pin.id === targetPinId)) {
        return {
            valid: false,
            reason: 'Cannot connect input to input or output to output',
            suggestedFix: 'Connect from an output pin to an input pin'
        };
    }

    // Check if connecting execution to execution or data to data
    const isSourceExecution = sourcePin.type.id === 'execution';
    const isTargetExecution = targetPin.type.id === 'execution';

    if (isSourceExecution !== isTargetExecution) {
        return {
            valid: false,
            reason: `Cannot connect ${isSourceExecution ? 'execution' : 'data'} output to ${isTargetExecution ? 'execution' : 'data'} input`,
            suggestedFix: `Connect ${isSourceExecution ? 'execution' : 'data'} pins to ${isSourceExecution ? 'execution' : 'data'} pins`
        };
    }

    // For execution pins, the connection is valid if we made it this far
    if (isSourceExecution) {
        return { valid: true };
    }

    // For data pins, check type compatibility
    if (!nodeRegistryStore.arePinsCompatible(sourcePin.type.id, targetPin.type.id)) {
        // Suggest possible conversions
        let suggestedFix = '';

        if (targetPin.type.id === 'string') {
            suggestedFix = 'Add a Type Conversion node to convert the value to a string';
        } else if (targetPin.type.id === 'number' && sourcePin.type.id === 'string') {
            suggestedFix = 'Add a Type Conversion node to parse the string as a number';
        } else if (targetPin.type.id === 'boolean') {
            suggestedFix = 'Add a Type Conversion node to convert the value to a boolean';
        } else if (targetPin.type.id === 'object' && sourcePin.type.id === 'string') {
            suggestedFix = 'Add a JSON Processor node to parse the string as an object';
        } else if (targetPin.type.id === 'array' && sourcePin.type.id === 'string') {
            suggestedFix = 'Add an Array Operations node to split the string into an array';
        }

        return {
            valid: false,
            reason: `Type mismatch: Cannot connect ${sourcePin.type.name} to ${targetPin.type.name}`,
            suggestedFix
        };
    }

    // If already connected, don't allow duplicate
    const existingConnections = blueprintStore.connections.filter(
        conn => conn.targetNodeId === targetNodeId && conn.targetPinId === targetPinId
    );

    if (existingConnections.length > 0) {
        return {
            valid: false,
            reason: 'This input pin already has a connection',
            suggestedFix: 'Delete the existing connection first'
        };
    }

    // If we made it here, the connection is valid
    return { valid: true };
}

/**
 * Checks if two pins can be connected
 */
export function canConnectPins(sourcePin: PinDefinition, targetPin: PinDefinition): boolean {
    // Check if connecting execution to execution or data to data
    const isSourceExecution = sourcePin.type.id === 'execution';
    const isTargetExecution = targetPin.type.id === 'execution';

    if (isSourceExecution !== isTargetExecution) {
        return false;
    }

    // For execution pins, always allow connection
    if (isSourceExecution) {
        return true;
    }

    // For data pins, check type compatibility
    const nodeRegistryStore = useNodeRegistryStore();
    return nodeRegistryStore.arePinsCompatible(sourcePin.type.id, targetPin.type.id);
}
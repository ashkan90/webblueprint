// Represents a pin type
export interface PinType {
    id: string
    name: string
    description: string
}

// Represents a pin on a node
export interface PinDefinition {
    id: string;
    name: string;
    description: string;
    type: PinTypeDefinition;
    optional?: boolean;
    default?: any;
}

// Type definitions for pin types
export interface PinTypeDefinition {
    id: string;
    name: string;
    description: string;
    color?: string; // Optional color for the pin
}

// Node property definition
export interface NodePropertyDefinition {
    name: string;
    displayName: string;
    description: string;
    type: string; // 'string', 'number', 'boolean', 'select', etc.
    value?: any;
    options?: string[]; // For select property type
}

// Represents a node type definition
export interface NodeTypeDefinition {
    typeId: string;
    name: string;
    description: string;
    category: string;
    version: string;
    inputs: PinDefinition[];
    outputs: PinDefinition[];
    properties?: NodePropertyDefinition[];
    icon?: string; // Optional icon for the node
}

// Built-in pin types
export const PinTypes = {
    EXECUTION: 'execution',
    STRING: 'string',
    NUMBER: 'number',
    BOOLEAN: 'boolean',
    OBJECT: 'object',
    ARRAY: 'array',
    ANY: 'any'
};

// Pin type colors
export const PinTypeColors = {
    [PinTypes.EXECUTION]: '#ffffff',
    [PinTypes.STRING]: '#f0883e',
    [PinTypes.NUMBER]: '#6ed69a',
    [PinTypes.BOOLEAN]: '#dc5050',
    [PinTypes.OBJECT]: '#8ab4f8',
    [PinTypes.ARRAY]: '#bb86fc',
    [PinTypes.ANY]: '#aaaaaa'
};

// Helper function to get a color for a pin type
export function getPinTypeColor(typeId: string): string {
    return PinTypeColors[typeId] || PinTypeColors[PinTypes.ANY];
}

// Helper function to check if two pin types are compatible
export function arePinTypesCompatible(sourceType: string, targetType: string): boolean {
    // Execution pins can only connect to execution pins
    if (sourceType === PinTypes.EXECUTION) {
        return targetType === PinTypes.EXECUTION;
    }

    if (targetType === PinTypes.EXECUTION) {
        return sourceType === PinTypes.EXECUTION;
    }

    // ANY can connect to anything (except execution)
    if (sourceType === PinTypes.ANY || targetType === PinTypes.ANY) {
        return true;
    }

    // Same types can connect
    if (sourceType === targetType) {
        return true;
    }

    // Special conversions
    // STRING can accept any type (automatic conversion)
    if (targetType === PinTypes.STRING) {
        return true;
    }

    // NUMBER can accept STRING if it contains a valid number
    if (targetType === PinTypes.NUMBER && sourceType === PinTypes.STRING) {
        return true; // Validation happens at runtime
    }

    // BOOLEAN can accept NUMBER (0 = false, anything else = true)
    return targetType === PinTypes.BOOLEAN && sourceType === PinTypes.NUMBER;
}

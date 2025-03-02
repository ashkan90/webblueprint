// Represents a pin type
export interface PinType {
    id: string
    name: string
    description: string
}

// Represents a pin on a node
export interface PinDefinition {
    id: string
    name: string
    description: string
    type: PinType
    optional?: boolean
    default?: any
}

// Represents a node type definition
export interface NodeTypeDefinition {
    typeId: string
    name: string
    description: string
    category: string
    version: string
    inputs: PinDefinition[]
    outputs: PinDefinition[]
}
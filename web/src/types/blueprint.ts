// Represents a position on the canvas
import {NodeTypeDefinition} from "./nodes";
import {EventBinding, EventDefinition} from "../services/eventService";

export interface Position {
    x: number
    y: number
}

// Properties for a node
export interface NodeProperty {
    name: string
    displayName: string
    value: any
}

// Represents a node in a blueprint
export interface Node {
    id: string
    type: string
    position: Position
    properties: NodeProperty[]
    data?: Record<string, any>
}

// Represents a connection between nodes
export interface Connection {
    id: string
    sourceNodeId: string
    sourcePinId: string
    targetNodeId: string
    targetPinId: string
    connectionType: 'execution' | 'data'
}

// Represents a blueprint variable
export interface Variable {
    id: string
    name: string
    type: string
    value: any
}

export interface Function {
    id: string
    name: string
    description: string
    nodeType: NodeTypeDefinition
    nodes: Node[]
    connections: Connection[]
    variables: Variable[]
    metadata: Record<string, string>
}

// Represents a complete blueprint
export interface Blueprint {
    id: string
    name: string
    description: string
    version: string
    nodes: Node[]
    functions: Function[]
    connections: Connection[]
    variables: Variable[]
    events: EventDefinition[]
    eventBindings: EventBinding[]
    metadata: Record<string, string>
}

// Represents a connection between nodes
export interface Connection {
    id: string
    sourceNodeId: string
    sourcePinId: string
    targetNodeId: string
    targetPinId: string
    connectionType: 'execution' | 'data'
    data?: Record<string, any>  // Additional metadata for the connection
}

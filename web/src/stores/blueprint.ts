import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Node, Connection, Blueprint, Position, NodeProperty } from '../types/blueprint'
import { v4 as uuid } from 'uuid'

export const useBlueprintStore = defineStore('blueprint', () => {
    // State
    const blueprint = ref<Blueprint>({
        id: '',
        name: '',
        description: '',
        version: '1.0',
        nodes: [],
        connections: [],
        variables: [],
        metadata: {}
    })

    const isLoading = ref(false)
    const error = ref<string | null>(null)

    // Getters
    const nodes = computed(() => blueprint.value.nodes)
    const connections = computed(() => blueprint.value.connections)

    const getNodeById = computed(() => (id: string) => {
        return blueprint.value.nodes.find(node => node.id === id)
    })

    const getNodeConnections = computed(() => (nodeId: string) => {
        return blueprint.value.connections.filter(
            conn => conn.sourceNodeId === nodeId || conn.targetNodeId === nodeId
        )
    })

    const getNodeInputConnections = computed(() => (nodeId: string) => {
        return blueprint.value.connections.filter(conn => conn.targetNodeId === nodeId)
    })

    const getNodeOutputConnections = computed(() => (nodeId: string) => {
        return blueprint.value.connections.filter(conn => conn.sourceNodeId === nodeId)
    })

    const isNodePinConnected = computed(() => (nodeId: string, pinId: string, direction: 'input' | 'output') => {
        if (direction === 'input') {
            return blueprint.value.connections.some(
                conn => conn.targetNodeId === nodeId && conn.targetPinId === pinId
            )
        } else {
            return blueprint.value.connections.some(
                conn => conn.sourceNodeId === nodeId && conn.sourcePinId === pinId
            )
        }
    })

    const findEntryPoints = computed(() => {
        const execInputs = new Set<string>()
        const execOutputs = new Set<string>()

        // Find all nodes with execution connections
        for (const conn of blueprint.value.connections) {
            if (conn.connectionType === 'execution') {
                execOutputs.add(conn.sourceNodeId)
                execInputs.add(conn.targetNodeId)
            }
        }

        // Find nodes with execution outputs but no execution inputs
        const entryPoints = blueprint.value.nodes
            .filter(node => execOutputs.has(node.id) && !execInputs.has(node.id))
            .map(node => node.id)

        // Also include special entry point nodes like DOM events
        blueprint.value.nodes.forEach(node => {
            if (node.type === 'dom-event' && !entryPoints.includes(node.id)) {
                entryPoints.push(node.id)
            }
        })

        return entryPoints
    })

    // Actions
    function createBlueprint(name: string, description: string = '') {
        blueprint.value = {
            id: uuid(),
            name,
            description,
            version: '1.0',
            nodes: [],
            connections: [],
            variables: [],
            metadata: {}
        }
    }

    async function loadBlueprint(id: string) {
        isLoading.value = true
        error.value = null

        try {
            const response = await fetch(`/api/blueprints/${id}`)
            if (!response.ok) {
                throw new Error(`Failed to load blueprint: ${response.statusText}`)
            }

            const data = await response.json()
            blueprint.value = data
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err)
            console.error('Error loading blueprint:', err)
        } finally {
            isLoading.value = false
        }
    }

    async function saveBlueprint() {
        isLoading.value = true
        error.value = null

        try {
            const method = blueprint.value.id ? 'PUT' : 'POST'
            const url = blueprint.value.id
                ? `/api/blueprints/${blueprint.value.id}`
                : '/api/blueprints'

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(blueprint.value)
            })

            if (!response.ok) {
                throw new Error(`Failed to save blueprint: ${response.statusText}`)
            }

            const data = await response.json()
            blueprint.value = data
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err)
            console.error('Error saving blueprint:', err)
            throw err
        } finally {
            isLoading.value = false
        }
    }

    function addNode(node: Node) {
        blueprint.value.nodes.push(node)
    }

    function updateNode(id: string, updates: Partial<Node>) {
        const index = blueprint.value.nodes.findIndex(node => node.id === id)
        if (index !== -1) {
            blueprint.value.nodes[index] = { ...blueprint.value.nodes[index], ...updates }
        }
    }

    function updateNodePosition(id: string, position: Position) {
        const node = blueprint.value.nodes.find(node => node.id === id)
        if (node) {
            node.position = position
        }
    }

    function updateNodeProperty(nodeId: string, propertyName: string, value: any) {
        const node = blueprint.value.nodes.find(node => node.id === nodeId)
        if (node) {
            // Find the property
            const propIndex = node.properties.findIndex(p => p.name === propertyName)
            if (propIndex !== -1) {
                // Update existing property
                node.properties[propIndex].value = value
            } else {
                // Add new property
                node.properties.push({ name: propertyName, value })
            }
        }
    }

    function removeNode(id: string) {
        // Remove connections first
        blueprint.value.connections = blueprint.value.connections.filter(
            conn => conn.sourceNodeId !== id && conn.targetNodeId !== id
        )

        // Remove node
        blueprint.value.nodes = blueprint.value.nodes.filter(node => node.id !== id)
    }

    function addConnection(connection: Connection) {
        // Ensure the connection has an ID
        if (!connection.id) {
            connection.id = uuid()
        }

        // Check if a similar connection already exists
        const exists = blueprint.value.connections.some(
            conn =>
                conn.sourceNodeId === connection.sourceNodeId &&
                conn.sourcePinId === connection.sourcePinId &&
                conn.targetNodeId === connection.targetNodeId &&
                conn.targetPinId === connection.targetPinId
        )

        if (!exists) {
            blueprint.value.connections.push(connection)
        }
    }

    function removeConnection(id: string) {
        blueprint.value.connections = blueprint.value.connections.filter(
            conn => conn.id !== id
        )
    }

    return {
        blueprint,
        isLoading,
        error,
        nodes,
        connections,
        getNodeById,
        getNodeConnections,
        getNodeInputConnections,
        getNodeOutputConnections,
        isNodePinConnected,
        findEntryPoints,
        createBlueprint,
        loadBlueprint,
        saveBlueprint,
        addNode,
        updateNode,
        updateNodePosition,
        updateNodeProperty,
        removeNode,
        addConnection,
        removeConnection
    }
})
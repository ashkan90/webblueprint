import {defineStore} from 'pinia'
import {computed, ref} from 'vue'
import type {Blueprint, Function, Connection, Node, Position, Variable} from '../types/blueprint'
import {v4 as uuid} from 'uuid'

export const useBlueprintStore = defineStore('blueprint', () => {
    // State
    const blueprint = ref<Blueprint>({
        id: '',
        name: '',
        description: '',
        version: '1.0',
        functions: [],
        nodes: [],
        connections: [],
        variables: [],
        metadata: {}
    })

    const isLoading = ref(false)
    const error = ref<string | null>(null)

    const currentEditingFunction = ref<string | null>(null)

    // Getters
    const nodes = computed(() => blueprint.value.nodes || [])
    const connections = computed(() => blueprint.value.connections || [])
    const variables = computed(() => {
        if (!blueprint.value.variables) {
            blueprint.value.variables = [];
        }
        return blueprint.value.variables;
    })
    const functions = computed(() => blueprint.value.functions || [])

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

    const isFunctionEditing = computed(() => currentEditingFunction.value !== null)

    // Actions
    function createBlueprint(name: string, description: string = '') {
        blueprint.value = {
            id: uuid(),
            name,
            description,
            version: '1.0',
            functions: [],
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

            blueprint.value = await response.json()
            
            // Ensure the blueprint structure is complete
            if (!blueprint.value.variables) {
                console.log('Variables array not found in blueprint response, creating empty array');
                blueprint.value.variables = [];
            }
            
            if (!blueprint.value.functions) {
                blueprint.value.functions = [];
            }
            
            if (!blueprint.value.connections) {
                blueprint.value.connections = [];
            }
            
            if (!blueprint.value.nodes) {
                blueprint.value.nodes = [];
            }
            
            console.log('Blueprint structure after loading:', blueprint.value);
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

            blueprint.value = await response.json()
            
            // Ensure the blueprint structure is complete after saving
            if (!blueprint.value.variables) {
                console.log('Variables array not found in blueprint response after saving, creating empty array');
                blueprint.value.variables = [];
            }
            
            if (!blueprint.value.functions) {
                blueprint.value.functions = [];
            }
            
            if (!blueprint.value.connections) {
                blueprint.value.connections = [];
            }
            
            if (!blueprint.value.nodes) {
                blueprint.value.nodes = [];
            }
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err)
            console.error('Error saving blueprint:', err)
            throw err
        } finally {
            isLoading.value = false
        }
    }

    function addNode(node: Node) {
        // Make sure we're creating a deep copy to avoid shared references
        const nodeCopy = JSON.parse(JSON.stringify(node));

        // Ensure properties is initialized
        if (!nodeCopy.properties) {
            nodeCopy.properties = [];
        }

        // Add the node
        blueprint.value.nodes.push(nodeCopy);
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

            // If this is a pin default value (input_*), add it to the node's data for easy access
            if (propertyName.startsWith('input_')) {
                if (!node.data) {
                    node.data = {}
                }
                if (!node.data.defaults) {
                    node.data.defaults = {}
                }
                const pinId = propertyName.substring(6) // Remove "input_" prefix
                node.data.defaults[pinId] = value
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

    async function addVariable(variable: Variable) {
        // [sink] try to save blueprint before saving variable
        await saveBlueprint()

        try {
            const url = `/api/blueprints/${blueprint.value.id}/variable`
            const variableResponse = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(variable),
            })

            if (!variableResponse.ok) {
                throw new Error('Failed to add variable into blueprint')
            }
            // Ensure variables array exists
            if (!blueprint.value.variables) {
                console.log('Creating variables array before adding variable');
                blueprint.value.variables = [];
            }

            // Add the variable
            console.log('Adding variable to blueprint:', variable);
            blueprint.value.variables.push(await variableResponse.json());
        } catch (e) {
            console.log('Something went wrong while adding variable to blueprint', e)
        }
    }

    function addFunction(fn: Function) {
        // Ensure function has execution pins
        let hasExecutionInput = false;
        let hasExecutionOutput = false;
        
        for (const input of fn.nodeType.inputs) {
            if (input.type?.id === 'execution') {
                hasExecutionInput = true;
                break;
            }
        }
        
        for (const output of fn.nodeType.outputs) {
            if (output.type?.id === 'execution') {
                hasExecutionOutput = true;
                break;
            }
        }
        
        // Add default execution input if needed
        if (!hasExecutionInput) {
            fn.nodeType.inputs.push({
                id: 'exec',
                name: 'Execute',
                description: 'Execution continues',
                type: {
                    id: 'execution',
                    name: 'Execution',
                    description: 'Controls execution flow',
                },
                optional: false
            });
        }
        
        // Add default execution output if needed
        if (!hasExecutionOutput) {
            fn.nodeType.outputs.push({
                id: 'then',
                name: 'Then',
                description: 'Execution continues',
                type: {
                    id: 'execution',
                    name: 'Execution',
                    description: 'Controls execution flow',
                },
                optional: false
            });
        }
        
        blueprint.value.functions.push(fn)
    }

    function addNodeToFunction(fnID: string, node: Node) {
        const fnIdx = blueprint.value.functions.findIndex((fn: Function) => fn.id === fnID)
        blueprint.value.functions[fnIdx].nodes.push(node)
    }

    return {
        blueprint,
        isLoading,
        error,
        currentEditingFunction,
        nodes,
        connections,
        getNodeById,
        getNodeConnections,
        getNodeInputConnections,
        getNodeOutputConnections,
        isNodePinConnected,
        isFunctionEditing,
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
        removeConnection,
        variables,
        addVariable,
        functions,
        addFunction,
        addNodeToFunction
    }
})
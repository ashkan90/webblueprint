import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWebSocketStore, WebSocketEvents } from './websocket'
import type { NodeTypeDefinition, PinDefinition } from '../types/nodes'
import { arePinTypesCompatible } from '../types/nodes'


export const useNodeRegistryStore = defineStore('nodeRegistry', () => {
    // State
    const nodeTypes = ref<Record<string, NodeTypeDefinition>>({})
    const isLoading = ref(false)
    const error = ref<string | null>(null)

    // WebSocket store
    const websocketStore = useWebSocketStore()

    // Getters
    const getNodeTypeById = computed(() =>
        (id: string) => nodeTypes.value[id] || null
    )

    const nodeTypesByCategory = computed(() => {
        const categories: Record<string, NodeTypeDefinition[]> = {}

        Object.values(nodeTypes.value).forEach(nodeType => {
            if (!categories[nodeType.category]) {
                categories[nodeType.category] = []
            }
            categories[nodeType.category].push(nodeType)
        })

        return categories
    })

    const categories = computed(() => {
        return Object.keys(nodeTypesByCategory.value).sort()
    })

    const getExecutionInputPins = computed(() => (typeId: string) => {
        const nodeType = nodeTypes.value[typeId]
        if (!nodeType) return []

        return nodeType.inputs.filter(pin => pin.type.id === 'execution')
    })

    const getExecutionOutputPins = computed(() => (typeId: string) => {
        const nodeType = nodeTypes.value[typeId]
        if (!nodeType) return []

        return nodeType.outputs.filter(pin => pin.type.id === 'execution')
    })

    const getDataInputPins = computed(() => (typeId: string) => {
        const nodeType = nodeTypes.value[typeId]
        if (!nodeType) return []

        return nodeType.inputs.filter(pin => pin.type.id !== 'execution')
    })

    const getDataOutputPins = computed(() => (typeId: string) => {
        const nodeType = nodeTypes.value[typeId]
        if (!nodeType) return []

        return nodeType.outputs.filter(pin => pin.type.id !== 'execution')
    })

    // Actions
    function registerNodeType(nodeType: NodeTypeDefinition) {
        nodeTypes.value[nodeType.typeId] = nodeType
    }

    async function fetchNodeTypes() {
        isLoading.value = true
        error.value = null

        try {
            const response = await fetch('/api/nodes')
            if (!response.ok) {
                throw new Error('Failed to fetch node types')
            }

            const data = await response.json()
            data.forEach((nodeType: NodeTypeDefinition) => {
                registerNodeType(nodeType)
            })
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err)
            console.error('Error fetching node types:', err)
        } finally {
            isLoading.value = false
        }
    }

    // Initialize WebSocket listener for node introductions
    function setupWebSocketListeners() {
        websocketStore.on(WebSocketEvents.NODE_INTRO, (data: NodeTypeDefinition) => {
            registerNodeType(data)
        })
    }

    // Check if pins are compatible for connection
    function arePinsCompatible(sourcePinType: string, targetPinType: string): boolean {
        return arePinTypesCompatible(sourcePinType, targetPinType);
    }

    // Call setup after store is created
    setupWebSocketListeners()

    return {
        nodeTypes,
        isLoading,
        error,
        getNodeTypeById,
        nodeTypesByCategory,
        categories,
        getExecutionInputPins,
        getExecutionOutputPins,
        getDataInputPins,
        getDataOutputPins,
        registerNodeType,
        fetchNodeTypes,
        arePinsCompatible
    }
})
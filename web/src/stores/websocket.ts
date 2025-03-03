import { defineStore } from 'pinia'
import { ref } from 'vue'

// Enhanced WebSocket message types with more precise typing
export const WebSocketEvents = {
    NODE_INTRO: 'node.intro',
    NODE_START: 'node.start',
    NODE_COMPLETE: 'node.complete',
    NODE_ERROR: 'node.error',
    DATA_FLOW: 'data.flow',
    DEBUG_DATA: 'debug.data',
    EXEC_START: 'execution.start',
    EXEC_END: 'execution.end',
    EXEC_STATUS: 'execution.status',
    RESULT: 'result',
    LOG: 'log'
}

export type MessageHandler = (data: unknown) => void

export interface WebSocketMessage {
    type: string
    payload: unknown
}

export const useWebSocketStore = defineStore('websocket', () => {
    // State
    const socket = ref<WebSocket | null>(null)
    const connectionStatus = ref<'connected' | 'connecting' | 'disconnected'>('disconnected')
    const reconnectInterval = ref(2000)
    const reconnectAttempts = ref(0)
    const maxReconnectAttempts = ref(5)
    const handlers = ref<Map<string, MessageHandler[]>>(new Map())

    // Actions
    function connect() {
        if (socket.value && socket.value.readyState === WebSocket.OPEN) {
            return
        }

        connectionStatus.value = 'connecting'

        // Determine WebSocket URL based on current location
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host
        const wsUrl = `${protocol}//${host}/ws`

        console.log(`Connecting to WebSocket at ${wsUrl}`)

        try {
            socket.value = new WebSocket(wsUrl)

            socket.value.onopen = () => {
                console.log('WebSocket connected')
                connectionStatus.value = 'connected'
                reconnectAttempts.value = 0
            }

            socket.value.onclose = (event) => {
                console.log('WebSocket disconnected', event)
                connectionStatus.value = 'disconnected'

                // Try to reconnect unless max attempts reached
                if (reconnectAttempts.value < maxReconnectAttempts.value) {
                    reconnectAttempts.value++
                    setTimeout(connect, reconnectInterval.value)
                }
            }

            socket.value.onerror = (error) => {
                console.error('WebSocket error', error)
            }

            socket.value.onmessage = (event) => {
                try {
                    if (event.data.includes('\n')) {
                        event.data.split('\n').forEach((eventData: string) => {
                            const message = JSON.parse(eventData) as WebSocketMessage
                            handleMessage(message)
                        })
                        return
                    }

                    const message = JSON.parse(event.data) as WebSocketMessage
                    handleMessage(message)
                } catch (error) {
                    console.error('Error parsing WebSocket message', error)
                }
            }
        } catch (error) {
            console.error('Error creating WebSocket', error)
            connectionStatus.value = 'disconnected'
        }
    }

    function disconnect() {
        if (socket.value) {
            socket.value.close()
            socket.value = null
        }
    }

    function on<T = unknown>(event: string, handler: (data: T) => void): () => void {
        if (!handlers.value.has(event)) {
            handlers.value.set(event, [])
        }

        const eventHandlers = handlers.value.get(event)
        if (eventHandlers) {
            eventHandlers.push(handler as MessageHandler)
        }

        // Return a function to remove this handler
        return () => {
            const currentHandlers = handlers.value.get(event) || []
            const index = currentHandlers.indexOf(handler as MessageHandler)
            if (index !== -1) {
                currentHandlers.splice(index, 1)
            }
        }
    }

    function send(type: string, payload: unknown): void {
        if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected')
            return
        }

        const message = { type, payload }
        socket.value.send(JSON.stringify(message))
    }

    function handleMessage(message: WebSocketMessage): void {
        // Call handlers for this message type
        const eventHandlers = handlers.value.get(message.type) || []
        eventHandlers.forEach(handler => handler(message.payload))

        // Also call "all" handlers
        const allHandlers = handlers.value.get('all') || []
        allHandlers.forEach(handler => handler(message))
    }

    return {
        socket,
        connectionStatus,
        connect,
        disconnect,
        on,
        send
    }
})
// File: web/src/stores/events.ts
import { defineStore } from 'pinia'
import { ref, computed, onMounted } from 'vue'
import { EventService, EventDefinition as APIEventDefinition, EventParameter } from '../services/eventService'

// Define event types
export interface EventDefinition {
  id: string
  name: string
  description: string
  type: 'entry' | 'dispatch'
  category: string
  parameters?: EventParameter[]
  blueprintId?: string
}

// Event store definition
export const useEventStore = defineStore('events', () => {
  // System events (fixed entry points)
  const systemEvents = ref<EventDefinition[]>([
    // Entry point events - similar to Unreal Engine
    {
      id: 'event-on-created',
      name: 'On Created',
      description: 'Called when the blueprint is created',
      type: 'entry',
      category: 'System Events',
    },
    {
      id: 'event-on-tick',
      name: 'On Tick',
      description: 'Called periodically during execution',
      type: 'entry',
      category: 'System Events',
      parameters: [
        {
          name: 'deltaTime',
          type: 'number',
          description: 'Time elapsed since the last tick',
          optional: false
        }
      ]
    },
    {
      id: 'event-on-input',
      name: 'On Input',
      description: 'Called when input is received',
      type: 'entry',
      category: 'System Events',
      parameters: [
        {
          name: 'inputName',
          type: 'string',
          description: 'Name of the input received',
          optional: false
        },
        {
          name: 'inputValue',
          type: 'any',
          description: 'Value of the input received',
          optional: false
        }
      ]
    },
  ])
  
  // Events from server (custom event dispatchers)
  const serverEvents = ref<EventDefinition[]>([])
  
  // Loading state
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  // Fetch events from server
  async function fetchEvents() {
    loading.value = true
    error.value = null
    
    try {
      const events = await EventService.fetchEvents()
      
      // Convert server events to our format
      serverEvents.value = events
        .filter(event => event.category !== 'System Events') // Filter out system events which are already in systemEvents
        .map(convertApiEventToStoreEvent)
    } catch (err) {
      console.error('Failed to fetch events:', err)
      error.value = 'Failed to load events from server'
    } finally {
      loading.value = false
    }
  }

  // Fetch blueprint-specific events
  async function fetchBlueprintEvents(blueprintId: string) {
    if (!blueprintId) return
    
    loading.value = true
    error.value = null
    
    try {
      const events = await EventService.fetchBlueprintEvents(blueprintId)
      
      // Add these events to our serverEvents, avoiding duplicates
      const existingIds = new Set(serverEvents.value.map(e => e.id))
      
      events.forEach(event => {
        if (!existingIds.has(event.id)) {
          serverEvents.value.push(convertApiEventToStoreEvent(event))
          existingIds.add(event.id)
        }
      })
    } catch (err) {
      console.error(`Failed to fetch blueprint events for ${blueprintId}:`, err)
      error.value = 'Failed to load blueprint events from server'
    } finally {
      loading.value = false
    }
  }
  
  // Create a new event dispatcher
  async function createEventDispatcher(name: string, description = '', blueprintId?: string): Promise<EventDefinition | null> {
    if (!name) return null
    
    try {
      // Prepare event data for API
      const eventData: Partial<APIEventDefinition> = {
        name,
        description,
        category: 'Custom Events',
        parameters: [],
      }
      
      // If we have a blueprint ID, associate with it
      if (blueprintId) {
        eventData.blueprintId = blueprintId
      }
      
      // Send to server
      const createdEvent = await EventService.createEvent(eventData)
      
      // Convert to our format
      const newEvent = convertApiEventToStoreEvent(createdEvent)
      
      // Add to local store
      serverEvents.value.push(newEvent)
      
      return newEvent
    } catch (err) {
      console.error('Failed to create event dispatcher:', err)
      throw err
    }
  }
  
  // Delete an event dispatcher
  async function deleteEventDispatcher(id: string) {
    try {
      await EventService.deleteEvent(id)
      
      // Remove from local store
      const index = serverEvents.value.findIndex(e => e.id === id)
      if (index !== -1) {
        serverEvents.value.splice(index, 1)
      }
    } catch (err) {
      console.error(`Failed to delete event dispatcher ${id}:`, err)
      throw err
    }
  }
  
  // Helper to convert API event format to our store format
  function convertApiEventToStoreEvent(apiEvent: APIEventDefinition): EventDefinition {
    return {
      id: apiEvent.id,
      name: apiEvent.name,
      description: apiEvent.description || '',
      type: 'dispatch', // All server events are dispatchers
      category: apiEvent.category || 'Custom Events',
      parameters: apiEvent.parameters || [],
      blueprintId: apiEvent.blueprintId
    }
  }
  
  // Getters
  const allEvents = computed(() => [...systemEvents.value, ...serverEvents.value])
  const entryPointEvents = computed(() => allEvents.value.filter(e => e.type === 'entry'))
  const dispatchEvents = computed(() => allEvents.value.filter(e => e.type === 'dispatch'))
  
  // Blueprint-specific events (useful for showing only relevant events)
  function getBlueprintEvents(blueprintId: string) {
    return serverEvents.value.filter(e => 
      e.blueprintId === blueprintId || // Events owned by this blueprint
      !e.blueprintId  // Global events
    )
  }
  
  // Load events on store initialization
  onMounted(async () => {
    await fetchEvents()
  })
  
  return {
    systemEvents,
    serverEvents,
    loading,
    error,
    allEvents,
    entryPointEvents,
    dispatchEvents,
    fetchEvents,
    fetchBlueprintEvents,
    createEventDispatcher,
    deleteEventDispatcher,
    getBlueprintEvents
  }
})

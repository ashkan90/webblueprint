// File: web/src/components/editor/event-panel-integration.js

/**
 * This code should be integrated into BlueprintLeftPanel.vue to connect
 * to the event store and properly handle event dispatchers
 */

import { useEventStore } from '../../stores/events'

// Add to script setup section:
const eventStore = useEventStore()

// Replace the existing filteredEventDispatchers computed property:
const filteredEventDispatchers = computed(() => {
  // Get dispatch events from the event store
  const events = eventStore.dispatchEvents
  
  if (!searchQuery.value) return events

  return events.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Replace the filteredEntryPointEvents computed property:
const filteredEntryPointEvents = computed(() => {
  // Get entry point events from the event store
  const events = eventStore.entryPointEvents
  
  if (!searchQuery.value) return events

  return events.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Replace the createEventDispatcher function:
async function createEventDispatcher() {
  if (!newEventDispatcher.value.name) {
    return
  }
  
  try {
    // Create the event dispatcher via the store
    await eventStore.createEventDispatcher(
      newEventDispatcher.value.name,
      newEventDispatcher.value.description
    )
    
    showCreateEventDispatcherModal.value = false

    // Reset form
    newEventDispatcher.value = {
      name: '',
      description: ''
    }
  } catch (error) {
    console.error('Failed to create event dispatcher:', error)
    // Show an error message to the user
    // You might want to add a toast or notification system for this
  }
}

// Update onEventDragStart function to handle different event types:
function onEventDragStart(event: DragEvent, eventDefinition: any) {
  if (!event.dataTransfer) return

  // Create appropriate node based on event type
  const nodeType = eventDefinition.type === 'entry' 
    ? eventDefinition.id  // Use ID for entry point nodes (e.g., event-on-created)
    : 'event-dispatcher'  // Use generic dispatcher for dispatch events
  
  // Create a node representation of this event
  const nodeData = {
    id: uuid(),
    type: nodeType,
    position: { x: 0, y: 0 },
    properties: [
      { name: 'eventId', value: eventDefinition.id },
      { name: 'eventName', value: eventDefinition.name }
    ]
  }

  // For dispatch events, we need to add more properties
  if (eventDefinition.type === 'dispatch') {
    // Add any parameters as properties
    if (eventDefinition.parameters && eventDefinition.parameters.length > 0) {
      nodeData.properties.push({ 
        name: 'dynamicParameters',
        value: true
      })
    }
  }

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

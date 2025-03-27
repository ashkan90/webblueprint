// Function to enhance the EVENT DISPATCHERS section in BlueprintLeftPanel.vue

// First, add the import for the event store at the top of the script section:
import { useEventStore } from '../../stores/events'

// Then, add this to the script section to fetch events:
const eventStore = useEventStore()

// Replace the filteredEventDispatchers computed property with this:
const filteredEventDispatchers = computed(() => {
  // Get events from the event store
  const events = eventStore.dispatchEvents
  
  if (!searchQuery.value) return events

  return events.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Add a new computed property for entry point events:
const filteredEntryPointEvents = computed(() => {
  // Get entry point events from the event store
  const events = eventStore.entryPointEvents
  
  if (!searchQuery.value) return events

  return events.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Modify the onEventDragStart function to handle both types of events:
function onEventDragStart(event: DragEvent, eventDefinition: any) {
  if (!event.dataTransfer) return

  // Create appropriate node type based on event type
  const nodeType = eventDefinition.type === 'entry' 
    ? eventDefinition.id  // Use event ID for entry point nodes
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

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
  event.dataTransfer.effectAllowed = 'copy'
}

// Function to create an event dispatcher:
function createEventDispatcher() {
  // Create a new custom event
  const eventId = `custom.${newEventDispatcher.value.name.toLowerCase().replace(/\s+/g, '-')}`
  
  const newEvent = {
    id: eventId,
    name: newEventDispatcher.value.name,
    description: newEventDispatcher.value.description,
    type: 'dispatch' as const,
    category: 'Custom Events',
    parameters: []
  }
  
  // Add the event to the store
  eventStore.addEvent(newEvent)
  
  showCreateEventDispatcherModal.value = false

  // Reset form
  newEventDispatcher.value = {
    name: '',
    description: ''
  }
}

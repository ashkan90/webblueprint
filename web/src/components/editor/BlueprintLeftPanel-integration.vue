// File: web/src/components/editor/BlueprintLeftPanel.vue 
// Note: This is only the relevant code for integration - paste these sections into the existing file

// Import section - add this to the imports
import { useEventStore } from '../../stores/events'

// Add this to the script setup section
const eventStore = useEventStore()

// Replace the existing filteredEventDispatchers computed property with this
const filteredEventDispatchers = computed(() => {
  // Get dispatch events from the event store
  const events = eventStore.dispatchEvents || []
  
  if (!searchQuery.value) return events

  return events.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Add this computed property
const filteredEntryPointEvents = computed(() => {
  // Get entry point events from the event store
  const events = eventStore.entryPointEvents || []
  
  if (!searchQuery.value) return events

  return events.filter(e =>
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Replace the existing createEventDispatcher function with this
async function createEventDispatcher() {
  if (!newEventDispatcher.value.name) {
    return
  }
  
  try {
    // Create the event dispatcher via the API
    const response = await fetch('/api/events', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: newEventDispatcher.value.name,
        description: newEventDispatcher.value.description,
        category: 'Custom Events'
      }),
    });
    
    if (!response.ok) {
      throw new Error('Failed to create event dispatcher');
    }
    
    // Refresh the event list from the server
    await eventStore.fetchEvents();
    
    showCreateEventDispatcherModal.value = false;

    // Reset form
    newEventDispatcher.value = {
      name: '',
      description: ''
    };
    
    console.log('Event dispatcher created successfully');
  } catch (error) {
    console.error('Failed to create event dispatcher:', error);
    // You might want to show an error message to the user
  }
}

// Update the onEventDragStart function to handle both types of events
function onEventDragStart(event, eventDispatcher) {
  if (!event.dataTransfer) return;

  // Create a node representation of this event
  const nodeData = {
    id: uuid(),
    type: 'event-dispatcher',  // Using the improved dispatcher node
    position: { x: 0, y: 0 },
    properties: [
      { name: 'eventId', value: eventDispatcher.id },
      { name: 'eventName', value: eventDispatcher.name }
    ]
  };

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData));
  event.dataTransfer.effectAllowed = 'copy';
}

// Add this code in the onMounted hook or a separate function that runs when the component is loaded
onMounted(() => {
  // Fetch events from the server
  eventStore.fetchEvents();
  
  // Also add entry points section to expandedSections if it's not there
  if (expandedSections.value.entryPointEvents === undefined) {
    expandedSections.value.entryPointEvents = true;
  }
});

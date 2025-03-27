// First, we need to add a real implementation for creating event dispatchers
// This code should be added directly to BlueprintLeftPanel.vue

/**
 * Creates an event dispatcher by calling the server API
 */
async function createEventDispatcher() {
  try {
    // Validate input
    if (!newEventDispatcher.value.name) {
      alert("Event name is required");
      return;
    }
    
    // Prepare request
    const requestData = {
      name: newEventDispatcher.value.name,
      description: newEventDispatcher.value.description || "",
      category: "Custom Events"
    };
    
    // Call API to create event dispatcher
    const response = await fetch('/api/events', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(requestData)
    });
    
    if (!response.ok) {
      throw new Error(`Failed to create event: ${response.statusText}`);
    }
    
    // Parse response
    const createdEvent = await response.json();
    console.log('Created event dispatcher:', createdEvent);
    
    // Add to our list of event dispatchers
    const eventDispatchers = [
      ...filteredEventDispatchers.value,
      {
        id: createdEvent.id,
        name: createdEvent.name,
        description: createdEvent.description
      }
    ];
    
    // Fetch all events to refresh the list
    fetchEventDispatchers();
    
    // Close modal and reset form
    showCreateEventDispatcherModal.value = false;
    newEventDispatcher.value = {
      name: '',
      description: ''
    };
    
  } catch (error) {
    console.error('Error creating event dispatcher:', error);
    alert(`Failed to create event dispatcher: ${error.message}`);
  }
}

/**
 * Fetches event dispatchers from the server
 */
async function fetchEventDispatchers() {
  try {
    const response = await fetch('/api/events');
    if (!response.ok) {
      throw new Error(`Failed to fetch events: ${response.statusText}`);
    }
    
    const events = await response.json();
    console.log('Fetched events:', events);
    
    // Filter to only custom events and update the model
    // This requires modifying the filteredEventDispatchers computed property to use this data
    customEvents.value = events.filter(e => e.id.startsWith('custom.'));
    
  } catch (error) {
    console.error('Error fetching event dispatchers:', error);
    // Don't alert here as this is likely called on component mount
  }
}

// Define a reactive array for custom events 
const customEvents = ref([]);

// Replace the filteredEventDispatchers computed property
const filteredEventDispatchers = computed(() => {
  // Combine hardcoded events with custom events from server
  const allEvents = [
    // Default events
    { id: 'event-with-payload', name: 'Custom Event With Payload' },
    ...customEvents.value
  ];
  
  if (!searchQuery.value) return allEvents;

  return allEvents.filter(e =>
    e.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

// Call this when the component is mounted
onMounted(() => {
  fetchEventDispatchers();
});

// This should replace your existing implementation of onEventDragStart
function onEventDragStart(event, eventDispatcher) {
  if (!event.dataTransfer) return;

  console.log('Dragging event dispatcher:', eventDispatcher);

  // Create a node representation of this event
  const nodeData = {
    id: uuid(),
    type: 'event-dispatcher',
    position: { x: 0, y: 0 },
    properties: [
      { name: 'eventId', value: eventDispatcher.id },
      { name: 'eventName', value: eventDispatcher.name },
      { name: 'dynamicParameters', value: true }
    ]
  };

  event.dataTransfer.setData('application/json', JSON.stringify(nodeData));
  event.dataTransfer.effectAllowed = 'copy';
}

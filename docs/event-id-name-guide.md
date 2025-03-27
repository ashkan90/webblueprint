# Event Dispatchers and Event IDs Guide

## Understanding Event IDs and Event Names

In the WebBlueprint system, events have both a **name** and an **ID**:

### Event Name
- Human-readable identifier (e.g., "Player Died", "Button Clicked")
- Used in the UI for display purposes
- Easy to understand but may contain spaces or special characters

### Event ID
- Unique technical identifier (e.g., "custom.player-died", "system.initialize")
- Used internally by the event system to route events
- Generated automatically from the name (with special characters removed)
- Prefixed with a namespace ("custom." for user-defined events, "system." for system events)

## Creating Event Dispatchers

When you create a new event dispatcher:

1. You provide a **name** (e.g., "Player Scored") in the UI
2. The system automatically generates an **ID** ("custom.player-scored")
3. Both are stored in the event registry

## Using Event Dispatchers

When you place an **Event Dispatcher** node in your blueprint:

1. Select the event from the dropdown (shows the name)
2. The system automatically fills in the event ID on the node
3. When executed, the event is dispatched using the ID

## Event Binding

When you place an **On Event** node in your blueprint:

1. Select the event to listen for (shows the name)
2. The node will only trigger for events with the matching ID

## Entry Point Events

Entry point events (On Created, On Tick, etc.) are special system events:

1. They have fixed IDs that start with "event-on-"
2. They serve as starting points for blueprint execution
3. They're triggered automatically by the system based on specific conditions

## Best Practices

1. **Use descriptive names** for your event dispatchers
2. Let the system handle the conversion between names and IDs
3. Use the same event name across related components for clarity
4. When scripting, reference events by ID rather than name for reliability

## Technical Implementation Notes

- The `EventDispatcherNode` takes both an event ID and an event name
- The ID is used for actual dispatching
- The name is used for display and debugging purposes
- If you need to programmatically reference an event, always use the ID

Remember: The event system handles the mapping between names and IDs automatically, so you generally don't need to worry about the technical details when using the blueprint editor.

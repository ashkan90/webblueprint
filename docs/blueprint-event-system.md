# Blueprint Event System

The Blueprint Event System provides a powerful way to communicate between different parts of your blueprints, similar to Unreal Engine's event system. It consists of two main types of events:

## Entry Point Events

Entry point events serve as starting points for blueprint execution. These events are automatically triggered by the system based on specific conditions:

1. **On Created**: Triggered when a blueprint instance is created. This is the main entry point for initialization.

2. **On Tick**: Triggered periodically during blueprint execution. Useful for continuous updates or animations.

3. **On Input**: Triggered when input is received from the user interface.

Entry point events are available in the **ENTRY POINT EVENTS** section of the Blueprint Editor left panel. Simply drag them onto the canvas to use them as starting points for your execution flow.

## Event Dispatchers

Event dispatchers allow you to create and trigger custom events to communicate between different parts of your blueprints:

1. **Dispatch Event**: Sends a named event that can be received by any matching event receiver.

2. **Dispatch Event With Payload**: Sends an event with custom data that can be accessed by receivers.

Event dispatchers are available in the **EVENT DISPATCHERS** section of the Blueprint Editor left panel. You can also create custom event dispatchers for your specific needs.

## How to Use Events

### Creating an Execution Chain Starting with an Entry Point

1. Drag an **On Created** event from the ENTRY POINT EVENTS section to the canvas
2. Connect other nodes to its execution output pin
3. These nodes will automatically execute when the blueprint is created

### Communication Between Separate Execution Chains

1. In one execution chain, add a **Dispatch Event** node
2. Give it a descriptive name like "PlayerScored"
3. In another part of the blueprint, add an **On Event** node that listens for "PlayerScored"
4. Connect other nodes to the event receiver's execution pin
5. These nodes will execute whenever the "PlayerScored" event is dispatched

### Creating Custom Event Dispatchers

1. Click the + button in the EVENT DISPATCHERS section
2. Enter a name and description for your custom event
3. Your new event will appear in the list and can be used like any other event dispatcher

## Best Practices

1. Use **On Created** for initialization code
2. Use **On Tick** sparingly as it runs frequently
3. Give events clear, descriptive names
4. Use event parameters to pass data rather than relying on global variables
5. Consider using events as an alternative to deeply nested node connections

Events make your blueprints more modular, easier to understand, and simpler to maintain.

// File: web/src/services/eventService.ts

/**
 * Event parameter interface
 */
export interface EventParameter {
  name: string;
  type: string;
  description: string;
  optional: boolean;
  default?: any;
}

/**
 * Event definition interface
 */
export interface EventDefinition {
  id: string;
  name: string;
  description: string;
  category: string;
  parameters: EventParameter[];
  blueprintId?: string;
  createdAt?: string;
}

/**
 * Event binding interface
 */
export interface EventBinding {
  id: string;
  eventId: string;
  handlerId: string;
  handlerType: string;
  blueprintId: string;
  priority: number;
  enabled: boolean;
  createdAt?: string;
}

/**
 * Service for interacting with event endpoints
 */
export class EventService {
  /**
   * Fetch all available events from the server
   */
  static async fetchEvents(): Promise<EventDefinition[]> {
    try {
      const response = await fetch('/api/events');
      if (!response.ok) {
        throw new Error(`Failed to fetch events: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching events:', error);
      return [];
    }
  }

  /**
   * Fetch a specific event by ID
   */
  static async fetchEvent(eventId: string): Promise<EventDefinition> {
    try {
      const response = await fetch(`/api/events/${eventId}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch event: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error(`Error fetching event ${eventId}:`, error);
      throw error;
    }
  }

  /**
   * Fetch events for a specific blueprint
   */
  static async fetchBlueprintEvents(blueprintId: string): Promise<EventDefinition[]> {
    try {
      const response = await fetch(`/api/events/blueprint/${blueprintId}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch blueprint events: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error(`Error fetching blueprint events for ${blueprintId}:`, error);
      return [];
    }
  }

  /**
   * Create a new event
   */
  static async createEvent(eventData: Partial<EventDefinition>): Promise<EventDefinition> {
    try {
      const response = await fetch('/api/events', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(eventData),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to create event: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('Error creating event:', error);
      throw error;
    }
  }

  /**
   * Update an existing event
   */
  static async updateEvent(eventId: string, eventData: Partial<EventDefinition>): Promise<EventDefinition> {
    try {
      const response = await fetch(`/api/events/${eventId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(eventData),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to update event: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error(`Error updating event ${eventId}:`, error);
      throw error;
    }
  }

  /**
   * Delete an event
   */
  static async deleteEvent(eventId: string): Promise<void> {
    try {
      const response = await fetch(`/api/events/${eventId}`, {
        method: 'DELETE',
      });
      
      if (!response.ok) {
        throw new Error(`Failed to delete event: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Error deleting event ${eventId}:`, error);
      throw error;
    }
  }

  /**
   * Fetch all event bindings
   */
  static async fetchBindings(): Promise<EventBinding[]> {
    try {
      const response = await fetch('/api/events/bindings');
      if (!response.ok) {
        throw new Error(`Failed to fetch bindings: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching bindings:', error);
      return [];
    }
  }

  /**
   * Create a new event binding
   */
  static async createBinding(bindingData: Partial<EventBinding>): Promise<EventBinding> {
    try {
      const response = await fetch('/api/events/bindings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(bindingData),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to create binding: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('Error creating binding:', error);
      throw error;
    }
  }

  /**
   * Dispatch an event (useful for testing)
   */
  static async dispatchEvent(eventId: string, params: Record<string, any> = {}): Promise<{success: boolean, eventId: string}> {
    try {
      const response = await fetch('/api/events/dispatch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          eventId,
          params
        }),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to dispatch event: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error(`Error dispatching event ${eventId}:`, error);
      throw error;
    }
  }

  /**
   * Fetch system events
   */
  static async fetchSystemEvents(): Promise<EventDefinition[]> {
    try {
      const response = await fetch('/api/events/system');
      if (!response.ok) {
        throw new Error(`Failed to fetch system events: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching system events:', error);
      return [];
    }
  }
}

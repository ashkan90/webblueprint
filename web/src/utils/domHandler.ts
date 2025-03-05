/**
 * DOM Handler provides utilities for DOM manipulation from WebBlueprint nodes
 */
export class DOMHandler {
    private _elements: Map<string, HTMLElement> = new Map();
    private _eventListeners: Map<string, Function[]> = new Map();

    /**
     * Process a DOM operation from a node
     */
    processOperation(operation: any): HTMLElement | null {
        const { mode, selector, tagName, innerHTML, textContent, attributes, styles, parentSelector } = operation;

        try {
            // Handle different operation modes
            if (mode === 'create') {
                return this.createElement(tagName, innerHTML, textContent, attributes, styles, parentSelector);
            } else if (mode === 'modify') {
                return this.modifyElement(selector, innerHTML, textContent, attributes, styles);
            } else if (mode === 'remove') {
                return this.removeElement(selector);
            } else {
                console.error(`Unknown DOM operation mode: ${mode}`);
                return null;
            }
        } catch (error) {
            console.error('Error processing DOM operation:', error);
            return null;
        }
    }

    /**
     * Create a new DOM element
     */
    createElement(
        tagName: string = 'div',
        innerHTML?: string,
        textContent?: string,
        attributes?: Record<string, string>,
        styles?: Record<string, string>,
        parentSelector: string = 'body'
    ): HTMLElement | null {
        try {
            // Create the element
            const element = document.createElement(tagName);

            // Set content
            if (innerHTML !== undefined) {
                element.innerHTML = innerHTML;
            } else if (textContent !== undefined) {
                element.textContent = textContent;
            }

            // Set attributes
            if (attributes) {
                Object.entries(attributes).forEach(([attr, value]) => {
                    element.setAttribute(attr, value);
                });
            }

            // Set styles
            if (styles) {
                Object.entries(styles).forEach(([prop, value]) => {
                    (element.style as any)[prop] = value;
                });
            }

            // Generate a unique ID if not provided
            if (!element.id) {
                element.id = `wb-element-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
            }

            // Find parent and append
            const parent = document.querySelector(parentSelector);
            if (parent) {
                parent.appendChild(element);
            } else {
                console.warn(`Parent element not found: ${parentSelector}, appending to body`);
                document.body.appendChild(element);
            }

            // Store the element for future reference
            this._elements.set(element.id, element);

            return element;
        } catch (error) {
            console.error('Error creating element:', error);
            return null;
        }
    }

    /**
     * Modify an existing DOM element
     */
    modifyElement(
        selector: string,
        innerHTML?: string,
        textContent?: string,
        attributes?: Record<string, string>,
        styles?: Record<string, string>
    ): HTMLElement | null {
        try {
            // Find the element
            const element = document.querySelector(selector) as HTMLElement;
            if (!element) {
                console.warn(`Element not found: ${selector}`);
                return null;
            }

            // Update content
            if (innerHTML !== undefined) {
                element.innerHTML = innerHTML;
            } else if (textContent !== undefined) {
                element.textContent = textContent;
            }

            // Update attributes
            if (attributes) {
                Object.entries(attributes).forEach(([attr, value]) => {
                    if (value === null) {
                        element.removeAttribute(attr);
                    } else {
                        element.setAttribute(attr, value);
                    }
                });
            }

            // Update styles
            if (styles) {
                Object.entries(styles).forEach(([prop, value]) => {
                    if (value === null) {
                        (element.style as any)[prop] = '';
                    } else {
                        (element.style as any)[prop] = value;
                    }
                });
            }

            return element;
        } catch (error) {
            console.error('Error modifying element:', error);
            return null;
        }
    }

    /**
     * Remove a DOM element
     */
    removeElement(selector: string): HTMLElement | null {
        try {
            // Find the element
            const element = document.querySelector(selector) as HTMLElement;
            if (!element) {
                console.warn(`Element not found: ${selector}`);
                return null;
            }

            // Remove the element
            element.parentNode?.removeChild(element);

            // Remove from stored elements if it was tracked
            if (element.id && this._elements.has(element.id)) {
                this._elements.delete(element.id);
            }

            return element;
        } catch (error) {
            console.error('Error removing element:', error);
            return null;
        }
    }

    /**
     * Add an event listener to a DOM element
     */
    addEventListenerToElement(
        selector: string,
        eventType: string,
        callback: Function,
        useCapture: boolean = false
    ): void {
        try {
            // Find the element
            const element = document.querySelector(selector) as HTMLElement;
            if (!element) {
                console.warn(`Element not found: ${selector}`);
                return;
            }

            // Create a wrapper to maintain reference
            const listener = (event: Event) => {
                callback(event);
            };

            // Add the event listener
            element.addEventListener(eventType, listener, useCapture);

            // Store the listener for cleanup
            const key = `${selector}:${eventType}`;
            if (!this._eventListeners.has(key)) {
                this._eventListeners.set(key, []);
            }
            this._eventListeners.get(key)?.push(listener);
        } catch (error) {
            console.error('Error adding event listener:', error);
        }
    }

    /**
     * Remove event listeners from a DOM element
     */
    removeEventListenersFromElement(selector: string, eventType?: string): void {
        try {
            // Find the element
            const element = document.querySelector(selector) as HTMLElement;
            if (!element) {
                console.warn(`Element not found: ${selector}`);
                return;
            }

            // If event type is specified, remove only those listeners
            if (eventType) {
                const key = `${selector}:${eventType}`;
                const listeners = this._eventListeners.get(key);
                if (listeners) {
                    listeners.forEach(listener => {
                        element.removeEventListener(eventType, listener as any);
                    });
                    this._eventListeners.delete(key);
                }
            } else {
                // Remove all listeners for this element
                this._eventListeners.forEach((listeners, key) => {
                    if (key.startsWith(`${selector}:`)) {
                        const eventType = key.split(':')[1];
                        listeners.forEach(listener => {
                            element.removeEventListener(eventType, listener as any);
                        });
                        this._eventListeners.delete(key);
                    }
                });
            }
        } catch (error) {
            console.error('Error removing event listeners:', error);
        }
    }

    /**
     * Process a DOM event operation
     */
    processEventOperation(operation: any): void {
        const { selector, eventType, useCapture, preventDefault, stopPropagation, nodeId, executionId } = operation;

        try {
            // Create a callback to handle the event
            const callback = (event: Event) => {
                // Apply event modifiers
                if (preventDefault) {
                    event.preventDefault();
                }

                if (stopPropagation) {
                    event.stopPropagation();
                }

                // Send event data to the server
                this.sendEventToServer({
                    nodeId,
                    executionId,
                    event: {
                        type: event.type,
                        target: {
                            tagName: (event.target as HTMLElement)?.tagName,
                            id: (event.target as HTMLElement)?.id,
                            classList: Array.from((event.target as HTMLElement)?.classList || []),
                            value: (event.target as HTMLInputElement)?.value
                        }
                    }
                });
            };

            // Add the event listener
            this.addEventListenerToElement(selector, eventType, callback, useCapture);
        } catch (error) {
            console.error('Error processing DOM event operation:', error);
        }
    }

    /**
     * Send event data to the server
     */
    private sendEventToServer(data: any): void {
        // This would use the WebSocket connection to send the event to the server
        // For now, we'll just log it
        console.log('Sending event to server:', data);

        // In a real implementation, this would use the WebSocket store
        // websocketStore.send('dom.event', data);
    }

    /**
     * Cleanup all event listeners
     */
    cleanup(): void {
        // Remove all event listeners
        this._eventListeners.forEach((listeners, key) => {
            const [selector, eventType] = key.split(':');
            const element = document.querySelector(selector) as HTMLElement;
            if (element) {
                listeners.forEach(listener => {
                    element.removeEventListener(eventType, listener as any);
                });
            }
        });

        // Clear maps
        this._eventListeners.clear();
        this._elements.clear();
    }
}

// Create singleton instance
export const domHandler = new DOMHandler();
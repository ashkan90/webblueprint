// File: web/src/utils/storageHandler.ts
/**
 * Storage Handler provides utilities for local and session storage operations
 * This is used by the Storage node to interact with the browser's storage APIs
 */

export interface StorageOperation {
    operation: 'get' | 'set' | 'remove' | 'clear';
    storageType: 'local' | 'session';
    key?: string;
    value?: any;
    isComplex?: boolean;
    nodeId: string;
    executionId: string;
    timestamp: number;
}

export class StorageHandler {
    /**
     * Execute a storage operation
     */
    executeOperation(operation: StorageOperation): any {
        try {
            // Get the appropriate storage object
            const storage = operation.storageType === 'local' ? localStorage : sessionStorage;

            // Execute based on operation type
            switch(operation.operation) {
                case 'get':
                    return this.getValue(storage, operation.key || '');

                case 'set':
                    return this.setValue(storage, operation.key || '', operation.value, operation.isComplex);

                case 'remove':
                    return this.removeValue(storage, operation.key || '');

                case 'clear':
                    return this.clearStorage(storage);

                default:
                    throw new Error(`Unknown storage operation: ${operation.operation}`);
            }
        } catch (error) {
            console.error('Storage operation error:', error);
            return {
                success: false,
                error: error instanceof Error ? error.message : String(error)
            };
        }
    }

    /**
     * Get a value from storage
     */
    private getValue(storage: Storage, key: string): any {
        if (!key) {
            throw new Error('Key is required for get operation');
        }

        const storedValue = storage.getItem(key);

        // No value found
        if (storedValue === null) {
            return {
                success: true,
                exists: false,
                value: null
            };
        }

        // Try to parse as JSON first
        try {
            const parsedValue = JSON.parse(storedValue);
            return {
                success: true,
                exists: true,
                value: parsedValue,
                isComplex: true
            };
        } catch (e) {
            // Not valid JSON, return as string
            return {
                success: true,
                exists: true,
                value: storedValue,
                isComplex: false
            };
        }
    }

    /**
     * Set a value in storage
     */
    private setValue(storage: Storage, key: string, value: any, isComplex?: boolean): any {
        if (!key) {
            throw new Error('Key is required for set operation');
        }

        // If value is null or undefined, store as empty string
        if (value === null || value === undefined) {
            storage.setItem(key, '');
            return { success: true };
        }

        // For complex objects, stringify them
        if (isComplex || typeof value === 'object') {
            try {
                const jsonValue = JSON.stringify(value);
                storage.setItem(key, jsonValue);
            } catch (e) {
                throw new Error(`Could not stringify value: ${e instanceof Error ? e.message : String(e)}`);
            }
        } else {
            // For primitive values, convert to string
            storage.setItem(key, String(value));
        }

        return { success: true };
    }

    /**
     * Remove a value from storage
     */
    private removeValue(storage: Storage, key: string): any {
        if (!key) {
            throw new Error('Key is required for remove operation');
        }

        // Check if item exists first
        const exists = storage.getItem(key) !== null;

        // Remove the item
        storage.removeItem(key);

        return {
            success: true,
            existed: exists
        };
    }

    /**
     * Clear all values from storage
     */
    private clearStorage(storage: Storage): any {
        // Get the count before clearing
        const itemCount = storage.length;

        // Clear all items
        storage.clear();

        return {
            success: true,
            itemsCleared: itemCount
        };
    }

    /**
     * Get all keys from storage
     */
    getAllKeys(storageType: 'local' | 'session'): string[] {
        const storage = storageType === 'local' ? localStorage : sessionStorage;
        const keys: string[] = [];

        for (let i = 0; i < storage.length; i++) {
            const key = storage.key(i);
            if (key !== null) {
                keys.push(key);
            }
        }

        return keys;
    }

    /**
     * Get all items (key-value pairs) from storage
     */
    getAllItems(storageType: 'local' | 'session'): Record<string, any> {
        const storage = storageType === 'local' ? localStorage : sessionStorage;
        const items: Record<string, any> = {};

        for (let i = 0; i < storage.length; i++) {
            const key = storage.key(i);
            if (key !== null) {
                const result = this.getValue(storage, key);
                items[key] = result.value;
            }
        }

        return items;
    }

    /**
     * Check if storage is available
     */
    isStorageAvailable(type: 'localStorage' | 'sessionStorage'): boolean {
        try {
            const storage = window[type];
            const testKey = '__storage_test__';
            storage.setItem(testKey, testKey);
            storage.removeItem(testKey);
            return true;
        } catch (e) {
            return false;
        }
    }
}

// Create singleton instance
export const storageHandler = new StorageHandler();
// MockPersistentStorage.js
// A mock implementation of persistent storage for the Content Browser

// SyncOptions interface as specified
/**
 * @typedef {Object} SyncOptions
 * @property {boolean} autoSync - Enable automatic synchronization
 * @property {number} syncInterval - Sync interval in milliseconds
 * @property {'auto'|'manual'} conflictResolution - How to resolve conflicts
 * @property {boolean} offlineSupport - Support for offline operation
 */

// Asset type definitions
/**
 * @typedef {Object} Asset
 * @property {string} id - Unique identifier
 * @property {string} name - Asset name
 * @property {string} type - Asset type (e.g., "blueprint")
 * @property {string} path - Full path to the asset
 * @property {Date} created - Creation timestamp
 * @property {Date} modified - Last modified timestamp
 * @property {Object} metadata - Additional metadata
 * @property {Object} content - The actual asset content
 */

import {Asset, AssetType, Folder, SyncOptions} from "../types/mockPersistent";

/**
 * @typedef {Object} Folder
 * @property {string} id - Unique identifier
 * @property {string} name - Folder name
 * @property {string} path - Full path to the folder
 * @property {Date} created - Creation timestamp
 * @property {Date} modified - Last modified timestamp
 */

export class MockPersistentStorage {
    private assets: Asset[];
    private folders: Folder[];
    private syncOptions: SyncOptions;
    private currentSyncJob: number;
    private changeListeners: any[];
    constructor() {
        // Initialize with some mock data
        this.assets = [];
        this.folders = [];
        this.syncOptions = {
            autoSync: true,
            syncInterval: 30000,
            conflictResolution: 'auto',
            offlineSupport: true
        };

        this.currentSyncJob = null;
        this.changeListeners = [];

        // Populate with initial mock data
        this._initializeMockData();
    }

    /**
     * Initialize mock data for testing
     * @private
     */
    _initializeMockData() {
        // Mock folders
        this.folders = [
            { id: 'f1', name: 'Blueprints', path: '/Blueprints', created: new Date('2023-01-01'), modified: new Date('2023-01-01') },
            { id: 'f2', name: 'Characters', path: '/Blueprints/Characters', created: new Date('2023-01-02'), modified: new Date('2023-01-02') },
            { id: 'f3', name: 'Environment', path: '/Blueprints/Environment', created: new Date('2023-01-03'), modified: new Date('2023-01-03') },
            { id: 'f4', name: 'UI', path: '/Blueprints/UI', created: new Date('2023-01-04'), modified: new Date('2023-01-04') },
            { id: 'f5', name: 'Weapons', path: '/Blueprints/Weapons', created: new Date('2023-01-05'), modified: new Date('2023-01-05') },
            { id: 'f6', name: 'Props', path: '/Blueprints/Environment/Props', created: new Date('2023-01-06'), modified: new Date('2023-01-06') },
            { id: 'f7', name: 'Terrain', path: '/Blueprints/Environment/Terrain', created: new Date('2023-01-07'), modified: new Date('2023-01-07') },
        ];

        // Mock assets (blueprints)
        this.assets = [
            {
                id: 'a1',
                name: 'PlayerCharacter',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Characters/PlayerCharacter',
                created: new Date('2023-02-01'),
                modified: new Date('2023-03-15'),
                metadata: { author: 'John', version: '1.0' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a2',
                name: 'EnemyBase',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Characters/EnemyBase',
                created: new Date('2023-02-05'),
                modified: new Date('2023-03-10'),
                metadata: { author: 'Sarah', version: '1.2' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a3',
                name: 'Tree_01',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Environment/Props/Tree_01',
                created: new Date('2023-02-10'),
                modified: new Date('2023-02-15'),
                metadata: { author: 'Mike', version: '1.0' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a4',
                name: 'Rock_01',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Environment/Props/Rock_01',
                created: new Date('2023-02-12'),
                modified: new Date('2023-02-15'),
                metadata: { author: 'Mike', version: '1.0' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a5',
                name: 'MainMenu',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/UI/MainMenu',
                created: new Date('2023-02-20'),
                modified: new Date('2023-03-25'),
                metadata: { author: 'Lisa', version: '2.1' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a6',
                name: 'InventoryPanel',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/UI/InventoryPanel',
                created: new Date('2023-02-25'),
                modified: new Date('2023-03-20'),
                metadata: { author: 'Lisa', version: '1.5' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a7',
                name: 'Pistol',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Weapons/Pistol',
                created: new Date('2023-03-01'),
                modified: new Date('2023-03-30'),
                metadata: { author: 'John', version: '1.3' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a8',
                name: 'Rifle',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Weapons/Rifle',
                created: new Date('2023-03-05'),
                modified: new Date('2023-04-01'),
                metadata: { author: 'John', version: '1.2' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a9',
                name: 'TerrainGenerator',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Environment/Terrain/TerrainGenerator',
                created: new Date('2023-03-10'),
                modified: new Date('2023-04-05'),
                metadata: { author: 'Sarah', version: '1.0' },
                content: { nodes: [], connections: [] }
            },
            {
                id: 'a10',
                name: 'WeaponBase',
                type: AssetType.BLUEPRINT,
                path: '/Blueprints/Weapons/WeaponBase',
                created: new Date('2023-03-15'),
                modified: new Date('2023-03-18'),
                metadata: { author: 'John', version: '1.0' },
                content: { nodes: [], connections: [] }
            }
        ];
    }

    /**
     * Get sync options
     * @returns {SyncOptions} Current sync options
     */
    getSyncOptions() {
        return { ...this.syncOptions };
    }

    /**
     * Update sync options
     * @param {Partial<SyncOptions>} options - New options to apply
     */
    updateSyncOptions(options) {
        this.syncOptions = { ...this.syncOptions, ...options };

        // If autoSync was toggled, start or stop sync
        if (options.autoSync !== undefined) {
            if (options.autoSync) {
                this._startAutoSync();
            } else {
                this._stopAutoSync();
            }
        }

        return this.syncOptions;
    }

    /**
     * Start automatic sync based on current options
     * @private
     */
    _startAutoSync() {
        if (this.currentSyncJob) {
            clearInterval(this.currentSyncJob);
        }

        if (this.syncOptions.autoSync) {
            this.currentSyncJob = setInterval(() => {
                this._performSync();
            }, this.syncOptions.syncInterval);
        }
    }

    /**
     * Stop automatic sync
     * @private
     */
    _stopAutoSync() {
        if (this.currentSyncJob) {
            clearInterval(this.currentSyncJob);
            this.currentSyncJob = null;
        }
    }

    /**
     * Perform a mock sync operation
     * @private
     */
    _performSync() {
        console.log('Performing sync...');

        // In a real implementation, this would sync with the backend
        // For the mock, we'll just trigger a change event occasionally

        if (Math.random() > 0.7) {
            // Simulate a change being detected
            setTimeout(() => {
                this._notifyChangeListeners({
                    type: 'asset_updated',
                    id: this.assets[Math.floor(Math.random() * this.assets.length)].id,
                    timestamp: new Date()
                });
            }, 500);
        }
    }

    /**
     * Manually trigger a sync
     * @returns {Promise<void>}
     */
    async syncNow() {
        console.log('Manual sync triggered');
        return new Promise(resolve => {
            setTimeout(() => {
                this._performSync();
                resolve(null);
            }, 800); // Simulate network delay
        });
    }

    /**
     * Add a change listener
     * @param {Function} listener - Function to call when changes occur
     * @returns {Function} Function to remove the listener
     */
    addChangeListener(listener) {
        this.changeListeners.push(listener);
        return () => {
            this.changeListeners = this.changeListeners.filter(l => l !== listener);
        };
    }

    /**
     * Notify all change listeners
     * @param {Object} changeEvent - Description of the change
     * @private
     */
    _notifyChangeListeners(changeEvent) {
        this.changeListeners.forEach(listener => {
            try {
                listener(changeEvent);
            } catch (error) {
                console.error('Error in change listener:', error);
            }
        });
    }

    /**
     * Watch for changes (similar to filesystem watch)
     * @returns {Function} Function to stop watching
     */
    watchForChanges() {
        // In a real implementation, this would set up WebSocket listeners

        // For the mock, we'll just periodically check for changes
        const interval = setInterval(() => {
            if (Math.random() > 0.9) {
                this._notifyChangeListeners({
                    type: 'asset_updated',
                    id: this.assets[Math.floor(Math.random() * this.assets.length)].id,
                    timestamp: new Date()
                });
            }
        }, 20000); // Every 20 seconds

        return () => {
            clearInterval(interval);
        };
    }

    /**
     * Get a list of all assets
     * @returns {Asset[]} All assets
     */
    getAllAssets() {
        return [...this.assets];
    }

    /**
     * Get a list of all folders
     * @returns {Folder[]} All folders
     */
    getAllFolders() {
        return [...this.folders];
    }

    /**
     * Load an asset by ID
     * @param {string} id - Asset ID
     * @returns {Promise<Asset>} The asset
     */
    async loadAsset(id) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                const asset = this.assets.find(a => a.id === id);
                if (asset) {
                    resolve({ ...asset });
                } else {
                    reject(new Error(`Asset not found: ${id}`));
                }
            }, 300); // Simulate network delay
        });
    }

    /**
     * Save an asset
     * @param {Asset} asset - The asset to save
     * @returns {Promise<Asset>} The saved asset
     */
    async saveAsset(asset) {
        return new Promise((resolve) => {
            setTimeout(() => {
                // Find if the asset already exists
                const index = this.assets.findIndex(a => a.id === asset.id);

                if (index >= 0) {
                    // Update existing asset
                    const updatedAsset = {
                        ...this.assets[index],
                        ...asset,
                        modified: new Date()
                    };
                    this.assets[index] = updatedAsset;

                    this._notifyChangeListeners({
                        type: 'asset_updated',
                        id: asset.id,
                        timestamp: new Date()
                    });

                    resolve({ ...updatedAsset });
                } else {
                    // Create new asset
                    const newAsset = {
                        ...asset,
                        id: asset.id || `a${Date.now()}`,
                        created: new Date(),
                        modified: new Date()
                    };
                    this.assets.push(newAsset);

                    this._notifyChangeListeners({
                        type: 'asset_created',
                        id: newAsset.id,
                        timestamp: new Date()
                    });

                    resolve({ ...newAsset });
                }
            }, 500); // Simulate network delay
        });
    }

    /**
     * Delete an asset
     * @param {string} id - Asset ID
     * @returns {Promise<void>}
     */
    async deleteAsset(id) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                const index = this.assets.findIndex(a => a.id === id);
                if (index >= 0) {
                    this.assets.splice(index, 1);

                    this._notifyChangeListeners({
                        type: 'asset_deleted',
                        id: id,
                        timestamp: new Date()
                    });

                    resolve(null);
                } else {
                    reject(new Error(`Asset not found: ${id}`));
                }
            }, 500); // Simulate network delay
        });
    }

    /**
     * Create a new folder
     * @param {string} path - Parent path
     * @param {string} name - Folder name
     * @returns {Promise<Folder>} The created folder
     */
    async createFolder(path, name) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                // Check if folder already exists
                const folderPath = path.endsWith('/')
                    ? `${path}${name}`
                    : `${path}/${name}`;

                if (this.folders.some(f => f.path === folderPath)) {
                    reject(new Error(`Folder already exists: ${folderPath}`));
                    return;
                }

                const newFolder = {
                    id: `f${Date.now()}`,
                    name,
                    path: folderPath,
                    created: new Date(),
                    modified: new Date()
                };

                this.folders.push(newFolder);

                this._notifyChangeListeners({
                    type: 'folder_created',
                    id: newFolder.id,
                    timestamp: new Date()
                });

                resolve({ ...newFolder });
            }, 500); // Simulate network delay
        });
    }

    /**
     * Delete a folder
     * @param {string} path - Folder path
     * @returns {Promise<void>}
     */
    async deleteFolder(path) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                const index = this.folders.findIndex(f => f.path === path);
                if (index >= 0) {
                    const folderId = this.folders[index].id;

                    // Delete the folder
                    this.folders.splice(index, 1);

                    // Delete all assets in the folder
                    this.assets = this.assets.filter(a => !a.path.startsWith(path));

                    // Delete all subfolders
                    this.folders = this.folders.filter(f => !f.path.startsWith(path + '/'));

                    this._notifyChangeListeners({
                        type: 'folder_deleted',
                        id: folderId,
                        timestamp: new Date()
                    });

                    resolve(null);
                } else {
                    reject(new Error(`Folder not found: ${path}`));
                }
            }, 500); // Simulate network delay
        });
    }

    /**
     * Get the contents of a folder
     * @param {string} path - Folder path
     * @returns {Promise<{assets: Asset[], folders: Folder[]}>} Folder contents
     */
    async loadFolderContents(path) {
        return new Promise((resolve) => {
            setTimeout(() => {
                // Normalize path
                const normalizedPath = path.endsWith('/') ? path : `${path}`;

                // Get direct child folders
                const childFolders = this.folders.filter(f => {
                    // Skip the folder itself
                    if (f.path === normalizedPath) return false;

                    // Check if it's a direct child
                    // e.g., for path "/a/b", we want "/a/b/c" but not "/a/b/c/d"
                    const relativePath = f.path.startsWith(normalizedPath)
                        ? f.path.slice(normalizedPath.length)
                        : null;

                    if (!relativePath) return false;

                    // Skip the leading slash
                    const parts = relativePath.split('/').filter(Boolean);
                    return parts.length === 1;
                });

                // Get assets in this folder
                const folderAssets = this.assets.filter(a => {
                    const assetDir = a.path.substring(0, a.path.lastIndexOf('/'));
                    return assetDir === normalizedPath || assetDir + '/' === normalizedPath;
                });

                resolve({
                    assets: folderAssets,
                    folders: childFolders
                });
            }, 300); // Simulate network delay
        });
    }

    /**
     * Get folders at a specific path (for sidebar tree)
     * @param {string} path - Folder path
     * @returns {Folder[]} Child folders
     */
    getFoldersAt(path) {
        // Normalize path
        const normalizedPath = path.endsWith('/') ? path.slice(0, -1) : path;

        // Special case for root
        if (normalizedPath === '') {
            return this.folders.filter(f => {
                const parts = f.path.split('/').filter(Boolean);
                return parts.length === 1;
            });
        }

        // Get direct child folders
        return this.folders.filter(f => {
            // Skip the folder itself
            if (f.path === normalizedPath) return false;

            const parent = f.path.substring(0, f.path.lastIndexOf('/'));
            return parent === normalizedPath;
        });
    }

    /**
     * Search for assets
     * @param {Object} options - Search options
     * @param {string} [options.term] - Search term
     * @param {string} [options.type] - Asset type filter
     * @param {string} [options.path] - Path filter
     * @returns {Promise<Asset[]>} Matching assets
     */
    async searchAssets({ term, type, path }) {
        return new Promise((resolve) => {
            setTimeout(() => {
                let results = [...this.assets];

                if (term) {
                    const lowercaseTerm = term.toLowerCase();
                    results = results.filter(a =>
                        a.name.toLowerCase().includes(lowercaseTerm) ||
                        a.path.toLowerCase().includes(lowercaseTerm)
                    );
                }

                if (type) {
                    results = results.filter(a => a.type === type);
                }

                if (path) {
                    results = results.filter(a => a.path.startsWith(path));
                }

                resolve(results);
            }, 300); // Simulate network delay
        });
    }
}
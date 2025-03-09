export interface SyncOptions {
    autoSync: boolean
    offlineSupport: boolean
    syncInterval: number
    conflictResolution: 'auto' | 'manual'
}

export interface Asset {
    id: string
    name: string
    path: string
    created: Date
    modified: Date
    type: AssetType
    metadata: Record<string, string>
    content: Record<string, any>
}

export interface Folder {
    id: string
    name: string
    path: string
    created: Date
    modified: Date
}

export enum AssetType {
    BLUEPRINT = 'blueprint',
    FUNCTION = 'function',
    VARIABLE = 'variable',
    EVENT_DISPATCHER = 'event-dispatcher',
    MACRO = 'macro',
    NODE_TYPE = 'node-type',
    TYPE_DEFINITION = 'type-definition',
    PLUGIN = 'plugin',
    PRESET = 'preset'
}
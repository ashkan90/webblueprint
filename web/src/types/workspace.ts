
export interface Workspace {
    id: string
    name: string
    description: string | null
    ownerType: string
    ownerId: string
    createdAt: Date
    updatedAt: Date
    isPublic: boolean
    thumbnailURL: string|null
    metadata: Record<string, any>
}
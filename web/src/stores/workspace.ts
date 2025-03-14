import {defineStore} from "pinia";
import {ref} from "vue";
import {Workspace} from "../types/workspace";

export const useWorkspaceStore = defineStore('workspace', () => {
    const currentWorkspace = ref<Workspace | null>(null);
    const workspaces = ref<Workspace[]>([]);
    const isLoading = ref(false);
    const error = ref<string | null>(null);

    // Get all workspaces for the current user
    async function fetchWorkspaces() {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch('/api/workspaces');
            if (!response.ok) {
                throw new Error(`Failed to fetch workspaces: ${response.statusText}`);
            }

            workspaces.value = await response.json();
            return workspaces.value;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error('Error fetching workspaces:', err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Get a specific workspace
    async function fetchWorkspace(id: string) {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch(`/api/workspaces/${id}`);
            if (!response.ok) {
                throw new Error(`Failed to fetch workspace: ${response.statusText}`);
            }

            const data = await response.json();
            currentWorkspace.value = data;
            return data;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error(`Error fetching workspace ${id}:`, err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Create a new workspace
    async function createWorkspace(name: string, description: string, isPublic: boolean) {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch('/api/workspaces', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, description, isPublic })
            });

            if (!response.ok) {
                throw new Error(`Failed to create workspace: ${response.statusText}`);
            }

            const data = await response.json();
            currentWorkspace.value = data;
            
            // Add to workspaces list if we have one loaded
            if (workspaces.value.length > 0) {
                workspaces.value.push(data);
            }
            
            return data;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error('Error creating workspace:', err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Update an existing workspace
    async function updateWorkspace(id: string, name: string, description: string, isPublic: boolean) {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch(`/api/workspaces/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, description, isPublic })
            });

            if (!response.ok) {
                throw new Error(`Failed to update workspace: ${response.statusText}`);
            }

            const data = await response.json();
            
            // Update current workspace if it's the one we're editing
            if (currentWorkspace.value && currentWorkspace.value.id === id) {
                currentWorkspace.value = data;
            }
            
            // Update workspace in the list if it exists
            const index = workspaces.value.findIndex(w => w.id === id);
            if (index !== -1) {
                workspaces.value[index] = data;
            }
            
            return data;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error(`Error updating workspace ${id}:`, err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Delete a workspace
    async function deleteWorkspace(id: string) {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch(`/api/workspaces/${id}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                throw new Error(`Failed to delete workspace: ${response.statusText}`);
            }

            // Remove from workspaces list if it exists
            workspaces.value = workspaces.value.filter(w => w.id !== id);
            
            // Reset current workspace if it's the one we're deleting
            if (currentWorkspace.value && currentWorkspace.value.id === id) {
                currentWorkspace.value = null;
            }
            
            return true;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error(`Error deleting workspace ${id}:`, err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Get blueprints for a workspace
    async function fetchWorkspaceBlueprints(workspaceId: string) {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch(`/api/workspaces/${workspaceId}/blueprints`);
            if (!response.ok) {
                throw new Error(`Failed to fetch workspace blueprints: ${response.statusText}`);
            }

            return await response.json();
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error(`Error fetching blueprints for workspace ${workspaceId}:`, err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Set the current workspace
    function setCurrentWorkspace(workspace: Workspace | null) {
        currentWorkspace.value = workspace;
    }

    return {
        currentWorkspace,
        workspaces,
        isLoading,
        error,
        fetchWorkspaces,
        fetchWorkspace,
        createWorkspace,
        updateWorkspace,
        deleteWorkspace,
        fetchWorkspaceBlueprints,
        setCurrentWorkspace
    };
});
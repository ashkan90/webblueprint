import {defineStore} from "pinia";
import {ref} from "vue";

interface User {
    id: string;
    username: string;
    name: string;
    email: string;
    avatarUrl: string | null;
    createdAt: Date;
    updatedAt: Date;
    preferences: Record<string, any>;
}

export const useUserStore = defineStore('user', () => {
    const currentUser = ref<User | null>(null);
    const isLoading = ref(false);
    const error = ref<string | null>(null);

    // Get the current user
    async function fetchCurrentUser() {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch('/api/users/me');
            if (!response.ok) {
                throw new Error(`Failed to fetch user: ${response.statusText}`);
            }

            const data = await response.json();
            currentUser.value = data;
            return data;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error('Error fetching current user:', err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    // Update user preferences
    async function updateUserPreferences(preferences: Record<string, any>) {
        isLoading.value = true;
        error.value = null;

        try {
            const response = await fetch('/api/users/me/preferences', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ preferences })
            });

            if (!response.ok) {
                throw new Error(`Failed to update preferences: ${response.statusText}`);
            }

            const data = await response.json();
            if (currentUser.value) {
                currentUser.value.preferences = data.preferences;
            }
            
            return data;
        } catch (err) {
            error.value = err instanceof Error ? err.message : String(err);
            console.error('Error updating user preferences:', err);
            throw err;
        } finally {
            isLoading.value = false;
        }
    }

    return {
        currentUser,
        isLoading,
        error,
        fetchCurrentUser,
        updateUserPreferences
    };
});
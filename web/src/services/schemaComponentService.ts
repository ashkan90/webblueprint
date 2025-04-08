// web/src/services/schemaComponentService.ts

import type { SchemaComponent } from '../types/schemaComponent'; // Use relative path

const API_BASE_URL = '/api'; // Adjust if your API base is different

// Basic error handling - replace with a more robust solution if available
async function handleResponse<T>(response: Response): Promise<T> {
	if (!response.ok) {
		const errorData = await response.text();
		throw new Error(`API Error (${response.status}): ${errorData || response.statusText}`);
	}
	if (response.status === 204) { // Handle No Content
		return null as T;
	}
	return response.json() as Promise<T>;
}

export const schemaComponentService = {
	async listSchemaComponents(): Promise<SchemaComponent[]> {
		const response = await fetch(`${API_BASE_URL}/schema-components`);
		return handleResponse<SchemaComponent[]>(response);
	},

	async createSchemaComponent(name: string, schemaDefinition: string): Promise<SchemaComponent> {
		const response = await fetch(`${API_BASE_URL}/schema-components`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({ name, schema_definition: schemaDefinition }),
		});
		return handleResponse<SchemaComponent>(response);
	},

	async getSchemaComponent(id: string): Promise<SchemaComponent> {
		const response = await fetch(`${API_BASE_URL}/schema-components/${id}`);
		return handleResponse<SchemaComponent>(response);
	},

	async updateSchemaComponent(id: string, name: string, schemaDefinition: string): Promise<SchemaComponent> {
		const response = await fetch(`${API_BASE_URL}/schema-components/${id}`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({ name, schema_definition: schemaDefinition }),
		});
		return handleResponse<SchemaComponent>(response);
	},

	async deleteSchemaComponent(id: string): Promise<void> {
		const response = await fetch(`${API_BASE_URL}/schema-components/${id}`, {
			method: 'DELETE',
		});
		await handleResponse<void>(response); // Checks for errors, returns null on 204
	},
};
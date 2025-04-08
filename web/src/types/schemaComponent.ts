// web/src/types/schemaComponent.ts

export interface SchemaComponent {
	id: string;
	name: string;
	schema_definition: string; // Matches Go backend field name
	created_at: string; // ISO 8601 date string
	updated_at: string; // ISO 8601 date string
}
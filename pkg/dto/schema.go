package dto

import "time"

type SchemaDefinition struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	SchemaDefinition string    `json:"schema_definition"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONB represents a PostgreSQL JSONB type
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(data, j)
}

// JSONArray represents a PostgreSQL JSONB array
type JSONArray []interface{}

// Value implements the driver.Valuer interface for JSONArray
func (a JSONArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for JSONArray
func (a *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(data, a)
}

// StringArray represents a PostgreSQL text[] type
type StringArray []string

// Value implements the driver.Valuer interface for StringArray
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return "{" + join(a, ",") + "}", nil
}

// Scan implements the sql.Scanner interface for StringArray
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	// Cast to string and remove braces
	strVal, ok := value.(string)
	if !ok {
		bytesVal, ok := value.([]byte)
		if !ok {
			return errors.New("failed to convert array to string")
		}
		strVal = string(bytesVal)
	}

	// Parse the string array
	// This is simplified and might need improvement for real escaping
	*a = parseArray(strVal)
	return nil
}

// Helper function to join string array with a separator
func join(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}

	result := ""
	for i, s := range a {
		if i > 0 {
			result += sep
		}
		result += `"` + s + `"`
	}
	return result
}

// Helper function to parse PostgreSQL array format
func parseArray(s string) []string {
	// This is a simplified version that doesn't handle all PostgreSQL array cases
	// In a real implementation, you'd want to handle escaping and more complex parsing
	if len(s) < 2 {
		return []string{}
	}

	// Remove braces { }
	s = s[1 : len(s)-1]

	// Split by comma and remove quotes
	// This is very simplified and doesn't handle quoted commas
	if s == "" {
		return []string{}
	}

	// Split and clean up
	result := make([]string, 0)
	current := ""
	inQuote := false

	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// User represents a user in the system
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	AvatarURL    sql.NullString
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  sql.NullTime
	IsActive     bool
}

// Workspace represents a workspace container for assets
type Workspace struct {
	ID           string
	Name         string
	Description  sql.NullString
	OwnerType    string // "user" or "team"
	OwnerID      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsPublic     bool
	ThumbnailURL sql.NullString
	Metadata     JSONB
}

// WorkspaceMember represents a user's membership in a workspace
type WorkspaceMember struct {
	WorkspaceID string
	UserID      string
	Role        string
	JoinedAt    time.Time
	User        *User // Optional related user data
}

// Asset represents any asset in the system
type Asset struct {
	ID           string
	WorkspaceID  string
	Name         string
	Description  sql.NullString
	Type         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
	IsPublic     bool
	Tags         StringArray
	ThumbnailURL sql.NullString
	Metadata     JSONB
}

// Blueprint represents a visual programming blueprint
type Blueprint struct {
	Asset            // Embedded Asset fields
	CurrentVersionID sql.NullString
	NodeCount        int
	ConnectionCount  int
	EntryPoints      StringArray
	IsTemplate       bool
	Category         sql.NullString
	CurrentVersion   *BlueprintVersion // Optional related current version
}

// BlueprintVersion represents a specific version of a blueprint
type BlueprintVersion struct {
	ID            string
	BlueprintID   string
	VersionNumber int
	CreatedAt     time.Time
	CreatedBy     string
	Comment       sql.NullString
	Nodes         JSONArray
	Connections   JSONArray
	Variables     JSONArray
	Functions     JSONArray
	Events        JSONArray
	Metadata      JSONB
}

// NodeCategory represents a category of node types
type NodeCategory struct {
	ID          string
	Name        string
	Description sql.NullString
	Color       sql.NullString
	Icon        sql.NullString
	SortOrder   int
}

// NodeType represents a type of node that can be used in blueprints
type NodeType struct {
	ID           string
	Name         string
	Description  sql.NullString
	CategoryID   sql.NullString
	Version      string
	Author       sql.NullString
	AuthorURL    sql.NullString
	Icon         sql.NullString
	IsCore       bool
	IsDeprecated bool
	Inputs       JSONArray
	Outputs      JSONArray
	Properties   JSONArray
	Metadata     JSONB
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Execution represents an execution of a blueprint
type Execution struct {
	ID               string
	BlueprintID      string
	VersionID        sql.NullString
	StartedAt        time.Time
	CompletedAt      sql.NullTime
	Status           string
	InitiatedBy      string
	ExecutionMode    string
	InitialVariables JSONB
	Result           JSONB
	Error            sql.NullString
	DurationMs       sql.NullInt32
}

// ExecutionNode represents execution data for a single node
type ExecutionNode struct {
	ExecutionID string
	NodeID      string
	NodeType    string
	StartedAt   sql.NullTime
	CompletedAt sql.NullTime
	Status      string
	Inputs      JSONB
	Outputs     JSONB
	Error       sql.NullString
	DurationMs  sql.NullInt32
	DebugData   JSONB
}

// ExecutionLog represents a log entry during execution
type ExecutionLog struct {
	ID          string
	ExecutionID string
	NodeID      sql.NullString
	LogLevel    string
	Message     string
	Details     JSONB
	Timestamp   time.Time
}

// AssetReference represents a reference between assets
type AssetReference struct {
	SourceAssetID  string
	TargetAssetID  string
	ReferenceType  string
	ReferenceCount int
	Details        JSONB
}

// BlueprintDependency represents a dependency between blueprints
type BlueprintDependency struct {
	BlueprintID       string
	DependencyID      string
	DependencyType    string
	IsOptional        bool
	VersionConstraint sql.NullString
}

// Variable represents a blueprint variable
type Variable struct {
	ID           string
	BlueprintID  string
	Name         string
	Type         string
	DefaultValue JSONB
	Description  sql.NullString
	IsExposed    bool
	Category     sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Function represents a blueprint function
type Function struct {
	ID            string
	BlueprintID   string
	Name          string
	Description   sql.NullString
	NodeInterface JSONB
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	UpdatedBy     string
	IsPublic      bool
	Category      sql.NullString
	Version       string
}

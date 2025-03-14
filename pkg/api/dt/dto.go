package dt

import (
	"time"
)

type User struct {
	ID          string    `json:"id,omitempty"`
	Username    string    `json:"username,omitempty"`
	Email       string    `json:"email,omitempty"`
	FullName    string    `json:"full_name,omitempty"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
	IsActive    bool      `json:"is_active,omitempty"`
}

type Workspace struct {
	ID           string                 `json:"id,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	OwnerType    string                 `json:"owner_type,omitempty"` // "user" or "team"
	OwnerID      string                 `json:"owner_id,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	IsPublic     bool                   `json:"is_public,omitempty"`
	ThumbnailURL string                 `json:"thumbnail_url,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

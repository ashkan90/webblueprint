package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"webblueprint/pkg/api/dt"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL-based user repository
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}

	query := `
		INSERT INTO users (
			id, username, email, password_hash, full_name, avatar_url,
			role, created_at, updated_at, last_login_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
		user.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash, full_name, avatar_url,
			role, created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash, full_name, avatar_url,
			role, created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE username = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("error retrieving user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash, full_name, avatar_url,
			role, created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("error retrieving user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	// Always update the updated_at timestamp
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, full_name = $4,
			avatar_url = $5, role = $6, updated_at = $7, is_active = $8
		WHERE id = $9
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Role,
		user.UpdatedAt,
		user.IsActive,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// Delete deletes a user by ID
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// UpdateLastLogin updates a user's last login time
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

func (r *PostgresUserRepository) ToDt(user *models.User) *dt.User {
	return &dt.User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		AvatarURL:   user.AvatarURL.String,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt.Time,
		IsActive:    user.IsActive,
	}
}

package service

import (
	"context"
	"fmt"
	"time"
	"webblueprint/pkg/api/dt"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides high-level operations for managing users
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with the provided information
func (s *UserService) CreateUser(
	ctx context.Context,
	username, email, password, fullName string,
) (string, error) {
	// Check if username is taken
	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return "", fmt.Errorf("username already taken")
	}

	// Check if email is taken
	_, err = s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return "", fmt.Errorf("email already taken")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	// Create the user
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		Role:         "user", // Default role
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Save to repository
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return "", fmt.Errorf("error creating user: %w", err)
	}

	return user.ID, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*dt.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	// Don't return the password hash
	user.PasswordHash = ""

	return s.userRepo.ToDt(user), nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	// Don't return the password hash
	user.PasswordHash = ""

	return user, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(
	ctx context.Context,
	id string,
	email, fullName, newPassword string,
) error {
	// Get the user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update the fields if provided
	if email != "" && email != user.Email {
		// Check if email is taken by another user
		existingUser, err := s.userRepo.GetByEmail(ctx, email)
		if err == nil && existingUser.ID != id {
			return fmt.Errorf("email already taken")
		}
		user.Email = email
	}

	if fullName != "" {
		user.FullName = fullName
	}

	if newPassword != "" {
		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("error hashing password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()

	// Save changes
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

// VerifyCredentials checks if the username/password combination is valid
func (s *UserService) VerifyCredentials(ctx context.Context, username, password string) (*models.User, error) {
	// Get the user
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// Don't reveal if the username exists
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Don't return the password hash
	user.PasswordHash = ""

	return user, nil
}

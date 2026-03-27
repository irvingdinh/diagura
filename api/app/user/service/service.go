package service

import (
	"context"
	"fmt"
	"time"

	authservice "localhost/app/auth/service"
	"localhost/app/core/sqlite"
	"localhost/app/core/sqlite/orm"
	"localhost/app/core/utils"
	"localhost/app/user/entity"
)

const userRoleID = "00000000-0000-7000-0000-000000000002"

// Service provides user management operations.
type Service struct {
	db *sqlite.DB
}

// NewService creates a Service with the given database.
func NewService(db *sqlite.DB) *Service {
	return &Service{db: db}
}

// List returns users with optional pagination.
func (s *Service) List(ctx context.Context, limit, offset int) ([]entity.User, error) {
	b := orm.Select("id", "role_id", "email", "name", "created_at", "updated_at").
		From("users").
		Where("deleted_at IS NULL").
		OrderBy("created_at", "DESC")
	if limit > 0 {
		b = b.Limit(limit)
	}
	if offset > 0 {
		b = b.Offset(offset)
	}
	query, args := b.Build()

	return orm.QueryAll[entity.User](s.db, query, args...)
}

// CreateInput holds the fields required to create a new user.
type CreateInput struct {
	Email    string
	Name     string
	Password string
}

// CreateResult holds the fields returned after creating a user.
type CreateResult struct {
	ID    string
	Email string
	Name  string
}

// Create registers a new user with the User role.
func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateResult, error) {
	passwordHash, err := authservice.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := utils.FormatTime(time.Now())
	id := orm.NewID()

	query, args := orm.Insert("users").
		Set("id", id).
		Set("role_id", userRoleID).
		Set("email", input.Email).
		Set("name", input.Name).
		Set("password_hash", passwordHash).
		Set("created_at", now).
		Set("updated_at", now).
		Build()

	if _, err := s.db.Exec(query, args...); err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &CreateResult{
		ID:    id,
		Email: input.Email,
		Name:  input.Name,
	}, nil
}

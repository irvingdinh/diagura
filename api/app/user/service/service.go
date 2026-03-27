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

// Service provides user management operations.
type Service struct {
	db *sqlite.DB
}

// NewService creates a Service with the given database.
func NewService(db *sqlite.DB) *Service {
	return &Service{db: db}
}

// ---------------------------------------------------------------------------
// Read operations
// ---------------------------------------------------------------------------

// List returns users with optional pagination.
func (s *Service) List(ctx context.Context, limit, offset int) ([]entity.User, error) {
	b := orm.Select("id", "role_id", "email", "name", "force_password_change", "created_at", "updated_at").
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

// ListFilter holds filter and pagination parameters for ListPaginated.
type ListFilter struct {
	Search  string
	Role    string
	Status  string
	Page    int
	PerPage int
}

// ListResult holds paginated results.
type ListResult struct {
	Users []entity.UserWithRole
	Total int
}

// ListPaginated returns a filtered, paginated list of users with role info.
func (s *Service) ListPaginated(ctx context.Context, f ListFilter) (*ListResult, error) {
	baseSelect := orm.Select(
		"u.id", "u.email", "u.name", "u.force_password_change",
		"u.deleted_at", "u.created_at", "u.updated_at",
		"r.slug", "r.name",
	).
		From("users u").
		Join("roles r", "r.id = u.role_id")

	countSelect := orm.Select("COUNT(*)").
		From("users u").
		Join("roles r", "r.id = u.role_id")

	// Status filter.
	if f.Status == "deleted" {
		baseSelect = baseSelect.Where("u.deleted_at IS NOT NULL")
		countSelect = countSelect.Where("u.deleted_at IS NOT NULL")
	} else {
		baseSelect = baseSelect.Where("u.deleted_at IS NULL")
		countSelect = countSelect.Where("u.deleted_at IS NULL")
	}

	// Search filter (name OR email partial match).
	if f.Search != "" {
		nameCond, namePattern := orm.LikeCondition("u.name", f.Search)
		emailCond, emailPattern := orm.LikeCondition("u.email", f.Search)
		baseSelect = baseSelect.Where("("+nameCond+" OR "+emailCond+")", "%"+namePattern+"%", "%"+emailPattern+"%")
		countSelect = countSelect.Where("("+nameCond+" OR "+emailCond+")", "%"+namePattern+"%", "%"+emailPattern+"%")
	}

	// Role filter.
	if f.Role != "" {
		baseSelect = baseSelect.Where("r.slug = ?", f.Role)
		countSelect = countSelect.Where("r.slug = ?", f.Role)
	}

	// Count query.
	countQuery, countArgs := countSelect.Build()
	total, err := orm.QueryVal[int64](s.db, countQuery, countArgs...)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	// Data query with pagination.
	offset := (f.Page - 1) * f.PerPage
	dataQuery, dataArgs := baseSelect.
		OrderBy("u.created_at", "DESC").
		Limit(f.PerPage).
		Offset(offset).
		Build()

	users, err := orm.QueryAll[entity.UserWithRole](s.db, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return &ListResult{Users: users, Total: int(total)}, nil
}

// GetByID returns a single user with role info.
func (s *Service) GetByID(ctx context.Context, id string) (entity.UserWithRole, error) {
	query, args := orm.Select(
		"u.id", "u.email", "u.name", "u.force_password_change",
		"u.deleted_at", "u.created_at", "u.updated_at",
		"r.slug", "r.name",
	).
		From("users u").
		Join("roles r", "r.id = u.role_id").
		Where("u.id = ?", id).
		Build()

	return orm.QueryOne[entity.UserWithRole](s.db, query, args...)
}

// GetRoleBySlug returns a role by its slug.
func (s *Service) GetRoleBySlug(ctx context.Context, slug string) (entity.Role, error) {
	query, args := orm.Select("id", "slug", "name").
		From("roles").
		Where("slug = ?", slug).
		Build()

	return orm.QueryOne[entity.Role](s.db, query, args...)
}

// EmailExists checks whether an active user with the given email exists,
// optionally excluding a specific user ID (for update uniqueness checks).
func (s *Service) EmailExists(ctx context.Context, email, excludeUserID string) (bool, error) {
	b := orm.Select("COUNT(*)").
		From("users").
		Where("email = ?", email).
		Where("deleted_at IS NULL")
	if excludeUserID != "" {
		b = b.Where("id != ?", excludeUserID)
	}
	query, args := b.Build()

	count, err := orm.QueryVal[int64](s.db, query, args...)
	if err != nil {
		return false, fmt.Errorf("check email: %w", err)
	}
	return count > 0, nil
}

// ---------------------------------------------------------------------------
// Write operations
// ---------------------------------------------------------------------------

// CreateInput holds the fields required to create a new user.
type CreateInput struct {
	Email    string
	Name     string
	Password string
	RoleID   string
}

// CreateResult holds the fields returned after creating a user.
type CreateResult struct {
	ID    string
	Email string
	Name  string
}

// Create registers a new user with the specified role.
func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateResult, error) {
	passwordHash, err := authservice.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := utils.FormatTime(time.Now())
	id := orm.NewID()

	query, args := orm.Insert("users").
		Set("id", id).
		Set("role_id", input.RoleID).
		Set("email", input.Email).
		Set("name", input.Name).
		Set("password_hash", passwordHash).
		Set("force_password_change", true).
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

// UpdateProfile updates a user's name (self-service).
func (s *Service) UpdateProfile(ctx context.Context, userID, name string) error {
	query, args := orm.Update("users").
		Set("name", name).
		Set("updated_at", utils.FormatTime(time.Now())).
		Where("id = ?", userID).
		Build()
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// UpdateUserInput holds optional fields for admin user updates.
type UpdateUserInput struct {
	Name   *string
	Email  *string
	RoleID *string
}

// UpdateUser applies partial updates to a user record.
func (s *Service) UpdateUser(ctx context.Context, userID string, input UpdateUserInput) error {
	b := orm.Update("users")
	if input.Name != nil {
		b = b.Set("name", *input.Name)
	}
	if input.Email != nil {
		b = b.Set("email", *input.Email)
	}
	if input.RoleID != nil {
		b = b.Set("role_id", *input.RoleID)
	}
	b = b.Set("updated_at", utils.FormatTime(time.Now())).
		Where("id = ?", userID)

	query, args := b.Build()
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// SetPassword updates a user's password hash and force_password_change flag.
func (s *Service) SetPassword(ctx context.Context, userID, passwordHash string, forceChange bool) error {
	query, args := orm.Update("users").
		Set("password_hash", passwordHash).
		Set("force_password_change", forceChange).
		Set("updated_at", utils.FormatTime(time.Now())).
		Where("id = ?", userID).
		Build()
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("set password: %w", err)
	}
	return nil
}

// SoftDelete marks a user as deleted by setting deleted_at.
func (s *Service) SoftDelete(ctx context.Context, userID string) error {
	now := utils.FormatTime(time.Now())
	query, args := orm.Update("users").
		Set("deleted_at", now).
		Set("updated_at", now).
		Where("id = ?", userID).
		Build()
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return nil
}

// Restore clears deleted_at and sets force_password_change to true.
func (s *Service) Restore(ctx context.Context, userID string) error {
	now := utils.FormatTime(time.Now())
	query, args := orm.Update("users").
		Set("deleted_at", nil).
		Set("force_password_change", true).
		Set("updated_at", now).
		Where("id = ?", userID).
		Build()
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("restore user: %w", err)
	}
	return nil
}

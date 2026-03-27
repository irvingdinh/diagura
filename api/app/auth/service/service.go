package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"localhost/app/core/sqlite"
	"localhost/app/core/sqlite/orm"
	"localhost/app/core/utils"
)

const (
	sessionSlidingDays  = 30
	sessionAbsoluteDays = 365
)

// ---------------------------------------------------------------------------
// Context types
// ---------------------------------------------------------------------------

type ctxKey int

const (
	ctxKeyUser ctxKey = iota
	ctxKeySession
)

// AuthUser holds the authenticated user's identity, set on the request
// context by the auth middleware.
type AuthUser struct {
	ID       string
	Email    string
	Name     string
	RoleSlug string
}

// AuthSession holds session metadata, set on the request context by
// the auth middleware.
type AuthSession struct {
	ID        string
	UserID    string
	IPAddress string
	UserAgent string
	CreatedAt string
}

// WithUser stores an AuthUser in the context.
func WithUser(ctx context.Context, u *AuthUser) context.Context {
	return context.WithValue(ctx, ctxKeyUser, u)
}

// UserFromContext extracts the AuthUser from the context.
func UserFromContext(ctx context.Context) (*AuthUser, bool) {
	u, ok := ctx.Value(ctxKeyUser).(*AuthUser)
	return u, ok
}

// WithSession stores an AuthSession in the context.
func WithSession(ctx context.Context, s *AuthSession) context.Context {
	return context.WithValue(ctx, ctxKeySession, s)
}

// SessionFromContext extracts the AuthSession from the context.
func SessionFromContext(ctx context.Context) (*AuthSession, bool) {
	s, ok := ctx.Value(ctxKeySession).(*AuthSession)
	return s, ok
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// Service provides session management and validation.
type Service struct {
	db *sqlite.DB
}

// NewService creates a Service with the given database.
func NewService(db *sqlite.DB) *Service {
	return &Service{db: db}
}

// HashToken returns the hex-encoded SHA-256 hash of a raw session token.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

const validateQuery = `SELECT s.id, s.user_id, s.expires_at, s.absolute_expires_at,
       s.ip_address, s.user_agent, s.created_at,
       u.id, u.email, u.name, u.deleted_at,
       r.slug
FROM sessions s
JOIN users u ON u.id = s.user_id
JOIN roles r ON r.id = u.role_id
WHERE s.token_hash = ?`

// ValidateSession looks up a session by its token hash, validates expiry
// and user state, and returns the authenticated user and session metadata.
func (s *Service) ValidateSession(ctx context.Context, tokenHash string) (*AuthUser, *AuthSession, error) {
	row := s.db.QueryRow(validateQuery, tokenHash)

	var (
		sessID, sessUserID, expiresAt, absoluteExpiresAt string
		ipAddress, userAgent, sessCreatedAt              string
		userID, email, name                              string
		deletedAt                                        any
		roleSlug                                         string
	)

	if err := row.Scan(
		&sessID, &sessUserID, &expiresAt, &absoluteExpiresAt,
		&ipAddress, &userAgent, &sessCreatedAt,
		&userID, &email, &name, &deletedAt,
		&roleSlug,
	); err != nil {
		return nil, nil, fmt.Errorf("session not found")
	}

	if deletedAt != nil {
		return nil, nil, fmt.Errorf("user deleted")
	}

	now := time.Now().UTC()
	exp, err := utils.ParseTime(expiresAt)
	if err != nil {
		return nil, nil, fmt.Errorf("parse expires_at: %w", err)
	}
	absExp, err := utils.ParseTime(absoluteExpiresAt)
	if err != nil {
		return nil, nil, fmt.Errorf("parse absolute_expires_at: %w", err)
	}
	if now.After(exp) || now.After(absExp) {
		return nil, nil, fmt.Errorf("session expired")
	}

	user := &AuthUser{
		ID:       userID,
		Email:    email,
		Name:     name,
		RoleSlug: roleSlug,
	}
	session := &AuthSession{
		ID:        sessID,
		UserID:    sessUserID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: sessCreatedAt,
	}

	return user, session, nil
}

// userRow is an internal type for scanning the user+role JOIN query.
type userRow struct {
	ID           string `db:"u.id"`
	Email        string `db:"u.email"`
	Name         string `db:"u.name"`
	PasswordHash string `db:"u.password_hash"`
	RoleSlug     string `db:"r.slug"`
}

// AuthenticateByEmail verifies the email/password combination and returns
// the authenticated user. Transparently rehashes the password if the stored
// hash uses outdated parameters.
func (s *Service) AuthenticateByEmail(ctx context.Context, email, password string) (*AuthUser, error) {
	query, args := orm.Select("u.id", "u.email", "u.name", "u.password_hash", "r.slug").
		From("users u").
		Join("roles r", "r.id = u.role_id").
		Where("u.email = ?", email).
		Where("u.deleted_at IS NULL").
		Build()

	row, err := orm.QueryOne[userRow](s.db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	ok, err := VerifyPassword(row.PasswordHash, password)
	if err != nil || !ok {
		return nil, fmt.Errorf("invalid credentials")
	}

	if NeedsRehash(row.PasswordHash) {
		if newHash, err := HashPassword(password); err == nil {
			q, a := orm.Update("users").
				Set("password_hash", newHash).
				Set("updated_at", utils.FormatTime(time.Now())).
				Where("id = ?", row.ID).
				Build()
			_, _ = s.db.Exec(q, a...)
		}
	}

	return &AuthUser{
		ID:       row.ID,
		Email:    row.Email,
		Name:     row.Name,
		RoleSlug: row.RoleSlug,
	}, nil
}

// CreateSession generates a new session token, stores a hashed version in
// the database, and returns the raw token to be set as a cookie.
func (s *Service) CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	rawToken := "dgr_s_" + base64.RawURLEncoding.EncodeToString(raw)
	tokenHash := HashToken(rawToken)

	now := time.Now().UTC()
	nowStr := orm.FormatTime(now)
	expiresAt := orm.FormatTime(now.Add(sessionSlidingDays * 24 * time.Hour))
	absoluteExpiresAt := orm.FormatTime(now.Add(sessionAbsoluteDays * 24 * time.Hour))

	query, args := orm.Insert("sessions").
		Set("id", orm.NewID()).
		Set("user_id", userID).
		Set("token_hash", tokenHash).
		Set("expires_at", expiresAt).
		Set("absolute_expires_at", absoluteExpiresAt).
		Set("ip_address", ipAddress).
		Set("user_agent", userAgent).
		Set("created_at", nowStr).
		Set("updated_at", nowStr).
		Build()

	if _, err := s.db.Exec(query, args...); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}

	return rawToken, nil
}

// DeleteSession removes a session by its ID.
func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	query, args := orm.Delete("sessions").Where("id = ?", sessionID).Build()
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// ExtendSession updates the sliding expiry window for an active session.
func (s *Service) ExtendSession(ctx context.Context, sessionID string) error {
	now := time.Now().UTC()
	query, args := orm.Update("sessions").
		Set("expires_at", orm.FormatTime(now.Add(sessionSlidingDays*24*time.Hour))).
		Set("updated_at", orm.FormatTime(now)).
		Where("id = ?", sessionID).
		Build()

	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("extend session: %w", err)
	}
	return nil
}

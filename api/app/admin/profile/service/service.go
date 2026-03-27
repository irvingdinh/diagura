package service

import (
	"context"
	"errors"
	"fmt"

	authservice "localhost/app/auth/service"
	userservice "localhost/app/user/service"
)

var ErrIncorrectPassword = errors.New("current password is incorrect")

// Service provides profile management operations.
type Service struct {
	userSvc *userservice.Service
	authSvc *authservice.Service
}

// NewService creates a Service with the given dependencies.
func NewService(userSvc *userservice.Service, authSvc *authservice.Service) *Service {
	return &Service{userSvc: userSvc, authSvc: authSvc}
}

// ChangePassword verifies the current password, sets the new one,
// and invalidates all other sessions.
func (s *Service) ChangePassword(ctx context.Context, email, currentPassword, newPassword, userID, keepSessionID string) error {
	if _, err := s.authSvc.AuthenticateByEmail(ctx, email, currentPassword); err != nil {
		return ErrIncorrectPassword
	}

	hash, err := authservice.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userSvc.SetPassword(ctx, userID, hash, false); err != nil {
		return fmt.Errorf("set password: %w", err)
	}

	if err := s.authSvc.DeleteOtherSessions(ctx, userID, keepSessionID); err != nil {
		return fmt.Errorf("delete other sessions: %w", err)
	}

	return nil
}

// LogoutOtherSessions invalidates all sessions except the current one.
func (s *Service) LogoutOtherSessions(ctx context.Context, userID, keepSessionID string) error {
	return s.authSvc.DeleteOtherSessions(ctx, userID, keepSessionID)
}

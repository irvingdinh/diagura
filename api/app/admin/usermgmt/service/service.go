package service

import (
	"context"

	authservice "localhost/app/auth/service"
)

// Service provides user management orchestration.
type Service struct {
	authSvc *authservice.Service
}

// NewService creates a Service with the given dependencies.
func NewService(authSvc *authservice.Service) *Service {
	return &Service{authSvc: authSvc}
}

// InvalidateAllSessions removes every session for the given user.
func (s *Service) InvalidateAllSessions(ctx context.Context, userID string) error {
	return s.authSvc.DeleteAllUserSessions(ctx, userID)
}

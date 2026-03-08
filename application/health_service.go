package application

import (
	"context"
	"log/slog"
)

// HealthService implements the HealthChecker interface.
type HealthService struct {
	logger *slog.Logger
	repo   MetricsRepository
}

// NewHealthService creates a new HealthService.
func NewHealthService(logger *slog.Logger, repo MetricsRepository) *HealthService {
	return &HealthService{
		logger: logger,
		repo:   repo,
	}
}

// Check verifies the service is healthy by pinging the database.
func (s *HealthService) Check(ctx context.Context) (string, error) {
	if err := s.repo.Ping(ctx); err != nil {
		s.logger.Error("health check failed: database unreachable", "error", err)
		return "unhealthy", err
	}
	return "ok", nil
}

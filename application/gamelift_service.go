package application

import (
	"context"
	"log/slog"
)

// GameLiftService implements GameLiftMetricsService.
type GameLiftService struct {
	logger *slog.Logger
	repo   GameLiftMetricsRepository
}

// NewGameLiftService creates a new GameLiftService.
func NewGameLiftService(logger *slog.Logger, repo GameLiftMetricsRepository) *GameLiftService {
	return &GameLiftService{logger: logger, repo: repo}
}

func (s *GameLiftService) Ingest(ctx context.Context, req GameLiftIngestRequest) error {
	m := GameLiftMetric{
		FleetID:              req.FleetID,
		ActiveSessions:       req.ActiveSessions,
		ActivePlayers:        req.ActivePlayers,
		AvailableSessions:    req.AvailableSessions,
		PercentIdleInstances: req.PercentIdleInstances,
		CurrentInstances:     req.CurrentInstances,
		DesiredInstances:     req.DesiredInstances,
	}
	if err := s.repo.InsertGameLiftMetric(ctx, m); err != nil {
		s.logger.Error("failed to ingest GameLift metric", "fleet_id", req.FleetID, "error", err)
		return err
	}
	s.logger.Info("GameLift metric ingested", "fleet_id", req.FleetID)
	return nil
}

func (s *GameLiftService) GetByFleet(ctx context.Context, fleetID string) ([]GameLiftMetric, error) {
	return s.repo.GetGameLiftMetricsByFleet(ctx, fleetID)
}

func (s *GameLiftService) List(ctx context.Context) ([]GameLiftMetric, error) {
	return s.repo.ListGameLiftMetrics(ctx)
}

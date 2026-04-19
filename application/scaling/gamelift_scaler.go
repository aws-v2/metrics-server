package scaling

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"metrics-gateway/repository"
	"metrics-gateway/internal/messaging"

	"github.com/nats-io/nats.go"
)

// GameLiftScaler periodically checks GameLift metrics and triggers scale events via NATS.
type GameLiftScaler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	repo    *repository.PostgresRepository
	ctx     context.Context
	cancel  context.CancelFunc
	profile string
}

func NewGameLiftScaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository, profile string) *GameLiftScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameLiftScaler{
		logger:  logger,
		nc:      nc,
		repo:    repo,
		ctx:     ctx,
		cancel:  cancel,
		profile: profile,
	}
}

func (s *GameLiftScaler) Start() {
	s.logger.Info("starting GameLift scaling module...")
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.evaluate()
			}
		}
	}()
}

func (s *GameLiftScaler) Stop() {
	s.cancel()
}

func (s *GameLiftScaler) evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	upCandidates, err := s.repo.GetGameLiftFleetsToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get GameLift scale up candidates", "error", err)
	} else {
		for _, c := range upCandidates {
			s.logger.Info("GameLift scale out triggered", "fleet", c.FleetID, "metric", c.MetricName, "value", c.AvgMetricValue)
			
			payload, _ := json.Marshal(map[string]interface{}{
				"tenant_id":   c.TenantID,
				"fleet_id":    c.FleetID,
				"reason":      "metric_above_threshold",
				"metric":      c.MetricName,
				"value":       c.AvgMetricValue,
				"action":      "INCREASE_DESIRED_INSTANCES",
			})
			
			_ = s.nc.Publish(messaging.BuildSubject(s.profile, "gamelift", "v1", "scale", "out"), payload)
		}
	}

	downCandidates, err := s.repo.GetGameLiftFleetsToScaleDown(ctx)
	if err != nil {
		s.logger.Error("failed to get GameLift scale down candidates", "error", err)
	} else {
		for _, c := range downCandidates {
			s.logger.Info("GameLift scale in triggered", "fleet", c.FleetID, "metric", c.MetricName, "value", c.AvgMetricValue)
			
			payload, _ := json.Marshal(map[string]interface{}{
				"tenant_id":   c.TenantID,
				"fleet_id":    c.FleetID,
				"reason":      "metric_below_threshold",
				"metric":      c.MetricName,
				"value":       c.AvgMetricValue,
				"action":      "DECREASE_DESIRED_INSTANCES",
			})
			
			_ = s.nc.Publish(messaging.BuildSubject(s.profile, "gamelift", "v1", "scale", "in"), payload)
		}
	}
}




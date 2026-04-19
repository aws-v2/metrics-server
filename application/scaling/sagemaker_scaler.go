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

// SageMakerScaler periodically checks SageMaker metrics and triggers scale events via NATS.
type SageMakerScaler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	repo    *repository.PostgresRepository
	ctx     context.Context
	cancel  context.CancelFunc
	profile string
}

func NewSageMakerScaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository, profile string) *SageMakerScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &SageMakerScaler{
		logger:  logger,
		nc:      nc,
		repo:    repo,
		ctx:     ctx,
		cancel:  cancel,
		profile: profile,
	}
}

func (s *SageMakerScaler) Start() {
	s.logger.Info("starting SageMaker scaling module...")
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

func (s *SageMakerScaler) Stop() {
	s.cancel()
}

func (s *SageMakerScaler) evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	upCandidates, err := s.repo.GetSageMakerEndpointsToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get SageMaker scale up candidates", "error", err)
	} else {
		for _, c := range upCandidates {
			s.logger.Info("SageMaker scale out triggered", "endpoint", c.EndpointName, "metric", c.MetricName, "value", c.AvgMetricValue)
			
			payload, _ := json.Marshal(map[string]interface{}{
				"tenant_id":   c.TenantID,
				"endpoint_id": c.EndpointName,
				"reason":      "metric_above_threshold",
				"metric":      c.MetricName,
				"value":       c.AvgMetricValue,
				"action":      "INCREASE_INSTANCE_COUNT",
			})
			
			_ = s.nc.Publish(messaging.BuildSubject(s.profile, "sagemaker", "v1", "scale", "out"), payload)
		}
	}

	downCandidates, err := s.repo.GetSageMakerEndpointsToScaleDown(ctx)
	if err != nil {
		s.logger.Error("failed to get SageMaker scale down candidates", "error", err)
	} else {
		for _, c := range downCandidates {
			s.logger.Info("SageMaker scale in triggered", "endpoint", c.EndpointName, "metric", c.MetricName, "value", c.AvgMetricValue)
			
			payload, _ := json.Marshal(map[string]interface{}{
				"tenant_id":   c.TenantID,
				"endpoint_id": c.EndpointName,
				"reason":      "metric_below_threshold",
				"metric":      c.MetricName,
				"value":       c.AvgMetricValue,
				"action":      "DECREASE_INSTANCE_COUNT",
			})
			
			_ = s.nc.Publish(messaging.BuildSubject(s.profile, "sagemaker", "v1", "scale", "in"), payload)
		}
	}
}
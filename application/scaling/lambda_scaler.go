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

// LambdaScaler periodically checks Lambda metrics and triggers scale events via NATS.
type LambdaScaler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	repo    *repository.PostgresRepository
	ctx     context.Context
	cancel  context.CancelFunc
	profile string
}

func NewLambdaScaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository, profile string) *LambdaScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &LambdaScaler{
		logger:  logger,
		nc:      nc,
		repo:    repo,
		ctx:     ctx,
		cancel:  cancel,
		profile: profile,
	}
}

func (s *LambdaScaler) Start() {
	s.logger.Info("starting Lambda scaling module...")
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

func (s *LambdaScaler) Stop() {
	s.cancel()
}

func (s *LambdaScaler) evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	upCandidates, err := s.repo.GetLambdaFunctionsToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get Lambda scale up candidates", "error", err)
	} else {
		for _, c := range upCandidates {
			s.logger.Info("Lambda scale out triggered", "function", c.FunctionID, "metric", c.MetricName, "value", c.AvgMetricValue)
			
			payload, _ := json.Marshal(map[string]interface{}{
				"tenant_id":   c.TenantID,
				"function_id": c.FunctionID,
				"reason":      "metric_above_threshold",
				"metric":      c.MetricName,
				"value":       c.AvgMetricValue,
				"action":      "INCREASE_PROVISIONED_CONCURRENCY",
			})
			
			_ = s.nc.Publish(messaging.BuildSubject(s.profile, "lambda", "v1", "scale", "out"), payload)
		}
	}

	downCandidates, err := s.repo.GetLambdaFunctionsToScaleDown(ctx)
	if err != nil {
		s.logger.Error("failed to get Lambda scale down candidates", "error", err)
	} else {
		for _, c := range downCandidates {
			s.logger.Info("Lambda scale in triggered", "function", c.FunctionID, "metric", c.MetricName, "value", c.AvgMetricValue)
			
			payload, _ := json.Marshal(map[string]interface{}{
				"tenant_id":   c.TenantID,
				"function_id": c.FunctionID,
				"reason":      "metric_below_threshold",
				"metric":      c.MetricName,
				"value":       c.AvgMetricValue,
				"action":      "DECREASE_PROVISIONED_CONCURRENCY",
			})
			
			_ = s.nc.Publish(messaging.BuildSubject(s.profile, "lambda", "v1", "scale", "in"), payload)
		}
	}
}

package scaling

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"metrics-gateway/repository"

	"github.com/nats-io/nats.go"
)

// LambdaScaler periodically checks Lambda metrics and triggers scale events via NATS.
type LambdaScaler struct {
	logger *slog.Logger
	nc     *nats.Conn
	repo   *repository.PostgresRepository
	ctx    context.Context
	cancel context.CancelFunc
}

func NewLambdaScaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository) *LambdaScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &LambdaScaler{logger: logger, nc: nc, repo: repo, ctx: ctx, cancel: cancel}
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

	candidates, err := s.repo.GetLambdaFunctionsToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get Lambda scale candidates", "error", err)
		return
	}

	for _, c := range candidates {
		s.logger.Info("Lambda scale/provisioned concurrency triggered", "function", c.FunctionID, "throttles", c.TotalThrottles)
		
		payload, _ := json.Marshal(map[string]interface{}{
			"function_id": c.FunctionID,
			"reason":      "high_throttles",
			"throttles":   c.TotalThrottles,
			"action":      "INCREASE_PROVISIONED_CONCURRENCY",
		})
		
		_ = s.nc.Publish("dev.lambda.v1.scale.concurrency", payload)
	}
}

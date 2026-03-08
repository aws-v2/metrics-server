package scaling

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"metrics-gateway/repository"

	"github.com/nats-io/nats.go"
)

// RDSScaler periodically checks RDS metrics and triggers scale events via NATS.
type RDSScaler struct {
	logger *slog.Logger
	nc     *nats.Conn
	repo   *repository.PostgresRepository
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRDSScaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository) *RDSScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &RDSScaler{logger: logger, nc: nc, repo: repo, ctx: ctx, cancel: cancel}
}

func (s *RDSScaler) Start() {
	s.logger.Info("starting RDS scaling module...")
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

func (s *RDSScaler) Stop() {
	s.cancel()
}

func (s *RDSScaler) evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	candidates, err := s.repo.GetRDSInstancesToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get RDS scale candidates", "error", err)
		return
	}

	for _, c := range candidates {
		s.logger.Info("RDS scale-up triggered", "instance", c.InstanceID, "avg_cpu", c.AvgCPUPercent)
		
		payload, _ := json.Marshal(map[string]interface{}{
			"instance_id": c.InstanceID,
			"reason":      "cpu_threshold_exceeded",
			"avg_cpu":     c.AvgCPUPercent,
			"action":      "SCALE_UP_STORAGE_OR_COMPUTE",
		})
		
		_ = s.nc.Publish("dev.rds.v1.scale.up", payload)
	}
}

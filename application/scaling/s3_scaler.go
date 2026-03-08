package scaling

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"metrics-gateway/repository"

	"github.com/nats-io/nats.go"
)

// S3Scaler periodically checks S3 metrics and triggers lifecycle or storage scale events via NATS.
type S3Scaler struct {
	logger *slog.Logger
	nc     *nats.Conn
	repo   *repository.PostgresRepository
	ctx    context.Context
	cancel context.CancelFunc
}

func NewS3Scaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository) *S3Scaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &S3Scaler{logger: logger, nc: nc, repo: repo, ctx: ctx, cancel: cancel}
}

func (s *S3Scaler) Start() {
	s.logger.Info("starting S3 scaling module...")
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

func (s *S3Scaler) Stop() {
	s.cancel()
}

func (s *S3Scaler) evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	candidates, err := s.repo.GetS3BucketsToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get S3 scale candidates", "error", err)
		return
	}

	for _, c := range candidates {
		s.logger.Info("S3 storage expansion requested", "bucket", c.BucketID, "storage_used_bytes", c.StorageUsedBytes)
		
		payload, _ := json.Marshal(map[string]interface{}{
			"bucket_id":    c.BucketID,
			"reason":       "storage_limit_approaching",
			"storage_used": c.StorageUsedBytes,
			"action":       "ALLOCATE_MORE_STORAGE",
		})
		
		_ = s.nc.Publish("dev.s3.v1.scale.storage", payload)
	}
}

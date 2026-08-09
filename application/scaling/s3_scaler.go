package scaling

import (
	"context" 
	"log/slog"
	"time"

	"metrics-gateway/repository" 

	"github.com/nats-io/nats.go"
)

// S3Scaler periodically checks S3 metrics and triggers lifecycle or storage scale events via NATS.
type S3Scaler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	repo    *repository.PostgresRepository
	ctx     context.Context
	cancel  context.CancelFunc
	profile string
}

func NewS3Scaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository, profile string) *S3Scaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &S3Scaler{
		logger:  logger,
		nc:      nc,
		repo:    repo,
		ctx:     ctx,
		cancel:  cancel,
		profile: profile,
	}
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
		// TODO: 	THE underscore here shouldbe replacedwith ctx,'
	//to get the background 
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
 
}

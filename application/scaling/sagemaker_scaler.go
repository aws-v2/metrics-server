package scaling

import (
	"context"
	"log/slog"
	"time"

	"metrics-gateway/repository"

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
	// TODO: 	THE underscore here shouldbe replacedwith ctx,'
	//to get the background contex
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
 
}
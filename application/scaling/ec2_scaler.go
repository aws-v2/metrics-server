package scaling

import (
	"context" 
	"log/slog"
	"time"

	"metrics-gateway/repository" 

	"github.com/nats-io/nats.go"
)

// EC2Scaler periodically checks EC2 metrics and triggers scale events via NATS.
type EC2Scaler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	repo    *repository.PostgresRepository
	ctx     context.Context
	cancel  context.CancelFunc
	profile string
}

// NewEC2Scaler initializes the scaler.
func NewEC2Scaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository, profile string) *EC2Scaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &EC2Scaler{
		logger:  logger,
		nc:      nc,
		repo:    repo,
		ctx:     ctx,
		cancel:  cancel,
		profile: profile,
	}
}

// Start begins the scaling background loop.
func (s *EC2Scaler) Start() {
	s.logger.Info("starting EC2 scaling module...")
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				s.logger.Info("stopping EC2 scaling module")
				return
			case <-ticker.C:
				s.evaluate()
			}
		}
	}()
}

// Stop halts the scaler.
func (s *EC2Scaler) Stop() {
	s.cancel()
}

func (s *EC2Scaler) evaluate() {
		// TODO: 	THE underscore here shouldbe replacedwith ctx,'
	//to get the background 
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()
}

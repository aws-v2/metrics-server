package scaling

import (
	"context" 
	"log/slog"
	"time"

	"metrics-gateway/repository"  

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
		// TODO: 	THE underscore here shouldbe replacedwith ctx,'
	//to get the background 
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

}

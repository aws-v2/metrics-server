package scaling

import (
	"context"
	"log/slog"
	"time"

	"metrics-gateway/repository"

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
	// TODO: 	THE underscore here shouldbe replacedwith ctx,'
	//to get the background 
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	
}




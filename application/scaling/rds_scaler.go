package scaling

import (
	"context" 
	"log/slog" 
	"time"
 
	"metrics-gateway/repository"

	"github.com/nats-io/nats.go"
)

// RDSScaler periodically checks RDS metrics and triggers scale events via NATS.
type RDSScaler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	repo    *repository.PostgresRepository
	ctx     context.Context
	cancel  context.CancelFunc
	profile string
}

func NewRDSScaler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository, profile string) *RDSScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &RDSScaler{
		logger:  logger,
		nc:      nc,
		repo:    repo,
		ctx:     ctx,
		cancel:  cancel,
		profile: profile,
	}
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

type RDSVerticalScaleAlarmPayload struct {
	InstanceID   string  `json:"instance_id"`
	Action       string  `json:"action"` // SCALE_UP or SCALE_DOWN
	ResourceType string  `json:"resource_type"` // CPU or MEMORY
	NewLimit     float64 `json:"new_limit"`
	TenantID     string  `json:"tenant_id"`
	Reason       string  `json:"reason"`
}

var lastRSDActions = make(map[string]time.Time)

func (s *RDSScaler) evaluate() {
		// TODO: 	THE underscore here shouldbe replacedwith ctx,'
	//to get the background 
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() 
}

func (s *RDSScaler) metricToResource(metricName string) string {
	if metricName == "MemoryUtilization" {
		return "MEMORY"
	}
	return "CPU"
}

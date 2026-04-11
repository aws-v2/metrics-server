package scaling

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"metrics-gateway/internal/messaging"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Evaluate Scale Out Candidates (Vertical Scale Up)
	outCandidates, err := s.repo.GetRDSVerticalScaleUpCandidates(ctx)
	if err != nil {
		s.logger.Error("failed to get RDS scale-out candidates", "error", err)
	} else {
		for _, c := range outCandidates {
			s.handleScaling(c, "SCALE_OUT")
		}
	}

	// 2. Evaluate Scale In Candidates (Vertical Scale Down)
	inCandidates, err := s.repo.GetRDSVerticalScaleDownCandidates(ctx)
	if err != nil {
		s.logger.Error("failed to get RDS scale-in candidates", "error", err)
	} else {
		for _, c := range inCandidates {
			s.handleScaling(c, "SCALE_IN")
		}
	}
}

func (s *RDSScaler) handleScaling(c repository.RDSVerticalScaleCandidate, action string) {
	key := c.InstanceID + "_" + c.MetricName + "_" + action
	if last, ok := lastRSDActions[key]; ok && time.Since(last) < 5*time.Minute {
		return // Cooldown
	}

	s.logger.Info("RDS vertical scaling triggered",
		"instance", c.InstanceID,
		"action", action,
		"metric", c.MetricName,
		"avg_value", c.AvgMetricValue)

	payload := RDSVerticalScaleAlarmPayload{
		InstanceID:   c.InstanceID,
		Action:       action,
		ResourceType: s.metricToResource(c.MetricName),
		NewLimit:     0, // Placeholder: rds-server will calculate based on its internal state
		TenantID:     c.TenantID,
		Reason:       "dynamic_threshold_breach",
	}

	data, _ := json.Marshal(payload)
	sub := messaging.BuildSubject(s.profile, "rds", "v1", "scale", strings.ToLower(action))
	if action == "SCALE_OUT" {
		sub = messaging.BuildSubject(s.profile, "rds", "v1", "scale", "out")
	} else if action == "SCALE_IN" {
		sub = messaging.BuildSubject(s.profile, "rds", "v1", "scale", "in")
	}

	if err := s.nc.Publish(sub, data); err != nil {
		s.logger.Error("failed to publish RDS scale alarm", "error", err, "subject", sub)
		return
	}

	lastRSDActions[key] = time.Now()
}

func (s *RDSScaler) metricToResource(metricName string) string {
	if metricName == "MemoryUtilization" {
		return "MEMORY"
	}
	return "CPU"
}

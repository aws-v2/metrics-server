package scaling

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"metrics-gateway/repository"
	"metrics-gateway/internal/messaging"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	candidates, err := s.repo.GetEC2InstancesToScaleUp(ctx)
	if err != nil {
		s.logger.Error("failed to get EC2 scale candidates", "error", err)
		return
	}

	// EC2ScalingAlarmPayload represents the struct sent to EC2 service.
	type EC2ScalingAlarmPayload struct {
		CorrelationID string `json:"correlation_id"`
		TenantID      string `json:"tenant_id"`
		Action        string `json:"action"`
		Policy        struct {
			TargetType     string  `json:"target_type"`
			TargetID       string  `json:"target_id"`
			MetricName     string  `json:"metric_name"`
			TargetValue    float64 `json:"target_value"`
			ScaleDownValue float64 `json:"scale_down_value"`
			MaxInstances   int     `json:"max_instances"`
		} `json:"policy"`
		CurrentValue float64 `json:"current_value"`
	}

	for _, c := range candidates {
		s.logger.Info("EC2 scale out alarm triggered", "instance", c.InstanceID, "avg_cpu", c.AvgCPUPercent)
		
		correlationID := "alarm-" + time.Now().Format("20060102150405.000000000") // simple unique ID without external deps
		
		alarm := EC2ScalingAlarmPayload{
			CorrelationID: correlationID,
			TenantID:      c.TenantID,
			Action:        "scale_out",
			CurrentValue:  c.AvgCPUPercent,
		}
		alarm.Policy.TargetType = c.TargetType
		alarm.Policy.TargetID = c.TargetID
		alarm.Policy.MetricName = c.MetricName
		alarm.Policy.TargetValue = c.TargetValue
		
		payload, _ := json.Marshal(alarm)
		
		_ = s.nc.Publish(messaging.BuildSubject(s.profile, "ec2", "v1", "scale", "out"), payload)
	}

	downCandidates, err := s.repo.GetEC2InstancesToScaleDown(ctx)
	if err != nil {
		s.logger.Error("failed to get EC2 scale down candidates", "error", err)
		return
	}

	for _, c := range downCandidates {
		s.logger.Info("EC2 scale in alarm triggered", "instance", c.InstanceID, "avg_cpu", c.AvgCPUPercent)

		correlationID := "alarm-" + time.Now().Format("20060102150405.000000000")

		alarm := EC2ScalingAlarmPayload{
			CorrelationID: correlationID,
			TenantID:      c.TenantID,
			Action:        "scale_in",
			CurrentValue:  c.AvgCPUPercent,
		}
		alarm.Policy.TargetType = c.TargetType
		alarm.Policy.TargetID = c.TargetID
		alarm.Policy.MetricName = c.MetricName
		alarm.Policy.ScaleDownValue = c.ScaleDownValue

		payload, _ := json.Marshal(alarm)

		_ = s.nc.Publish(messaging.BuildSubject(s.profile, "ec2", "v1", "scale", "in"), payload)
	}
}

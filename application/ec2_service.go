package application

import (
	"context"
	"log/slog"
)

// EC2Service implements EC2MetricsService.
type EC2Service struct {
	logger *slog.Logger
	repo   EC2MetricsRepository
}

// NewEC2Service creates a new EC2Service.
func NewEC2Service(logger *slog.Logger, repo EC2MetricsRepository) *EC2Service {
	return &EC2Service{logger: logger, repo: repo}
}

func (s *EC2Service) Ingest(ctx context.Context, req EC2IngestRequest) error {
	m := EC2Metric{
		InstanceID:  req.InstanceID,
		CPUPercent:  req.CPUPercent,
		MemTotalMB:  req.MemTotalMB,
		MemUsedMB:   req.MemUsedMB,
		MemPercent:  req.MemPercent,
		DiskTotalGB: req.DiskTotalGB,
		DiskUsedGB:  req.DiskUsedGB,
	}
	if err := s.repo.InsertEC2Metric(ctx, m); err != nil {
		s.logger.Error("failed to ingest EC2 metric", "instance_id", req.InstanceID, "error", err)
		return err
	}
	s.logger.Info("EC2 metric ingested", "instance_id", req.InstanceID)
	return nil
}

func (s *EC2Service) GetByInstance(ctx context.Context, instanceID string) ([]EC2Metric, error) {
	return s.repo.GetEC2MetricsByInstance(ctx, instanceID)
}

func (s *EC2Service) List(ctx context.Context) ([]EC2Metric, error) {
	return s.repo.ListEC2Metrics(ctx)
}

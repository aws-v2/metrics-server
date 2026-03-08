package application

import (
	"context"
	"log/slog"
)

// RDSService implements RDSMetricsService.
type RDSService struct {
	logger *slog.Logger
	repo   RDSMetricsRepository
}

// NewRDSService creates a new RDSService.
func NewRDSService(logger *slog.Logger, repo RDSMetricsRepository) *RDSService {
	return &RDSService{logger: logger, repo: repo}
}

func (s *RDSService) Ingest(ctx context.Context, req RDSIngestRequest) error {
	m := RDSMetric{
		InstanceID:     req.InstanceID,
		CPUPercent:     req.CPUPercent,
		MemUsedMB:      req.MemUsedMB,
		MemPercent:     req.MemPercent,
		StorageUsedGB:  req.StorageUsedGB,
		StorageTotalGB: req.StorageTotalGB,
		Connections:    req.Connections,
		IOPS:           req.IOPS,
	}
	if err := s.repo.InsertRDSMetric(ctx, m); err != nil {
		s.logger.Error("failed to ingest RDS metric", "instance_id", req.InstanceID, "error", err)
		return err
	}
	s.logger.Info("RDS metric ingested", "instance_id", req.InstanceID)
	return nil
}

func (s *RDSService) GetByInstance(ctx context.Context, instanceID string) ([]RDSMetric, error) {
	return s.repo.GetRDSMetricsByInstance(ctx, instanceID)
}

func (s *RDSService) List(ctx context.Context) ([]RDSMetric, error) {
	return s.repo.ListRDSMetrics(ctx)
}

package application

import (
	"context"
	"log/slog"
)

// SageMakerService implements SageMakerMetricsService.
type SageMakerService struct {
	logger *slog.Logger
	repo   SageMakerMetricsRepository
}

// NewSageMakerService creates a new SageMakerService.
func NewSageMakerService(logger *slog.Logger, repo SageMakerMetricsRepository) *SageMakerService {
	return &SageMakerService{logger: logger, repo: repo}
}

func (s *SageMakerService) Ingest(ctx context.Context, req SageMakerIngestRequest) error {
	m := SageMakerMetric{
		EndpointID:     req.EndpointID,
		Invocations:    req.Invocations,
		ModelLatencyMs: req.ModelLatencyMs,
		CPUPercent:     req.CPUPercent,
		MemPercent:     req.MemPercent,
		GPUPercent:     req.GPUPercent,
		GPUMemPercent:  req.GPUMemPercent,
		Errors4xx:      req.Errors4xx,
		Errors5xx:      req.Errors5xx,
	}
	if err := s.repo.InsertSageMakerMetric(ctx, m); err != nil {
		s.logger.Error("failed to ingest SageMaker metric", "endpoint_id", req.EndpointID, "error", err)
		return err
	}
	s.logger.Info("SageMaker metric ingested", "endpoint_id", req.EndpointID)
	return nil
}

func (s *SageMakerService) GetByEndpoint(ctx context.Context, endpointID string) ([]SageMakerMetric, error) {
	return s.repo.GetSageMakerMetricsByEndpoint(ctx, endpointID)
}

func (s *SageMakerService) List(ctx context.Context) ([]SageMakerMetric, error) {
	return s.repo.ListSageMakerMetrics(ctx)
}

package application

import (
	"context"
	"log/slog"
)

// LambdaService implements LambdaMetricsService.
type LambdaService struct {
	logger *slog.Logger
	repo   LambdaMetricsRepository
}

// NewLambdaService creates a new LambdaService.
func NewLambdaService(logger *slog.Logger, repo LambdaMetricsRepository) *LambdaService {
	return &LambdaService{logger: logger, repo: repo}
}

func (s *LambdaService) Ingest(ctx context.Context, req LambdaIngestRequest) error {
	m := LambdaMetric{
		FunctionID:     req.FunctionID,
		Invocations:    req.Invocations,
		Errors:         req.Errors,
		Throttles:      req.Throttles,
		DurationAvgMs:  req.DurationAvgMs,
		DurationMaxMs:  req.DurationMaxMs,
		MemUsedMB:      req.MemUsedMB,
		MemAllocatedMB: req.MemAllocatedMB,
	}
	if err := s.repo.InsertLambdaMetric(ctx, m); err != nil {
		s.logger.Error("failed to ingest Lambda metric", "function_id", req.FunctionID, "error", err)
		return err
	}
	s.logger.Info("Lambda metric ingested", "function_id", req.FunctionID)
	return nil
}

func (s *LambdaService) GetByFunction(ctx context.Context, functionID string) ([]LambdaMetric, error) {
	return s.repo.GetLambdaMetricsByFunction(ctx, functionID)
}

func (s *LambdaService) List(ctx context.Context) ([]LambdaMetric, error) {
	return s.repo.ListLambdaMetrics(ctx)
}

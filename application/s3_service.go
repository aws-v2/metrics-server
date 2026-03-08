package application

import (
	"context"
	"log/slog"
)

// S3Service implements S3MetricsService.
type S3Service struct {
	logger *slog.Logger
	repo   S3MetricsRepository
}

// NewS3Service creates a new S3Service.
func NewS3Service(logger *slog.Logger, repo S3MetricsRepository) *S3Service {
	return &S3Service{logger: logger, repo: repo}
}

func (s *S3Service) Ingest(ctx context.Context, req S3IngestRequest) error {
	m := S3Metric{
		BucketID:         req.BucketID,
		StorageUsedBytes: req.StorageUsedBytes,
		ObjectCount:      req.ObjectCount,
		GetRequests:      req.GetRequests,
		PutRequests:      req.PutRequests,
		BytesDownloaded:  req.BytesDownloaded,
		BytesUploaded:    req.BytesUploaded,
	}
	if err := s.repo.InsertS3Metric(ctx, m); err != nil {
		s.logger.Error("failed to ingest S3 metric", "bucket_id", req.BucketID, "error", err)
		return err
	}
	s.logger.Info("S3 metric ingested", "bucket_id", req.BucketID)
	return nil
}

func (s *S3Service) GetByBucket(ctx context.Context, bucketID string) ([]S3Metric, error) {
	return s.repo.GetS3MetricsByBucket(ctx, bucketID)
}

func (s *S3Service) List(ctx context.Context) ([]S3Metric, error) {
	return s.repo.ListS3Metrics(ctx)
}

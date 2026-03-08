package application

import (
	"context"
	"log/slog"
	"time"
)

// billingService implementation.
type billingService struct {
	logger *slog.Logger
	repo   BillingRepository
}

// NewBillingService creates a new BillingService.
func NewBillingService(logger *slog.Logger, repo BillingRepository) BillingService {
	return &billingService{logger: logger, repo: repo}
}

func (s *billingService) CalculateEC2Usage(ctx context.Context, instanceID string, start, end time.Time) (EC2UsageResponse, error) {
	return s.repo.GetEC2Usage(ctx, instanceID, start, end)
}

func (s *billingService) CalculateRDSUsage(ctx context.Context, instanceID string, start, end time.Time) (RDSUsageResponse, error) {
	return s.repo.GetRDSUsage(ctx, instanceID, start, end)
}

func (s *billingService) CalculateLambdaUsage(ctx context.Context, functionID string, start, end time.Time) (LambdaUsageResponse, error) {
	return s.repo.GetLambdaUsage(ctx, functionID, start, end)
}

func (s *billingService) CalculateS3Usage(ctx context.Context, bucketID string, start, end time.Time) (S3UsageResponse, error) {
	return s.repo.GetS3Usage(ctx, bucketID, start, end)
}

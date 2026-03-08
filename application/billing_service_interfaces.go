package application

import (
	"context"
	"time"
)

// BillingService calculates resource usage for billing purposes over a time window.
type BillingService interface {
	CalculateEC2Usage(ctx context.Context, instanceID string, start, end time.Time) (EC2UsageResponse, error)
	CalculateRDSUsage(ctx context.Context, instanceID string, start, end time.Time) (RDSUsageResponse, error)
	CalculateLambdaUsage(ctx context.Context, functionID string, start, end time.Time) (LambdaUsageResponse, error)
	CalculateS3Usage(ctx context.Context, bucketID string, start, end time.Time) (S3UsageResponse, error)
}

// BillingRepository provides methods to aggregate metrics over a time window for billing.
type BillingRepository interface {
	GetEC2Usage(ctx context.Context, instanceID string, start, end time.Time) (EC2UsageResponse, error)
	GetRDSUsage(ctx context.Context, instanceID string, start, end time.Time) (RDSUsageResponse, error)
	GetLambdaUsage(ctx context.Context, functionID string, start, end time.Time) (LambdaUsageResponse, error)
	GetS3Usage(ctx context.Context, bucketID string, start, end time.Time) (S3UsageResponse, error)
}

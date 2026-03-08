package application

import "context"

// ──────────────────────────────────────────────────────────────────────────────
// Service Interfaces
// ──────────────────────────────────────────────────────────────────────────────

// HealthChecker provides health-check capabilities for the service.
type HealthChecker interface {
	Check(ctx context.Context) (string, error)
}

// EC2MetricsService defines business operations for EC2 metrics.
type EC2MetricsService interface {
	Ingest(ctx context.Context, req EC2IngestRequest) error
	GetByInstance(ctx context.Context, instanceID string) ([]EC2Metric, error)
	List(ctx context.Context) ([]EC2Metric, error)
}

// RDSMetricsService defines business operations for RDS metrics.
type RDSMetricsService interface {
	Ingest(ctx context.Context, req RDSIngestRequest) error
	GetByInstance(ctx context.Context, instanceID string) ([]RDSMetric, error)
	List(ctx context.Context) ([]RDSMetric, error)
}

// LambdaMetricsService defines business operations for Lambda metrics.
type LambdaMetricsService interface {
	Ingest(ctx context.Context, req LambdaIngestRequest) error
	GetByFunction(ctx context.Context, functionID string) ([]LambdaMetric, error)
	List(ctx context.Context) ([]LambdaMetric, error)
}

// S3MetricsService defines business operations for S3 metrics.
type S3MetricsService interface {
	Ingest(ctx context.Context, req S3IngestRequest) error
	GetByBucket(ctx context.Context, bucketID string) ([]S3Metric, error)
	List(ctx context.Context) ([]S3Metric, error)
}

// SageMakerMetricsService defines business operations for SageMaker metrics.
type SageMakerMetricsService interface {
	Ingest(ctx context.Context, req SageMakerIngestRequest) error
	GetByEndpoint(ctx context.Context, endpointID string) ([]SageMakerMetric, error)
	List(ctx context.Context) ([]SageMakerMetric, error)
}

// GameLiftMetricsService defines business operations for GameLift metrics.
type GameLiftMetricsService interface {
	Ingest(ctx context.Context, req GameLiftIngestRequest) error
	GetByFleet(ctx context.Context, fleetID string) ([]GameLiftMetric, error)
	List(ctx context.Context) ([]GameLiftMetric, error)
}

// ──────────────────────────────────────────────────────────────────────────────
// Repository Interfaces
// ──────────────────────────────────────────────────────────────────────────────

// MetricsRepository defines persistence operations for metrics data.
type MetricsRepository interface {
	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error
}

// EC2MetricsRepository defines persistence for EC2 metrics.
type EC2MetricsRepository interface {
	InsertEC2Metric(ctx context.Context, m EC2Metric) error
	GetEC2MetricsByInstance(ctx context.Context, instanceID string) ([]EC2Metric, error)
	ListEC2Metrics(ctx context.Context) ([]EC2Metric, error)
}

// RDSMetricsRepository defines persistence for RDS metrics.
type RDSMetricsRepository interface {
	InsertRDSMetric(ctx context.Context, m RDSMetric) error
	GetRDSMetricsByInstance(ctx context.Context, instanceID string) ([]RDSMetric, error)
	ListRDSMetrics(ctx context.Context) ([]RDSMetric, error)
}

// LambdaMetricsRepository defines persistence for Lambda metrics.
type LambdaMetricsRepository interface {
	InsertLambdaMetric(ctx context.Context, m LambdaMetric) error
	GetLambdaMetricsByFunction(ctx context.Context, functionID string) ([]LambdaMetric, error)
	ListLambdaMetrics(ctx context.Context) ([]LambdaMetric, error)
}

// S3MetricsRepository defines persistence for S3 metrics.
type S3MetricsRepository interface {
	InsertS3Metric(ctx context.Context, m S3Metric) error
	GetS3MetricsByBucket(ctx context.Context, bucketID string) ([]S3Metric, error)
	ListS3Metrics(ctx context.Context) ([]S3Metric, error)
}

// SageMakerMetricsRepository defines persistence for SageMaker metrics.
type SageMakerMetricsRepository interface {
	InsertSageMakerMetric(ctx context.Context, m SageMakerMetric) error
	GetSageMakerMetricsByEndpoint(ctx context.Context, endpointID string) ([]SageMakerMetric, error)
	ListSageMakerMetrics(ctx context.Context) ([]SageMakerMetric, error)
}

// GameLiftMetricsRepository defines persistence for GameLift metrics.
type GameLiftMetricsRepository interface {
	InsertGameLiftMetric(ctx context.Context, m GameLiftMetric) error
	GetGameLiftMetricsByFleet(ctx context.Context, fleetID string) ([]GameLiftMetric, error)
	ListGameLiftMetrics(ctx context.Context) ([]GameLiftMetric, error)
}

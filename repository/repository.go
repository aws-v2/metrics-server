package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"metrics-gateway/application"
)

// PostgresRepository implements all metrics repository interfaces.
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new repository backed by PostgreSQL.
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Ping verifies the database connection is alive.
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// ── EC2 ─────────────────────────────────────────────────────────────────────

func (r *PostgresRepository) InsertEC2Metric(ctx context.Context, m application.EC2Metric) error {
	query := `INSERT INTO ec2_metrics (instance_id, cpu_percent, mem_total_mb, mem_used_mb, mem_percent, disk_total_gb, disk_used_gb)
	           VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		m.InstanceID, m.CPUPercent, m.MemTotalMB, m.MemUsedMB, m.MemPercent, m.DiskTotalGB, m.DiskUsedGB)
	return err
}

func (r *PostgresRepository) GetEC2MetricsByInstance(ctx context.Context, instanceID string) ([]application.EC2Metric, error) {
	var metrics []application.EC2Metric
	query := `SELECT id, instance_id, cpu_percent, mem_total_mb, mem_used_mb, mem_percent, disk_total_gb, disk_used_gb, created_at
	           FROM ec2_metrics WHERE instance_id = $1 ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query, instanceID)
	return metrics, err
}

func (r *PostgresRepository) ListEC2Metrics(ctx context.Context) ([]application.EC2Metric, error) {
	var metrics []application.EC2Metric
	query := `SELECT id, instance_id, cpu_percent, mem_total_mb, mem_used_mb, mem_percent, disk_total_gb, disk_used_gb, created_at
	           FROM ec2_metrics ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query)
	return metrics, err
}

// ── RDS ─────────────────────────────────────────────────────────────────────

func (r *PostgresRepository) InsertRDSMetric(ctx context.Context, m application.RDSMetric) error {
	query := `INSERT INTO rds_metrics (instance_id, cpu_percent, mem_used_mb, mem_percent, storage_used_gb, storage_total_gb, connections, iops)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		m.InstanceID, m.CPUPercent, m.MemUsedMB, m.MemPercent, m.StorageUsedGB, m.StorageTotalGB, m.Connections, m.IOPS)
	return err
}

func (r *PostgresRepository) GetRDSMetricsByInstance(ctx context.Context, instanceID string) ([]application.RDSMetric, error) {
	var metrics []application.RDSMetric
	query := `SELECT id, instance_id, cpu_percent, mem_used_mb, mem_percent, storage_used_gb, storage_total_gb, connections, iops, created_at
	           FROM rds_metrics WHERE instance_id = $1 ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query, instanceID)
	return metrics, err
}

func (r *PostgresRepository) ListRDSMetrics(ctx context.Context) ([]application.RDSMetric, error) {
	var metrics []application.RDSMetric
	query := `SELECT id, instance_id, cpu_percent, mem_used_mb, mem_percent, storage_used_gb, storage_total_gb, connections, iops, created_at
	           FROM rds_metrics ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query)
	return metrics, err
}

// ── Lambda ──────────────────────────────────────────────────────────────────

func (r *PostgresRepository) InsertLambdaMetric(ctx context.Context, m application.LambdaMetric) error {
	query := `INSERT INTO lambda_metrics (function_id, invocations, errors, throttles, duration_avg_ms, duration_max_ms, mem_used_mb, mem_allocated_mb)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		m.FunctionID, m.Invocations, m.Errors, m.Throttles, m.DurationAvgMs, m.DurationMaxMs, m.MemUsedMB, m.MemAllocatedMB)
	return err
}

func (r *PostgresRepository) GetLambdaMetricsByFunction(ctx context.Context, functionID string) ([]application.LambdaMetric, error) {
	var metrics []application.LambdaMetric
	query := `SELECT id, function_id, invocations, errors, throttles, duration_avg_ms, duration_max_ms, mem_used_mb, mem_allocated_mb, created_at
	           FROM lambda_metrics WHERE function_id = $1 ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query, functionID)
	return metrics, err
}

func (r *PostgresRepository) ListLambdaMetrics(ctx context.Context) ([]application.LambdaMetric, error) {
	var metrics []application.LambdaMetric
	query := `SELECT id, function_id, invocations, errors, throttles, duration_avg_ms, duration_max_ms, mem_used_mb, mem_allocated_mb, created_at
	           FROM lambda_metrics ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query)
	return metrics, err
}

// ── S3 ──────────────────────────────────────────────────────────────────────

func (r *PostgresRepository) InsertS3Metric(ctx context.Context, m application.S3Metric) error {
	query := `INSERT INTO s3_metrics (bucket_id, storage_used_bytes, object_count, get_requests, put_requests, bytes_downloaded, bytes_uploaded)
	           VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		m.BucketID, m.StorageUsedBytes, m.ObjectCount, m.GetRequests, m.PutRequests, m.BytesDownloaded, m.BytesUploaded)
	return err
}

func (r *PostgresRepository) GetS3MetricsByBucket(ctx context.Context, bucketID string) ([]application.S3Metric, error) {
	var metrics []application.S3Metric
	query := `SELECT id, bucket_id, storage_used_bytes, object_count, get_requests, put_requests, bytes_downloaded, bytes_uploaded, created_at
	           FROM s3_metrics WHERE bucket_id = $1 ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query, bucketID)
	return metrics, err
}

func (r *PostgresRepository) ListS3Metrics(ctx context.Context) ([]application.S3Metric, error) {
	var metrics []application.S3Metric
	query := `SELECT id, bucket_id, storage_used_bytes, object_count, get_requests, put_requests, bytes_downloaded, bytes_uploaded, created_at
	           FROM s3_metrics ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query)
	return metrics, err
}

// ── SageMaker ───────────────────────────────────────────────────────────────

func (r *PostgresRepository) InsertSageMakerMetric(ctx context.Context, m application.SageMakerMetric) error {
	query := `INSERT INTO sagemaker_metrics (endpoint_id, invocations, model_latency_ms, cpu_percent, mem_percent, gpu_percent, gpu_mem_percent, errors_4xx, errors_5xx)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		m.EndpointID, m.Invocations, m.ModelLatencyMs, m.CPUPercent, m.MemPercent, m.GPUPercent, m.GPUMemPercent, m.Errors4xx, m.Errors5xx)
	return err
}

func (r *PostgresRepository) GetSageMakerMetricsByEndpoint(ctx context.Context, endpointID string) ([]application.SageMakerMetric, error) {
	var metrics []application.SageMakerMetric
	query := `SELECT id, endpoint_id, invocations, model_latency_ms, cpu_percent, mem_percent, gpu_percent, gpu_mem_percent, errors_4xx, errors_5xx, created_at
	           FROM sagemaker_metrics WHERE endpoint_id = $1 ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query, endpointID)
	return metrics, err
}

func (r *PostgresRepository) ListSageMakerMetrics(ctx context.Context) ([]application.SageMakerMetric, error) {
	var metrics []application.SageMakerMetric
	query := `SELECT id, endpoint_id, invocations, model_latency_ms, cpu_percent, mem_percent, gpu_percent, gpu_mem_percent, errors_4xx, errors_5xx, created_at
	           FROM sagemaker_metrics ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query)
	return metrics, err
}

// ── GameLift ────────────────────────────────────────────────────────────────

func (r *PostgresRepository) InsertGameLiftMetric(ctx context.Context, m application.GameLiftMetric) error {
	query := `INSERT INTO gamelift_metrics (fleet_id, active_sessions, active_players, available_sessions, percent_idle_instances, current_instances, desired_instances)
	           VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		m.FleetID, m.ActiveSessions, m.ActivePlayers, m.AvailableSessions, m.PercentIdleInstances, m.CurrentInstances, m.DesiredInstances)
	return err
}

func (r *PostgresRepository) GetGameLiftMetricsByFleet(ctx context.Context, fleetID string) ([]application.GameLiftMetric, error) {
	var metrics []application.GameLiftMetric
	query := `SELECT id, fleet_id, active_sessions, active_players, available_sessions, percent_idle_instances, current_instances, desired_instances, created_at
	           FROM gamelift_metrics WHERE fleet_id = $1 ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query, fleetID)
	return metrics, err
}

func (r *PostgresRepository) ListGameLiftMetrics(ctx context.Context) ([]application.GameLiftMetric, error) {
	var metrics []application.GameLiftMetric
	query := `SELECT id, fleet_id, active_sessions, active_players, available_sessions, percent_idle_instances, current_instances, desired_instances, created_at
	           FROM gamelift_metrics ORDER BY created_at DESC LIMIT 100`
	err := r.db.SelectContext(ctx, &metrics, query)
	return metrics, err
}

package repository

import (
	"context"
	"time"

	"metrics-gateway/application"
)

// GetEC2Usage calculates average EC2 metrics over a given timeframe for billing.
func (r *PostgresRepository) GetEC2Usage(ctx context.Context, instanceID string, start, end time.Time) (application.EC2UsageResponse, error) {
	var res application.EC2UsageResponse
	query := `
		SELECT 
			COALESCE(AVG(cpu_percent), 0) as avg_cpu_percent,
			COALESCE(AVG(mem_used_mb), 0) as avg_mem_used_mb,
			COALESCE(MAX(disk_used_gb), 0) as max_disk_used_gb,
			COUNT(*) as data_points_analyzed
		FROM ec2_metrics 
		WHERE instance_id = $1 AND created_at >= $2 AND created_at <= $3`
	err := r.db.GetContext(ctx, &res, query, instanceID, start, end)
	res.ResourceID = instanceID
	return res, err
}

// GetRDSUsage calculates average RDS metrics over a given timeframe for billing.
func (r *PostgresRepository) GetRDSUsage(ctx context.Context, instanceID string, start, end time.Time) (application.RDSUsageResponse, error) {
	var res application.RDSUsageResponse
	query := `
		SELECT 
			COALESCE(AVG(cpu_percent), 0) as avg_cpu_percent,
			COALESCE(AVG(mem_used_mb), 0) as avg_mem_used_mb,
			COALESCE(MAX(storage_used_gb), 0) as max_storage_used_gb,
			COALESCE(AVG(connections), 0) as avg_connections,
			COUNT(*) as data_points_analyzed
		FROM rds_metrics 
		WHERE instance_id = $1 AND created_at >= $2 AND created_at <= $3`
	err := r.db.GetContext(ctx, &res, query, instanceID, start, end)
	res.ResourceID = instanceID
	return res, err
}

// GetLambdaUsage calculates aggregate Lambda metrics over a given timeframe for billing.
func (r *PostgresRepository) GetLambdaUsage(ctx context.Context, functionID string, start, end time.Time) (application.LambdaUsageResponse, error) {
	var res application.LambdaUsageResponse
	query := `
		SELECT 
			COALESCE(SUM(invocations), 0) as total_invocations,
			COALESCE(SUM(errors), 0) as total_errors,
			COALESCE(AVG(duration_avg_ms), 0) as avg_duration_ms,
			COALESCE(MAX(mem_used_mb), 0) as max_mem_used_mb,
			COUNT(*) as data_points_analyzed
		FROM lambda_metrics 
		WHERE function_id = $1 AND created_at >= $2 AND created_at <= $3`
	err := r.db.GetContext(ctx, &res, query, functionID, start, end)
	res.ResourceID = functionID
	return res, err
}

// GetS3Usage calculates aggregate S3 metrics over a given timeframe for billing.
func (r *PostgresRepository) GetS3Usage(ctx context.Context, bucketID string, start, end time.Time) (application.S3UsageResponse, error) {
	var res application.S3UsageResponse
	query := `
		SELECT 
			COALESCE(MAX(storage_used_bytes), 0) as max_storage_used_bytes,
			COALESCE(SUM(get_requests), 0) as total_get_requests,
			COALESCE(SUM(put_requests), 0) as total_put_requests,
			COALESCE(SUM(bytes_downloaded), 0) as total_bytes_downloaded,
			COALESCE(SUM(bytes_uploaded), 0) as total_bytes_uploaded,
			COUNT(*) as data_points_analyzed
		FROM s3_metrics 
		WHERE bucket_id = $1 AND created_at >= $2 AND created_at <= $3`
	err := r.db.GetContext(ctx, &res, query, bucketID, start, end)
	res.ResourceID = bucketID
	return res, err
}

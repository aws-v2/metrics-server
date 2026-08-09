package application

import "time"

// BillingUsageRequest is the payload received from the billing service via NATS.
type BillingUsageRequest struct {
	Service string `json:"service"`
	Type string `json:"type"`
	ResourceID string `json:"resource_id"`
	Data interface{} `json:"data"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

// EC2UsageResponse contains aggregated EC2 metrics for billing.
type EC2UsageResponse struct {
	ResourceID         string  `json:"resource_id"`
	AvgCPUPercent      float64 `json:"avg_cpu_percent"`
	AvgMemUsedMB       float64 `json:"avg_mem_used_mb"`
	MaxDiskUsedGB      int64   `json:"max_disk_used_gb"`
	DataPointsAnalyzed int     `json:"data_points_analyzed"`
}

// RDSUsageResponse contains aggregated RDS metrics for billing.
type RDSUsageResponse struct {
	ResourceID         string  `json:"resource_id"`
	AvgCPUPercent      float64 `json:"avg_cpu_percent"`
	AvgMemUsedMB       float64 `json:"avg_mem_used_mb"`
	MaxStorageUsedGB   int64   `json:"max_storage_used_gb"`
	AvgConnections     float64 `json:"avg_connections"`
	DataPointsAnalyzed int     `json:"data_points_analyzed"`
}

// LambdaUsageResponse contains aggregated Lambda metrics for billing.
type LambdaUsageResponse struct {
	ResourceID         string  `json:"resource_id"`
	TotalInvocations   int64   `json:"total_invocations"`
	TotalErrors        int64   `json:"total_errors"`
	AvgDurationMs      float64 `json:"avg_duration_ms"`
	MaxMemUsedMB       int64   `json:"max_mem_used_mb"`
	DataPointsAnalyzed int     `json:"data_points_analyzed"`
}

// S3UsageResponse contains aggregated S3 metrics for billing.
type S3UsageResponse struct {
	ResourceID         string `json:"resource_id"`
	MaxStorageUsedByte int64  `json:"max_storage_used_bytes"`
	TotalGetRequests   int64  `json:"total_get_requests"`
	TotalPutRequests   int64  `json:"total_put_requests"`
	TotalBytesDownload int64  `json:"total_bytes_downloaded"`
	TotalBytesUpload   int64  `json:"total_bytes_uploaded"`
	DataPointsAnalyzed int    `json:"data_points_analyzed"`
}

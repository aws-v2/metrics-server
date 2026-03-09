package application

import "time"

// ── EC2 ─────────────────────────────────────────────────────────────────────

// EC2Metric represents a single metrics snapshot from an EC2 instance.
type EC2Metric struct {
	ID          int64     `json:"id"          db:"id"`
	InstanceID  string    `json:"instance_id" db:"instance_id"`
	CPUPercent  float64   `json:"cpu_percent" db:"cpu_percent"`
	MemTotalMB  int64     `json:"mem_total_mb" db:"mem_total_mb"`
	MemUsedMB   int64     `json:"mem_used_mb"  db:"mem_used_mb"`
	MemPercent  float64   `json:"mem_percent"  db:"mem_percent"`
	DiskTotalGB int64     `json:"disk_total_gb" db:"disk_total_gb"`
	DiskUsedGB  int64     `json:"disk_used_gb"  db:"disk_used_gb"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
}

// EC2IngestRequest is the payload the agent POSTs when sending EC2 metrics.
type EC2IngestRequest struct {
	InstanceID  string  `json:"instance_id"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemTotalMB  int64   `json:"mem_total_mb"`
	MemUsedMB   int64   `json:"mem_used_mb"`
	MemPercent  float64 `json:"mem_percent"`
	DiskTotalGB int64   `json:"disk_total_gb"`
	DiskUsedGB  int64   `json:"disk_used_gb"`
}

// ── RDS ─────────────────────────────────────────────────────────────────────

// RDSMetric represents a single metrics snapshot from an RDS database.
type RDSMetric struct {
	ID              int64     `json:"id"               db:"id"`
	InstanceID      string    `json:"instance_id"      db:"instance_id"`
	CPUPercent      float64   `json:"cpu_percent"      db:"cpu_percent"`
	MemUsedMB       int64     `json:"mem_used_mb"      db:"mem_used_mb"`
	MemPercent      float64   `json:"mem_percent"      db:"mem_percent"`
	StorageUsedGB   int64     `json:"storage_used_gb"  db:"storage_used_gb"`
	StorageTotalGB  int64     `json:"storage_total_gb" db:"storage_total_gb"`
	Connections     int64     `json:"connections"      db:"connections"`
	IOPS            int64     `json:"iops"             db:"iops"`
	CreatedAt       time.Time `json:"created_at"       db:"created_at"`
}

// RDSIngestRequest is the payload for ingesting RDS metrics.
type RDSIngestRequest struct {
	InstanceID     string  `json:"instance_id"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemUsedMB      int64   `json:"mem_used_mb"`
	MemPercent     float64 `json:"mem_percent"`
	StorageUsedGB  int64   `json:"storage_used_gb"`
	StorageTotalGB int64   `json:"storage_total_gb"`
	Connections    int64   `json:"connections"`
	IOPS           int64   `json:"iops"`
}

// ── Lambda ──────────────────────────────────────────────────────────────────

// LambdaMetric represents a single metrics snapshot from a Lambda function.
type LambdaMetric struct {
	ID             int64     `json:"id"              db:"id"`
	FunctionID     string    `json:"function_id"     db:"function_id"`
	Invocations    int64     `json:"invocations"     db:"invocations"`
	Errors         int64     `json:"errors"          db:"errors"`
	Throttles      int64     `json:"throttles"       db:"throttles"`
	DurationAvgMs  float64   `json:"duration_avg_ms" db:"duration_avg_ms"`
	DurationMaxMs  float64   `json:"duration_max_ms" db:"duration_max_ms"`
	MemUsedMB      int64     `json:"mem_used_mb"     db:"mem_used_mb"`
	MemAllocatedMB int64     `json:"mem_allocated_mb" db:"mem_allocated_mb"`
	CreatedAt      time.Time `json:"created_at"      db:"created_at"`
}

// LambdaIngestRequest is the payload for ingesting Lambda metrics.
type LambdaIngestRequest struct {
	FunctionID     string  `json:"function_id"`
	Invocations    int64   `json:"invocations"`
	Errors         int64   `json:"errors"`
	Throttles      int64   `json:"throttles"`
	DurationAvgMs  float64 `json:"duration_avg_ms"`
	DurationMaxMs  float64 `json:"duration_max_ms"`
	MemUsedMB      int64   `json:"mem_used_mb"`
	MemAllocatedMB int64   `json:"mem_allocated_mb"`
}

// ── S3 ──────────────────────────────────────────────────────────────────────
// S3Metric represents a single metrics snapshot from an S3 bucket.
type S3Metric struct {
	ID               int64     `json:"id"                 db:"id"`
	BucketID         string    `json:"bucket_id"          db:"bucket_id"`
	OwnerID          string    `json:"owner_id"           db:"owner_id"` // 👈 Crucial for billing
	Region           string    `json:"region"             db:"region"`   // 👈 Crucial for regional pricing
	
	// Capacity Metrics
	StorageUsedBytes int64     `json:"storage_used_bytes"  db:"storage_used_bytes"`
	ObjectCount      int64     `json:"object_count"       db:"object_count"`
	
	// Operation Count Metrics (Granular Tier-based)
	GetRequests    int64 `json:"get_requests"    db:"get_requests"`    // Tier 2
	PutRequests    int64 `json:"put_requests"    db:"put_requests"`    // Tier 1
	ListRequests   int64 `json:"list_requests"   db:"list_requests"`   // Tier 2
	DeleteRequests int64 `json:"delete_requests" db:"delete_requests"` // Free/Tier 1 depending on policy
	HeadRequests   int64 `json:"head_requests"   db:"head_requests"`   // Tier 2
	
	// Bandwidth Metrics (The "Flow")
	BytesDownloaded int64 `json:"bytes_downloaded" db:"bytes_downloaded"`
	BytesUploaded   int64 `json:"bytes_uploaded"   db:"bytes_uploaded"`
	
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// S3IngestRequest is the payload for ingesting S3 metrics.
type S3IngestRequest struct {
	BucketID         string `json:"bucket_id"`
	OwnerID          string `json:"owner_id"`
	Region           string `json:"region"`
	
	StorageUsedBytes int64 `json:"storage_used_bytes"`
	ObjectCount      int64 `json:"object_count"`
	
	GetRequests    int64 `json:"get_requests"`
	PutRequests    int64 `json:"put_requests"`
	ListRequests   int64 `json:"list_requests"`
	DeleteRequests int64 `json:"delete_requests"`
	HeadRequests   int64 `json:"head_requests"`
	
	BytesDownloaded int64 `json:"bytes_downloaded"`
	BytesUploaded   int64 `json:"bytes_uploaded"`
}

// ── SageMaker ───────────────────────────────────────────────────────────────

// SageMakerMetric represents a single metrics snapshot from a SageMaker endpoint.
type SageMakerMetric struct {
	ID              int64     `json:"id"               db:"id"`
	EndpointID      string    `json:"endpoint_id"      db:"endpoint_id"`
	Invocations     int64     `json:"invocations"      db:"invocations"`
	ModelLatencyMs  float64   `json:"model_latency_ms" db:"model_latency_ms"`
	CPUPercent      float64   `json:"cpu_percent"      db:"cpu_percent"`
	MemPercent      float64   `json:"mem_percent"      db:"mem_percent"`
	GPUPercent      float64   `json:"gpu_percent"      db:"gpu_percent"`
	GPUMemPercent   float64   `json:"gpu_mem_percent"  db:"gpu_mem_percent"`
	Errors4xx       int64     `json:"errors_4xx"       db:"errors_4xx"`
	Errors5xx       int64     `json:"errors_5xx"       db:"errors_5xx"`
	CreatedAt       time.Time `json:"created_at"       db:"created_at"`
}

// SageMakerIngestRequest is the payload for ingesting SageMaker metrics.
type SageMakerIngestRequest struct {
	EndpointID     string  `json:"endpoint_id"`
	Invocations    int64   `json:"invocations"`
	ModelLatencyMs float64 `json:"model_latency_ms"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemPercent     float64 `json:"mem_percent"`
	GPUPercent     float64 `json:"gpu_percent"`
	GPUMemPercent  float64 `json:"gpu_mem_percent"`
	Errors4xx      int64   `json:"errors_4xx"`
	Errors5xx      int64   `json:"errors_5xx"`
}

// ── GameLift ────────────────────────────────────────────────────────────────

// GameLiftMetric represents a single metrics snapshot from a GameLift fleet.
type GameLiftMetric struct {
	ID                int64     `json:"id"                 db:"id"`
	FleetID           string    `json:"fleet_id"           db:"fleet_id"`
	ActiveSessions    int64     `json:"active_sessions"    db:"active_sessions"`
	ActivePlayers     int64     `json:"active_players"     db:"active_players"`
	AvailableSessions int64     `json:"available_sessions" db:"available_sessions"`
	PercentIdleInstances float64 `json:"percent_idle_instances" db:"percent_idle_instances"`
	CurrentInstances  int64     `json:"current_instances"  db:"current_instances"`
	DesiredInstances  int64     `json:"desired_instances"  db:"desired_instances"`
	CreatedAt         time.Time `json:"created_at"         db:"created_at"`
}

// GameLiftIngestRequest is the payload for ingesting GameLift metrics.
type GameLiftIngestRequest struct {
	FleetID              string  `json:"fleet_id"`
	ActiveSessions       int64   `json:"active_sessions"`
	ActivePlayers        int64   `json:"active_players"`
	AvailableSessions    int64   `json:"available_sessions"`
	PercentIdleInstances float64 `json:"percent_idle_instances"`
	CurrentInstances     int64   `json:"current_instances"`
	DesiredInstances     int64   `json:"desired_instances"`
}

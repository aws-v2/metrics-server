package postgres

// Schema defines the initial database tables for the metrics-gateway service.
// It is executed via database.Migrate on startup.
const Schema = `
-- ── EC2 Metrics ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ec2_metrics (
    id            BIGSERIAL        PRIMARY KEY,
    instance_id   TEXT             NOT NULL,
    cpu_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_total_mb  BIGINT           NOT NULL DEFAULT 0,
    mem_used_mb   BIGINT           NOT NULL DEFAULT 0,
    mem_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_total_gb BIGINT           NOT NULL DEFAULT 0,
    disk_used_gb  BIGINT           NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ec2_metrics_instance_id ON ec2_metrics (instance_id);
CREATE INDEX IF NOT EXISTS idx_ec2_metrics_created_at  ON ec2_metrics (created_at);

-- ── EC2 Scaling Policies ────────────────────────────────────────────────────
DROP TABLE IF EXISTS ec2_scaling_policies;
CREATE TABLE IF NOT EXISTS ec2_scaling_policies (
    id                 TEXT             PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id          TEXT             NOT NULL,
    target_type        TEXT             NOT NULL,
    target_id          TEXT             NOT NULL,
    metric_name        TEXT             NOT NULL,
    target_value       DOUBLE PRECISION NOT NULL,
    scale_down_value   DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_instances      BIGINT           NOT NULL DEFAULT 5,
    scale_out_cooldown BIGINT           NOT NULL DEFAULT 300,
    scale_in_cooldown  BIGINT           NOT NULL DEFAULT 300,
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE(target_id, metric_name)
);

CREATE INDEX IF NOT EXISTS idx_ec2_scaling_policies_tenant_id ON ec2_scaling_policies (tenant_id);

-- ── RDS Metrics ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS rds_metrics (
    id               BIGSERIAL        PRIMARY KEY,
    instance_id      TEXT             NOT NULL,
    cpu_percent      DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_used_mb      BIGINT           NOT NULL DEFAULT 0,
    mem_percent      DOUBLE PRECISION NOT NULL DEFAULT 0,
    storage_used_gb  BIGINT           NOT NULL DEFAULT 0,
    storage_total_gb BIGINT           NOT NULL DEFAULT 0,
    connections      BIGINT           NOT NULL DEFAULT 0,
    iops             BIGINT           NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rds_metrics_instance_id ON rds_metrics (instance_id);
CREATE INDEX IF NOT EXISTS idx_rds_metrics_created_at  ON rds_metrics (created_at);

-- ── Lambda Metrics ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS lambda_metrics (
    id               BIGSERIAL        PRIMARY KEY,
    function_id      TEXT             NOT NULL,
    invocations      BIGINT           NOT NULL DEFAULT 0,
    errors           BIGINT           NOT NULL DEFAULT 0,
    throttles        BIGINT           NOT NULL DEFAULT 0,
    duration_avg_ms  DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration_max_ms  DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_used_mb      BIGINT           NOT NULL DEFAULT 0,
    mem_allocated_mb BIGINT           NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lambda_metrics_function_id ON lambda_metrics (function_id);
CREATE INDEX IF NOT EXISTS idx_lambda_metrics_created_at  ON lambda_metrics (created_at);

-- ── S3 Metrics ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS s3_metrics (
    id                 BIGSERIAL   PRIMARY KEY,
    bucket_id          TEXT        NOT NULL,
    storage_used_bytes BIGINT      NOT NULL DEFAULT 0,
    object_count       BIGINT      NOT NULL DEFAULT 0,
    get_requests       BIGINT      NOT NULL DEFAULT 0,
    put_requests       BIGINT      NOT NULL DEFAULT 0,
    bytes_downloaded   BIGINT      NOT NULL DEFAULT 0,
    bytes_uploaded     BIGINT      NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_s3_metrics_bucket_id   ON s3_metrics (bucket_id);
CREATE INDEX IF NOT EXISTS idx_s3_metrics_created_at  ON s3_metrics (created_at);

-- ── SageMaker Metrics ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sagemaker_metrics (
    id               BIGSERIAL        PRIMARY KEY,
    endpoint_id      TEXT             NOT NULL,
    invocations      BIGINT           NOT NULL DEFAULT 0,
    model_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    cpu_percent      DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_percent      DOUBLE PRECISION NOT NULL DEFAULT 0,
    gpu_percent      DOUBLE PRECISION NOT NULL DEFAULT 0,
    gpu_mem_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    errors_4xx       BIGINT           NOT NULL DEFAULT 0,
    errors_5xx       BIGINT           NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sagemaker_metrics_endpoint_id ON sagemaker_metrics (endpoint_id);
CREATE INDEX IF NOT EXISTS idx_sagemaker_metrics_created_at  ON sagemaker_metrics (created_at);

-- ── GameLift Metrics ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS gamelift_metrics (
    id                     BIGSERIAL        PRIMARY KEY,
    fleet_id               TEXT             NOT NULL,
    active_sessions        BIGINT           NOT NULL DEFAULT 0,
    active_players         BIGINT           NOT NULL DEFAULT 0,
    available_sessions     BIGINT           NOT NULL DEFAULT 0,
    percent_idle_instances DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_instances      BIGINT           NOT NULL DEFAULT 0,
    desired_instances      BIGINT           NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gamelift_metrics_fleet_id   ON gamelift_metrics (fleet_id);
CREATE INDEX IF NOT EXISTS idx_gamelift_metrics_created_at ON gamelift_metrics (created_at);
`

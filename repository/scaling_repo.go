package repository

import (
	"context"
)

// EC2ScaleCandidate represents an EC2 instance that breached its dynamic thresholds.
type EC2ScaleCandidate struct {
	InstanceID    string  `db:"instance_id"`
	AvgCPUPercent float64 `db:"avg_cpu"`
	TenantID      string  `db:"tenant_id"`
	TargetType    string  `db:"target_type"`
	TargetID      string  `db:"target_id"`
	MetricName    string  `db:"metric_name"`
	TargetValue   float64 `db:"target_value"`
}

// EC2ScalingPolicy defines the scaling rule matching the NATS payload.
type EC2ScalingPolicy struct {
	ID               string  `json:"id" db:"id"`
	TenantID         string  `json:"tenant_id" db:"tenant_id"`
	TargetType       string  `json:"target_type" db:"target_type"`
	TargetID         string  `json:"target_id" db:"target_id"`
	MetricName       string  `json:"metric_name" db:"metric_name"`
	TargetValue      float64 `json:"target_value" db:"target_value"`
	ScaleDownValue   float64 `json:"scale_down_value" db:"scale_down_value"`
	MaxInstances     int     `json:"max_instances" db:"max_instances"`
	ScaleOutCooldown int     `json:"scale_out_cooldown" db:"scale_out_cooldown"`
	ScaleInCooldown  int     `json:"scale_in_cooldown" db:"scale_in_cooldown"`
	CreatedAt        string  `json:"created_at" db:"created_at"`
	UpdatedAt        string  `json:"updated_at" db:"updated_at"`
}

// CreateEC2ScalingPolicy inserts or updates a dynamic scaling rule for an instance or ASg.
func (r *PostgresRepository) CreateEC2ScalingPolicy(ctx context.Context, p EC2ScalingPolicy) error {
	query := `
		INSERT INTO ec2_scaling_policies (tenant_id, target_type, target_id, metric_name, target_value, scale_down_value, max_instances, scale_out_cooldown, scale_in_cooldown, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (target_id, metric_name) 
		DO UPDATE SET 
			target_value = EXCLUDED.target_value, 
			scale_down_value = EXCLUDED.scale_down_value,
			max_instances = EXCLUDED.max_instances,
			scale_out_cooldown = EXCLUDED.scale_out_cooldown,
			scale_in_cooldown = EXCLUDED.scale_in_cooldown,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, p.TenantID, p.TargetType, p.TargetID, p.MetricName, p.TargetValue, p.ScaleDownValue, p.MaxInstances, p.ScaleOutCooldown, p.ScaleInCooldown)
	return err
}

// UpdateEC2ScalingPolicy updates an existing policy.
func (r *PostgresRepository) UpdateEC2ScalingPolicy(ctx context.Context, tenantID, policyID string, targetValue, scaleDownValue float64, maxInstances, scaleOut, scaleIn int) error {
	query := `
		UPDATE ec2_scaling_policies 
		SET target_value = $1, scale_down_value = $2, max_instances = $3, scale_out_cooldown = $4, scale_in_cooldown = $5, updated_at = NOW()
		WHERE id = $6 AND tenant_id = $7
	`
	_, err := r.db.ExecContext(ctx, query, targetValue, scaleDownValue, maxInstances, scaleOut, scaleIn, policyID, tenantID)
	return err
}

// DeleteEC2ScalingPolicy deletes a policy.
func (r *PostgresRepository) DeleteEC2ScalingPolicy(ctx context.Context, tenantID, policyID string) error {
	query := `
		DELETE FROM ec2_scaling_policies 
		WHERE id = $1 AND tenant_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, policyID, tenantID)
	return err
}

// ListEC2ScalingPolicies lists policies for a given tenant.
func (r *PostgresRepository) ListEC2ScalingPolicies(ctx context.Context, tenantID string) ([]EC2ScalingPolicy, error) {
	var policies []EC2ScalingPolicy
	query := `
		SELECT id, tenant_id, target_type, target_id, metric_name, target_value, scale_down_value, max_instances, scale_out_cooldown, scale_in_cooldown, created_at, updated_at
		FROM ec2_scaling_policies
		WHERE tenant_id = $1
	`
	err := r.db.SelectContext(ctx, &policies, query, tenantID)
	if policies == nil {
		policies = []EC2ScalingPolicy{}
	}
	return policies, err
}

// GetEC2InstancesToScaleUp finds instances where avg CPU > dynamic policy target over the last 3 data points.
func (r *PostgresRepository) GetEC2InstancesToScaleUp(ctx context.Context) ([]EC2ScaleCandidate, error) {
	var candidates []EC2ScaleCandidate
	query := `
		SELECT 
			m.instance_id, 
			AVG(m.cpu_percent) as avg_cpu,
			p.tenant_id,
			p.target_type,
			p.target_id,
			p.metric_name,
			p.target_value
		FROM (
			SELECT instance_id, cpu_percent,
				ROW_NUMBER() OVER(PARTITION BY instance_id ORDER BY created_at DESC) as rn
			FROM ec2_metrics
		) m
		JOIN ec2_scaling_policies p ON m.instance_id = p.target_id
		WHERE m.rn <= 3 AND p.metric_name = 'CPUUtilization' AND p.target_type = 'instance'
		GROUP BY m.instance_id, p.tenant_id, p.target_type, p.target_id, p.metric_name, p.target_value
		HAVING AVG(m.cpu_percent) > p.target_value AND COUNT(m.*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// EC2ScaleDownCandidate represents an EC2 instance that breached its dynamic lower thresholds.
type EC2ScaleDownCandidate struct {
	InstanceID     string  `db:"instance_id"`
	AvgCPUPercent  float64 `db:"avg_cpu"`
	TenantID       string  `db:"tenant_id"`
	TargetType     string  `db:"target_type"`
	TargetID       string  `db:"target_id"`
	MetricName     string  `db:"metric_name"`
	ScaleDownValue float64 `db:"scale_down_value"`
}

// GetEC2InstancesToScaleDown finds instances where avg CPU < dynamic policy lower target over the last 3 data points.
func (r *PostgresRepository) GetEC2InstancesToScaleDown(ctx context.Context) ([]EC2ScaleDownCandidate, error) {
	var candidates []EC2ScaleDownCandidate
	query := `
		SELECT 
			m.instance_id, 
			AVG(m.cpu_percent) as avg_cpu,
			p.tenant_id,
			p.target_type,
			p.target_id,
			p.metric_name,
			p.scale_down_value
		FROM (
			SELECT instance_id, cpu_percent,
				ROW_NUMBER() OVER(PARTITION BY instance_id ORDER BY created_at DESC) as rn
			FROM ec2_metrics
		) m
		JOIN ec2_scaling_policies p ON m.instance_id = p.target_id
		WHERE m.rn <= 3 AND p.metric_name = 'CPUUtilization' AND p.target_type = 'instance'
		GROUP BY m.instance_id, p.tenant_id, p.target_type, p.target_id, p.metric_name, p.scale_down_value
		HAVING AVG(m.cpu_percent) < p.scale_down_value AND COUNT(m.*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// RDSScaleCandidate represents an RDS instance that breached thresholds.
type RDSScaleCandidate struct {
	InstanceID    string  `db:"instance_id"`
	AvgCPUPercent float64 `db:"avg_cpu"`
}

// GetRDSInstancesToScaleUp finds RDS instances where avg CPU > 80% over the last 3 data points.
func (r *PostgresRepository) GetRDSInstancesToScaleUp(ctx context.Context) ([]RDSScaleCandidate, error) {
	var candidates []RDSScaleCandidate
	query := `
		SELECT instance_id, AVG(cpu_percent) as avg_cpu
		FROM (
			SELECT instance_id, cpu_percent,
				ROW_NUMBER() OVER(PARTITION BY instance_id ORDER BY created_at DESC) as rn
			FROM rds_metrics
		) tmp
		WHERE rn <= 3
		GROUP BY instance_id
		HAVING AVG(cpu_percent) > 80.0 AND COUNT(*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// RDSScalingPolicy defines the vertical scaling rule for an RDS container.
type RDSScalingPolicy struct {
	ID                 string  `json:"id" db:"id"`
	TenantID           string  `json:"tenant_id" db:"tenant_id"`
	InstanceID         string  `json:"instance_id" db:"instance_id"`
	MetricName         string  `json:"metric_name" db:"metric_name"`
	ScaleUpThreshold   float64 `json:"scale_up_threshold" db:"scale_up_threshold"`
	ScaleDownThreshold float64 `json:"scale_down_threshold" db:"scale_down_threshold"`
	MaxLimit           float64 `json:"max_limit" db:"max_limit"`
	MinLimit           float64 `json:"min_limit" db:"min_limit"`
	ScaleStep          float64 `json:"scale_step" db:"scale_step"`
	CooldownSeconds    int64   `json:"cooldown_seconds" db:"cooldown_seconds"`
	CreatedAt          string  `json:"created_at" db:"created_at"`
	UpdatedAt          string  `json:"updated_at" db:"updated_at"`
}

// RDSVerticalScaleCandidate represents an RDS instance that breached its dynamic thresholds.
type RDSVerticalScaleCandidate struct {
	InstanceID       string  `db:"instance_id"`
	AvgMetricValue   float64 `db:"avg_metric"`
	TenantID         string  `db:"tenant_id"`
	MetricName       string  `db:"metric_name"`
	ScaleUpThreshold float64 `db:"scale_up_threshold"`
	ScaleDownThreshold float64 `db:"scale_down_threshold"`
}

func (r *PostgresRepository) CreateRDSScalingPolicy(ctx context.Context, p RDSScalingPolicy) error {
	query := `
		INSERT INTO rds_scaling_policies (tenant_id, instance_id, metric_name, scale_up_threshold, scale_down_threshold, max_limit, min_limit, scale_step, cooldown_seconds, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (instance_id, metric_name) 
		DO UPDATE SET 
			scale_up_threshold = EXCLUDED.scale_up_threshold,
			scale_down_threshold = EXCLUDED.scale_down_threshold,
			max_limit = EXCLUDED.max_limit,
			min_limit = EXCLUDED.min_limit,
			scale_step = EXCLUDED.scale_step,
			cooldown_seconds = EXCLUDED.cooldown_seconds,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, p.TenantID, p.InstanceID, p.MetricName, p.ScaleUpThreshold, p.ScaleDownThreshold, p.MaxLimit, p.MinLimit, p.ScaleStep, p.CooldownSeconds)
	return err
}

func (r *PostgresRepository) UpdateRDSScalingPolicy(ctx context.Context, tenantID, instanceID, metricName string, up, down float64, max, min, step float64, cooldown int64) error {
	query := `
		UPDATE rds_scaling_policies 
		SET scale_up_threshold = $1, scale_down_threshold = $2, max_limit = $3, min_limit = $4, scale_step = $5, cooldown_seconds = $6, updated_at = NOW()
		WHERE instance_id = $7 AND tenant_id = $8 AND metric_name = $9
	`
	_, err := r.db.ExecContext(ctx, query, up, down, max, min, step, cooldown, instanceID, tenantID, metricName)
	return err
}

func (r *PostgresRepository) DeleteRDSScalingPolicy(ctx context.Context, tenantID, instanceID, metricName string) error {
	query := `
		DELETE FROM rds_scaling_policies 
		WHERE instance_id = $1 AND tenant_id = $2 AND metric_name = $3
	`
	_, err := r.db.ExecContext(ctx, query, instanceID, tenantID, metricName)
	return err
}

func (r *PostgresRepository) ListRDSScalingPolicies(ctx context.Context, tenantID string) ([]RDSScalingPolicy, error) {
	var policies []RDSScalingPolicy
	query := `
		SELECT id, tenant_id, instance_id, metric_name, scale_up_threshold, scale_down_threshold, max_limit, min_limit, scale_step, cooldown_seconds, created_at, updated_at
		FROM rds_scaling_policies
		WHERE tenant_id = $1
	`
	err := r.db.SelectContext(ctx, &policies, query, tenantID)
	if policies == nil {
		policies = []RDSScalingPolicy{}
	}
	return policies, err
}

// GetRDSVerticalScaleUpCandidates finds instances where avg metric > scale_up_threshold over the last 3 data points.
func (r *PostgresRepository) GetRDSVerticalScaleUpCandidates(ctx context.Context) ([]RDSVerticalScaleCandidate, error) {
	var candidates []RDSVerticalScaleCandidate
	query := `
		SELECT 
			m.instance_id, 
			AVG(CASE 
				WHEN p.metric_name = 'CPUUtilization' THEN m.cpu_percent 
				WHEN p.metric_name = 'MemoryUtilization' THEN m.mem_percent 
				ELSE 0 
			END) as avg_metric,
			p.tenant_id,
			p.metric_name,
			p.scale_up_threshold,
			p.scale_down_threshold
		FROM (
			SELECT instance_id, cpu_percent, mem_percent,
				ROW_NUMBER() OVER(PARTITION BY instance_id ORDER BY created_at DESC) as rn
			FROM rds_metrics
		) m
		JOIN rds_scaling_policies p ON m.instance_id = p.instance_id
		WHERE m.rn <= 3
		GROUP BY m.instance_id, p.tenant_id, p.metric_name, p.scale_up_threshold, p.scale_down_threshold
		HAVING AVG(CASE 
				WHEN p.metric_name = 'CPUUtilization' THEN m.cpu_percent 
				WHEN p.metric_name = 'MemoryUtilization' THEN m.mem_percent 
				ELSE 0 
			END) > p.scale_up_threshold AND COUNT(m.*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// GetRDSVerticalScaleDownCandidates finds instances where avg metric < scale_down_threshold over the last 3 data points.
func (r *PostgresRepository) GetRDSVerticalScaleDownCandidates(ctx context.Context) ([]RDSVerticalScaleCandidate, error) {
	var candidates []RDSVerticalScaleCandidate
	query := `
		SELECT 
			m.instance_id, 
			AVG(CASE 
				WHEN p.metric_name = 'CPUUtilization' THEN m.cpu_percent 
				WHEN p.metric_name = 'MemoryUtilization' THEN m.mem_percent 
				ELSE 0 
			END) as avg_metric,
			p.tenant_id,
			p.metric_name,
			p.scale_up_threshold,
			p.scale_down_threshold
		FROM (
			SELECT instance_id, cpu_percent, mem_percent,
				ROW_NUMBER() OVER(PARTITION BY instance_id ORDER BY created_at DESC) as rn
			FROM rds_metrics
		) m
		JOIN rds_scaling_policies p ON m.instance_id = p.instance_id
		WHERE m.rn <= 3
		GROUP BY m.instance_id, p.tenant_id, p.metric_name, p.scale_up_threshold, p.scale_down_threshold
		HAVING AVG(CASE 
				WHEN p.metric_name = 'CPUUtilization' THEN m.cpu_percent 
				WHEN p.metric_name = 'MemoryUtilization' THEN m.mem_percent 
				ELSE 0 
			END) < p.scale_down_threshold AND COUNT(m.*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// LambdaScalingPolicy defines the scaling rule for a Lambda function.
type LambdaScalingPolicy struct {
	ID                   string  `json:"id" db:"id"`
	TenantID             string  `json:"tenant_id" db:"tenant_id"`
	FunctionID           string  `json:"function_id" db:"function_id"`
	MetricName           string  `json:"metric_name" db:"metric_name"`
	ScaleUpThreshold     float64 `json:"scale_up_threshold" db:"scale_up_threshold"`
	ScaleDownThreshold   float64 `json:"scale_down_threshold" db:"scale_down_threshold"`
	MaxConcurrencyLimit  int64   `json:"max_concurrency_limit" db:"max_concurrency_limit"`
	MinConcurrencyLimit  int64   `json:"min_concurrency_limit" db:"min_concurrency_limit"`
	ScaleStep            int64   `json:"scale_step" db:"scale_step"`
	CooldownSeconds      int64   `json:"cooldown_seconds" db:"cooldown_seconds"`
	CreatedAt            string  `json:"created_at" db:"created_at"`
	UpdatedAt            string  `json:"updated_at" db:"updated_at"`
}

func (r *PostgresRepository) CreateLambdaScalingPolicy(ctx context.Context, p LambdaScalingPolicy) error {
	query := `
		INSERT INTO lambda_scaling_policies (tenant_id, function_id, metric_name, scale_up_threshold, scale_down_threshold, max_concurrency_limit, min_concurrency_limit, scale_step, cooldown_seconds, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (function_id, metric_name) 
		DO UPDATE SET 
			scale_up_threshold = EXCLUDED.scale_up_threshold,
			scale_down_threshold = EXCLUDED.scale_down_threshold,
			max_concurrency_limit = EXCLUDED.max_concurrency_limit,
			min_concurrency_limit = EXCLUDED.min_concurrency_limit,
			scale_step = EXCLUDED.scale_step,
			cooldown_seconds = EXCLUDED.cooldown_seconds,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, p.TenantID, p.FunctionID, p.MetricName, p.ScaleUpThreshold, p.ScaleDownThreshold, p.MaxConcurrencyLimit, p.MinConcurrencyLimit, p.ScaleStep, p.CooldownSeconds)
	return err
}

func (r *PostgresRepository) UpdateLambdaScalingPolicy(ctx context.Context, tenantID, functionID, metricName string, up, down float64, max, min, step int64, cooldown int64) error {
	query := `
		UPDATE lambda_scaling_policies 
		SET scale_up_threshold = $1, scale_down_threshold = $2, max_concurrency_limit = $3, min_concurrency_limit = $4, scale_step = $5, cooldown_seconds = $6, updated_at = NOW()
		WHERE function_id = $7 AND tenant_id = $8 AND metric_name = $9
	`
	_, err := r.db.ExecContext(ctx, query, up, down, max, min, step, cooldown, functionID, tenantID, metricName)
	return err
}

func (r *PostgresRepository) DeleteLambdaScalingPolicy(ctx context.Context, tenantID, functionID, metricName string) error {
	query := `
		DELETE FROM lambda_scaling_policies 
		WHERE function_id = $1 AND tenant_id = $2 AND metric_name = $3
	`
	_, err := r.db.ExecContext(ctx, query, functionID, tenantID, metricName)
	return err
}

func (r *PostgresRepository) ListLambdaScalingPolicies(ctx context.Context, tenantID string) ([]LambdaScalingPolicy, error) {
	var policies []LambdaScalingPolicy
	query := `
		SELECT id, tenant_id, function_id, metric_name, scale_up_threshold, scale_down_threshold, max_concurrency_limit, min_concurrency_limit, scale_step, cooldown_seconds, created_at, updated_at
		FROM lambda_scaling_policies
		WHERE tenant_id = $1
	`
	err := r.db.SelectContext(ctx, &policies, query, tenantID)
	if policies == nil {
		policies = []LambdaScalingPolicy{}
	}
	return policies, err
}

// LambdaScaleCandidate represents a Lambda function that breached its dynamic thresholds.
type LambdaScaleCandidate struct {
	FunctionID         string  `db:"function_id"`
	AvgMetricValue     float64 `db:"avg_metric"`
	TenantID           string  `db:"tenant_id"`
	MetricName         string  `db:"metric_name"`
	ScaleUpThreshold   float64 `db:"scale_up_threshold"`
	ScaleDownThreshold float64 `db:"scale_down_threshold"`
}

// GetLambdaFunctionsToScaleUp finds Lambda functions where metric > scale_up_threshold over the last 3 data points.
func (r *PostgresRepository) GetLambdaFunctionsToScaleUp(ctx context.Context) ([]LambdaScaleCandidate, error) {
	var candidates []LambdaScaleCandidate
	query := `
		SELECT 
			m.function_id, 
			AVG(CASE 
				WHEN p.metric_name = 'throttles' THEN m.throttles 
				WHEN p.metric_name = 'invocations' THEN m.invocations
				WHEN p.metric_name = 'duration_avg_ms' THEN m.duration_avg_ms
				ELSE 0 
			END) as avg_metric,
			p.tenant_id,
			p.metric_name,
			p.scale_up_threshold,
			p.scale_down_threshold
		FROM (
			SELECT function_id, throttles, invocations, duration_avg_ms,
				ROW_NUMBER() OVER(PARTITION BY function_id ORDER BY created_at DESC) as rn
			FROM lambda_metrics
		) m
		JOIN lambda_scaling_policies p ON m.function_id = p.function_id
		WHERE m.rn <= 3
		GROUP BY m.function_id, p.tenant_id, p.metric_name, p.scale_up_threshold, p.scale_down_threshold
		HAVING AVG(CASE 
				WHEN p.metric_name = 'throttles' THEN m.throttles 
				WHEN p.metric_name = 'invocations' THEN m.invocations
				WHEN p.metric_name = 'duration_avg_ms' THEN m.duration_avg_ms
				ELSE 0 
			END) > p.scale_up_threshold AND COUNT(m.*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// GetLambdaFunctionsToScaleDown finds Lambda functions where metric < scale_down_threshold over the last 3 data points.
func (r *PostgresRepository) GetLambdaFunctionsToScaleDown(ctx context.Context) ([]LambdaScaleCandidate, error) {
	var candidates []LambdaScaleCandidate
	query := `
		SELECT 
			m.function_id, 
			AVG(CASE 
				WHEN p.metric_name = 'throttles' THEN m.throttles 
				WHEN p.metric_name = 'invocations' THEN m.invocations
				WHEN p.metric_name = 'duration_avg_ms' THEN m.duration_avg_ms
				ELSE 0 
			END) as avg_metric,
			p.tenant_id,
			p.metric_name,
			p.scale_up_threshold,
			p.scale_down_threshold
		FROM (
			SELECT function_id, throttles, invocations, duration_avg_ms,
				ROW_NUMBER() OVER(PARTITION BY function_id ORDER BY created_at DESC) as rn
			FROM lambda_metrics
		) m
		JOIN lambda_scaling_policies p ON m.function_id = p.function_id
		WHERE m.rn <= 3
		GROUP BY m.function_id, p.tenant_id, p.metric_name, p.scale_up_threshold, p.scale_down_threshold
		HAVING AVG(CASE 
				WHEN p.metric_name = 'throttles' THEN m.throttles 
				WHEN p.metric_name = 'invocations' THEN m.invocations
				WHEN p.metric_name = 'duration_avg_ms' THEN m.duration_avg_ms
				ELSE 0 
			END) < p.scale_down_threshold AND COUNT(m.*) >= 3
	`
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}

// S3ScaleCandidate represents an S3 bucket that breached thresholds.
type S3ScaleCandidate struct {
	BucketID         string `db:"bucket_id"`
	StorageUsedBytes int64  `db:"max_storage"`
}

// GetS3BucketsToScaleUp finds S3 buckets nearing extreme storage limits (e.g. > 100GB).
func (r *PostgresRepository) GetS3BucketsToScaleUp(ctx context.Context) ([]S3ScaleCandidate, error) {
	var candidates []S3ScaleCandidate
	query := `
		SELECT bucket_id, MAX(storage_used_bytes) as max_storage
		FROM (
			SELECT bucket_id, storage_used_bytes,
				ROW_NUMBER() OVER(PARTITION BY bucket_id ORDER BY created_at DESC) as rn
			FROM s3_metrics
		) tmp
		WHERE rn <= 1
		GROUP BY bucket_id
		HAVING MAX(storage_used_bytes) > 107374182400
	` // 100 GB in bytes
	err := r.db.SelectContext(ctx, &candidates, query)
	return candidates, err
}


package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"metrics-gateway/repository"

	"github.com/nats-io/nats.go"
)

// ScalingPolicyNATSHandler listens for events to create or update dynamic scaling policies.
type ScalingPolicyNATSHandler struct {
	logger *slog.Logger
	nc     *nats.Conn
	repo   *repository.PostgresRepository
}

// NewScalingPolicyNATSHandler creates a new handler.
func NewScalingPolicyNATSHandler(logger *slog.Logger, nc *nats.Conn, repo *repository.PostgresRepository) *ScalingPolicyNATSHandler {
	return &ScalingPolicyNATSHandler{
		logger: logger,
		nc:     nc,
		repo:   repo,
	}
}

// Start subscribes to the dynamic scaling policy subjects.
func (h *ScalingPolicyNATSHandler) Start() error {
	queueGroup := "metrics-service"

	subs := map[string]nats.MsgHandler{
		"dev.metrics.v1.scaling_policy.create": h.handleCreatePolicy,
		"dev.metrics.v1.scaling_policy.update": h.handleUpdatePolicy,
		"dev.metrics.v1.scaling_policy.delete": h.handleDeletePolicy,
		"dev.metrics.v1.scaling_policy.list":   h.handleListPolicies,
		"dev.rds.v1.scaling_policy.create":     h.handleCreateRDSPolicy,
		"dev.rds.v1.scaling_policy.update":     h.handleUpdateRDSPolicy,
		"dev.rds.v1.scaling_policy.delete":     h.handleDeleteRDSPolicy,
		"dev.rds.v1.scaling_policy.list":       h.handleListRDSPolicies,
		"dev.lambda.v1.scaling_policy.create":  h.handleCreateLambdaPolicy,
		"dev.lambda.v1.scaling_policy.update":  h.handleUpdateLambdaPolicy,
		"dev.lambda.v1.scaling_policy.delete":  h.handleDeleteLambdaPolicy,
		"dev.lambda.v1.scaling_policy.list":    h.handleListLambdaPolicies,
	}

	for subject, handler := range subs {
		if _, err := h.nc.QueueSubscribe(subject, queueGroup, handler); err != nil {
			return err
		}
		h.logger.Info("subscribed to NATS scaling policy event", "subject", subject, "queue", queueGroup)
	}

	return nil
}

// ── Event Structs ────────────────────────────────────────────────────────────

type ScalingPolicyCreateEvent struct {
	CorrelationID string                      `json:"correlation_id"`
	TenantID      string                      `json:"tenant_id"`
	Policy        repository.EC2ScalingPolicy `json:"policy"`
}

type ScalingPolicyUpdateEvent struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
	PolicyID      string `json:"policy_id"`
	Update        struct {
		TargetValue      float64 `json:"target_value"`
		ScaleDownValue   float64 `json:"scale_down_value"`
		MaxInstances     int     `json:"max_instances"`
		ScaleOutCooldown int     `json:"scale_out_cooldown"`
		ScaleInCooldown  int     `json:"scale_in_cooldown"`
	} `json:"update"`
}

type ScalingPolicyDeleteEvent struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
	PolicyID      string `json:"policy_id"`
}

type ScalingPolicyListRequest struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
}

type ScalingPolicyListResponse struct {
	Policies []repository.EC2ScalingPolicy `json:"policies"`
	Error    string                        `json:"error,omitempty"`
}

type RDSScalingPolicyCreateEvent struct {
	CorrelationID string                      `json:"correlation_id"`
	TenantID      string                      `json:"tenant_id"`
	Policy        repository.RDSScalingPolicy `json:"policy"`
}

type RDSScalingPolicyUpdateEvent struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
	InstanceID    string `json:"instance_id"`
	MetricName    string `json:"metric_name"`
	Update        struct {
		ScaleUpThreshold   float64 `json:"scale_up_threshold"`
		ScaleDownThreshold float64 `json:"scale_down_threshold"`
		MaxLimit           float64 `json:"max_limit"`
		MinLimit           float64 `json:"min_limit"`
		ScaleStep          float64 `json:"scale_step"`
		CooldownSeconds    int64   `json:"cooldown_seconds"`
	} `json:"update"`
}

type RDSScalingPolicyDeleteEvent struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
	InstanceID    string `json:"instance_id"`
	MetricName    string `json:"metric_name"`
}

type RDSScalingPolicyListResponse struct {
	Policies []repository.RDSScalingPolicy `json:"policies"`
	Error    string                        `json:"error,omitempty"`
}

type LambdaScalingPolicyCreateEvent struct {
	CorrelationID string                          `json:"correlation_id"`
	TenantID      string                          `json:"tenant_id"`
	Policy        repository.LambdaScalingPolicy  `json:"policy"`
}

type LambdaScalingPolicyUpdateEvent struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
	FunctionID    string `json:"function_id"`
	MetricName    string `json:"metric_name"`
	Update        struct {
		ScaleUpThreshold     float64 `json:"scale_up_threshold"`
		ScaleDownThreshold   float64 `json:"scale_down_threshold"`
		MaxConcurrencyLimit  int64   `json:"max_concurrency_limit"`
		MinConcurrencyLimit  int64   `json:"min_concurrency_limit"`
		ScaleStep            int64   `json:"scale_step"`
		CooldownSeconds      int64   `json:"cooldown_seconds"`
	} `json:"update"`
}

type LambdaScalingPolicyDeleteEvent struct {
	CorrelationID string `json:"correlation_id"`
	TenantID      string `json:"tenant_id"`
	FunctionID    string `json:"function_id"`
	MetricName    string `json:"metric_name"`
}

type LambdaScalingPolicyListResponse struct {
	Policies []repository.LambdaScalingPolicy `json:"policies"`
	Error    string                           `json:"error,omitempty"`
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func (h *ScalingPolicyNATSHandler) handleCreatePolicy(msg *nats.Msg) {
	var event ScalingPolicyCreateEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode scaling policy create request", "error", err)
		return
	}

	req := event.Policy
	req.TenantID = event.TenantID

	if req.TargetID == "" || req.TargetValue <= 0 || req.TargetType == "" || req.TenantID == "" {
		h.logger.Error("invalid scaling policy create payload", "tenant_id", event.TenantID)
		return
	}

	// Set defaults for optional parameters if they were provided as zero values.
	if req.MaxInstances <= 0 {
		req.MaxInstances = 5
	}

	if req.MetricName == "" {
		req.MetricName = "CPUUtilization" // default
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateEC2ScalingPolicy(ctx, req); err != nil {
		h.logger.Error("failed to save dynamic EC2 scaling policy", "error", err, "target_id", req.TargetID)
		return
	}

	h.logger.Info("successfully created scaling policy", "tenant_id", req.TenantID, "target_id", req.TargetID)
}

func (h *ScalingPolicyNATSHandler) handleUpdatePolicy(msg *nats.Msg) {
	var event ScalingPolicyUpdateEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode scaling policy update request", "error", err)
		return
	}

	if event.TenantID == "" || event.PolicyID == "" || event.Update.TargetValue <= 0 {
		h.logger.Error("invalid scaling policy update payload", "tenant_id", event.TenantID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateEC2ScalingPolicy(ctx, event.TenantID, event.PolicyID, event.Update.TargetValue, event.Update.ScaleDownValue, event.Update.MaxInstances, event.Update.ScaleOutCooldown, event.Update.ScaleInCooldown); err != nil {
		h.logger.Error("failed to update dynamic EC2 scaling policy", "error", err, "policy_id", event.PolicyID)
		return
	}

	h.logger.Info("successfully updated scaling policy", "tenant_id", event.TenantID, "policy_id", event.PolicyID)
}

func (h *ScalingPolicyNATSHandler) handleDeletePolicy(msg *nats.Msg) {
	var event ScalingPolicyDeleteEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode scaling policy delete request", "error", err)
		return
	}

	if event.TenantID == "" || event.PolicyID == "" {
		h.logger.Error("invalid scaling policy delete payload", "tenant_id", event.TenantID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteEC2ScalingPolicy(ctx, event.TenantID, event.PolicyID); err != nil {
		h.logger.Error("failed to delete dynamic EC2 scaling policy", "error", err, "policy_id", event.PolicyID)
		return
	}

	h.logger.Info("successfully deleted scaling policy", "tenant_id", event.TenantID, "policy_id", event.PolicyID)
}

func (h *ScalingPolicyNATSHandler) handleListPolicies(msg *nats.Msg) {
	var req ScalingPolicyListRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("failed to decode scaling policy list request", "error", err)
		h.replyError(msg, "invalid request format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	policies, err := h.repo.ListEC2ScalingPolicies(ctx, req.TenantID)
	if err != nil {
		h.logger.Error("failed to list scaling policies", "error", err, "tenant_id", req.TenantID)
		h.replyError(msg, "internal server error")
		return
	}

	resp := ScalingPolicyListResponse{
		Policies: policies,
	}

	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func (h *ScalingPolicyNATSHandler) handleCreateRDSPolicy(msg *nats.Msg) {
	var event RDSScalingPolicyCreateEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode RDS scaling policy create request", "error", err)
		return
	}

	req := event.Policy
	req.TenantID = event.TenantID

	if req.InstanceID == "" || req.ScaleUpThreshold <= 0 || req.MetricName == "" || req.TenantID == "" {
		h.logger.Error("invalid RDS scaling policy create payload", "tenant_id", event.TenantID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateRDSScalingPolicy(ctx, req); err != nil {
		h.logger.Error("failed to save dynamic RDS scaling policy", "error", err, "instance_id", req.InstanceID)
		return
	}

	h.logger.Info("successfully created RDS scaling policy", "tenant_id", req.TenantID, "instance_id", req.InstanceID)
}

func (h *ScalingPolicyNATSHandler) handleUpdateRDSPolicy(msg *nats.Msg) {
	var event RDSScalingPolicyUpdateEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode RDS scaling policy update request", "error", err)
		return
	}

	if event.TenantID == "" || event.InstanceID == "" || event.MetricName == "" {
		h.logger.Error("invalid RDS scaling policy update payload", "tenant_id", event.TenantID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateRDSScalingPolicy(ctx, event.TenantID, event.InstanceID, event.MetricName, event.Update.ScaleUpThreshold, event.Update.ScaleDownThreshold, event.Update.MaxLimit, event.Update.MinLimit, event.Update.ScaleStep, event.Update.CooldownSeconds); err != nil {
		h.logger.Error("failed to update dynamic RDS scaling policy", "error", err, "instance_id", event.InstanceID)
		return
	}

	h.logger.Info("successfully updated RDS scaling policy", "tenant_id", event.TenantID, "instance_id", event.InstanceID)
}

func (h *ScalingPolicyNATSHandler) handleDeleteRDSPolicy(msg *nats.Msg) {
	var event RDSScalingPolicyDeleteEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode RDS scaling policy delete request", "error", err)
		return
	}

	if event.TenantID == "" || event.InstanceID == "" || event.MetricName == "" {
		h.logger.Error("invalid RDS scaling policy delete payload", "tenant_id", event.TenantID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteRDSScalingPolicy(ctx, event.TenantID, event.InstanceID, event.MetricName); err != nil {
		h.logger.Error("failed to delete dynamic RDS scaling policy", "error", err, "instance_id", event.InstanceID)
		return
	}

	h.logger.Info("successfully deleted RDS scaling policy", "tenant_id", event.TenantID, "instance_id", event.InstanceID)
}

func (h *ScalingPolicyNATSHandler) handleListRDSPolicies(msg *nats.Msg) {
	var req ScalingPolicyListRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("failed to decode RDS scaling policy list request", "error", err)
		h.replyRDSError(msg, "invalid request format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	policies, err := h.repo.ListRDSScalingPolicies(ctx, req.TenantID)
	if err != nil {
		h.logger.Error("failed to list RDS scaling policies", "error", err, "tenant_id", req.TenantID)
		h.replyRDSError(msg, "internal server error")
		return
	}

	resp := RDSScalingPolicyListResponse{
		Policies: policies,
	}

	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func (h *ScalingPolicyNATSHandler) replyRDSError(msg *nats.Msg, errMsg string) {
	resp := RDSScalingPolicyListResponse{
		Error: errMsg,
	}
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func (h *ScalingPolicyNATSHandler) replyError(msg *nats.Msg, errMsg string) {
	resp := ScalingPolicyListResponse{
		Error: errMsg,
	}
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func (h *ScalingPolicyNATSHandler) handleCreateLambdaPolicy(msg *nats.Msg) {



	fmt.Println("dev.lambda.v1.scaling_policy.create")
	fmt.Println(string(msg.Data))
	var event LambdaScalingPolicyCreateEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode Lambda scaling policy create request", "error", err)
		h.replyLambdaError(msg, "invalid request format")
		return
	}

	req := event.Policy
	req.TenantID = event.TenantID

	if req.FunctionID == "" || req.ScaleUpThreshold <= 0 || req.MetricName == "" || req.TenantID == "" {
		h.logger.Error("invalid Lambda scaling policy create payload", "tenant_id", event.TenantID)
		h.replyLambdaError(msg, "invalid payload")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateLambdaScalingPolicy(ctx, req); err != nil {
		h.logger.Error("failed to save dynamic Lambda scaling policy", "error", err, "function_id", req.FunctionID)
		h.replyLambdaError(msg, "failed to create policy")
		return
	}

	h.logger.Info("successfully created Lambda scaling policy", "tenant_id", req.TenantID, "function_id", req.FunctionID)
	msg.Respond([]byte(`{"message": "success"}`))
}

func (h *ScalingPolicyNATSHandler) handleUpdateLambdaPolicy(msg *nats.Msg) {
	var event LambdaScalingPolicyUpdateEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode Lambda scaling policy update request", "error", err)
		h.replyLambdaError(msg, "invalid request format")
		return
	}

	if event.TenantID == "" || event.FunctionID == "" || event.MetricName == "" {
		h.logger.Error("invalid Lambda scaling policy update payload", "tenant_id", event.TenantID)
		h.replyLambdaError(msg, "invalid payload")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateLambdaScalingPolicy(ctx, event.TenantID, event.FunctionID, event.MetricName, event.Update.ScaleUpThreshold, event.Update.ScaleDownThreshold, event.Update.MaxConcurrencyLimit, event.Update.MinConcurrencyLimit, event.Update.ScaleStep, event.Update.CooldownSeconds); err != nil {
		h.logger.Error("failed to update dynamic Lambda scaling policy", "error", err, "function_id", event.FunctionID)
		h.replyLambdaError(msg, "failed to update policy")
		return
	}

	h.logger.Info("successfully updated Lambda scaling policy", "tenant_id", event.TenantID, "function_id", event.FunctionID)
	msg.Respond([]byte(`{"message": "success"}`))
}

func (h *ScalingPolicyNATSHandler) handleDeleteLambdaPolicy(msg *nats.Msg) {
	var event LambdaScalingPolicyDeleteEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Error("failed to decode Lambda scaling policy delete request", "error", err)
		h.replyLambdaError(msg, "invalid request format")
		return
	}

	if event.TenantID == "" || event.FunctionID == "" || event.MetricName == "" {
		h.logger.Error("invalid Lambda scaling policy delete payload", "tenant_id", event.TenantID)
		h.replyLambdaError(msg, "invalid payload")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteLambdaScalingPolicy(ctx, event.TenantID, event.FunctionID, event.MetricName); err != nil {
		h.logger.Error("failed to delete dynamic Lambda scaling policy", "error", err, "function_id", event.FunctionID)
		h.replyLambdaError(msg, "failed to delete policy")
		return
	}

	h.logger.Info("successfully deleted Lambda scaling policy", "tenant_id", event.TenantID, "function_id", event.FunctionID)
	msg.Respond([]byte(`{"message": "success"}`))
}

func (h *ScalingPolicyNATSHandler) handleListLambdaPolicies(msg *nats.Msg) {
	var req ScalingPolicyListRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("failed to decode Lambda scaling policy list request", "error", err)
		h.replyLambdaError(msg, "invalid request format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	policies, err := h.repo.ListLambdaScalingPolicies(ctx, req.TenantID)
	if err != nil {
		h.logger.Error("failed to list Lambda scaling policies", "error", err, "tenant_id", req.TenantID)
		h.replyLambdaError(msg, "internal server error")
		return
	}

	resp := LambdaScalingPolicyListResponse{
		Policies: policies,
	}

	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)



fmt.Println("Lambda scaling policies listed---- successfully")
}

func (h *ScalingPolicyNATSHandler) replyLambdaError(msg *nats.Msg, errMsg string) {
	resp := LambdaScalingPolicyListResponse{
		Error: errMsg,
	}
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

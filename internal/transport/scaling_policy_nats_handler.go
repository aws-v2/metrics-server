package transport

import (
	"context"
	"encoding/json"
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

func (h *ScalingPolicyNATSHandler) replyError(msg *nats.Msg, errMsg string) {
	resp := ScalingPolicyListResponse{
		Error: errMsg,
	}
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

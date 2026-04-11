package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"metrics-gateway/application"
	"metrics-gateway/internal/messaging"

	"github.com/nats-io/nats.go"
)

// BillingNATSHandler listens to NATS and responds to billing aggregation requests.
type BillingNATSHandler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	svc     application.BillingService
	profile string
}

// NewBillingNATSHandler creates a new handler.
func NewBillingNATSHandler(logger *slog.Logger, nc *nats.Conn, svc application.BillingService, profile string) *BillingNATSHandler {
	return &BillingNATSHandler{
		logger:  logger,
		nc:      nc,
		svc:     svc,
		profile: profile,
	}
}

// Start subscribes to the billing subject.
func (h *BillingNATSHandler) Start() error {
	subject := messaging.BuildSubject(h.profile, "metrics", "v1", "billing", "usage.get")
	_, err := h.nc.Subscribe(subject, h.handleBillingRequest)
	if err != nil {
		return err
	}
	h.logger.Info("subscribed to NATS billing requests", "subject", subject)
	return nil
}

// handleBillingRequest routes the request based on the resource prefix string
// (e.g. i-xxx = ec2, f-xxx = lambda, etc.) or a specified type if provided.
// For simplicity, we assume the Request contains the ResourceID and time bounds.
func (h *BillingNATSHandler) handleBillingRequest(msg *nats.Msg) {
	if msg.Reply == "" {
		h.logger.Warn("billing request received without reply subject")
		return
	}

	var req application.BillingUsageRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("failed to decode billing request", "error", err)
		h.replyError(msg.Reply, "invalid JSON payload")
		return
	}

	if req.ResourceID == "" || req.StartTime.IsZero() || req.EndTime.IsZero() {
		h.replyError(msg.Reply, "missing resource_id, start_time, or end_time")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var respPayload []byte
	var err error

	// Minimal heuristic to determine service type based on prefix if needed,
	// or assume the caller knows. Here we use basic prefix routing:
	// "i-" -> EC2
	// "rds-" -> RDS
	// "arn:aws:lambda:" or similar -> Lambda
	// We'll just do simple string matching for this demo format.
	if strings.HasPrefix(req.ResourceID, "i-") {
		var res application.EC2UsageResponse
		res, err = h.svc.CalculateEC2Usage(ctx, req.ResourceID, req.StartTime, req.EndTime)
		respPayload, _ = json.Marshal(res)
	} else if strings.HasPrefix(req.ResourceID, "rds-") {
		var res application.RDSUsageResponse
		res, err = h.svc.CalculateRDSUsage(ctx, req.ResourceID, req.StartTime, req.EndTime)
		respPayload, _ = json.Marshal(res)
	} else if strings.HasPrefix(req.ResourceID, "fun-") || strings.Contains(req.ResourceID, "lambda") {
		var res application.LambdaUsageResponse
		res, err = h.svc.CalculateLambdaUsage(ctx, req.ResourceID, req.StartTime, req.EndTime)
		respPayload, _ = json.Marshal(res)
	} else {
		// Default to S3 if unknown, or return error. We'll attempt S3.
		var res application.S3UsageResponse
		res, err = h.svc.CalculateS3Usage(ctx, req.ResourceID, req.StartTime, req.EndTime)
		respPayload, _ = json.Marshal(res)
	}

	if err != nil {
		h.logger.Error("failed to calculate billing usage", "resource", req.ResourceID, "error", err)
		h.replyError(msg.Reply, "internal server error calculating usage")
		return
	}

	h.nc.Publish(msg.Reply, respPayload)
}

func (h *BillingNATSHandler) replyError(replyTo, errMsg string) {
	resp := map[string]string{"error": errMsg}
	payload, _ := json.Marshal(resp)
	h.nc.Publish(replyTo, payload)
}

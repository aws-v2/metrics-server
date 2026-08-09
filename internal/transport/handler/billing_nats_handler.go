package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	// "strings"
	"time"

	"metrics-gateway/application"

	"github.com/nats-io/nats.go"
)

// BillingNATSHandler listens to NATS and responds to billing aggregation requests.
type BillingNATSHandler struct {
	logger  *slog.Logger
	nc      *nats.Conn
	svc     application.BillingService
	prefix string
}

// NewBillingNATSHandler creates a new handler.
func NewBillingNATSHandler(logger *slog.Logger, nc *nats.Conn, svc application.BillingService, prefix string) *BillingNATSHandler {
	return &BillingNATSHandler{
		logger:  logger,
		nc:      nc,
		svc:     svc,
		prefix: prefix,
	}
}

// Start subscribes to the billing subject.
func (h *BillingNATSHandler) Start() error {
	
	subject := h.prefix+".metrics.raw.logs"
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
	fmt.Println("Recieved metrics from a* service")
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

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var respPayload []byte
	var err error


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

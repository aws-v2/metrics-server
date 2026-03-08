package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
)

// EC2Handler handles EC2 metrics HTTP requests.
type EC2Handler struct {
	logger  *slog.Logger
	service application.EC2MetricsService
}

// NewEC2Handler creates a new EC2Handler.
func NewEC2Handler(logger *slog.Logger, service application.EC2MetricsService) *EC2Handler {
	return &EC2Handler{logger: logger, service: service}
}

// Ingest handles POST /api/v1/metrics-server/ec2/ingest
func (h *EC2Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	// Debug logging of the raw body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read raw body", "error", err)
	} else {
		h.logger.Info("received raw EC2 ingest payload", "body", string(bodyBytes))
		// Restore body for the decoder
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var req application.EC2IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid EC2 ingest payload", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.InstanceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "instance_id is required"})
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ingest metric"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// GetByInstance handles GET /api/v1/metrics-server/ec2/{instanceId}
func (h *EC2Handler) GetByInstance(w http.ResponseWriter, r *http.Request) {
	instanceID := r.PathValue("instanceId")
	if instanceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "instanceId is required"})
		return
	}

	metrics, err := h.service.GetByInstance(r.Context(), instanceID)
	if err != nil {
		h.logger.Error("failed to get EC2 metrics", "instance_id", instanceID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// List handles GET /api/v1/metrics-server/ec2
func (h *EC2Handler) List(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list EC2 metrics", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

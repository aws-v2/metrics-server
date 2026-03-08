package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
)

// RDSHandler handles RDS metrics HTTP requests.
type RDSHandler struct {
	logger  *slog.Logger
	service application.RDSMetricsService
}

// NewRDSHandler creates a new RDSHandler.
func NewRDSHandler(logger *slog.Logger, service application.RDSMetricsService) *RDSHandler {
	return &RDSHandler{logger: logger, service: service}
}

// Ingest handles POST /api/v1/metrics-server/rds/ingest
func (h *RDSHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req application.RDSIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid RDS ingest payload", "error", err)
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

// GetByInstance handles GET /api/v1/metrics-server/rds/{instanceId}
func (h *RDSHandler) GetByInstance(w http.ResponseWriter, r *http.Request) {
	instanceID := r.PathValue("instanceId")
	if instanceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "instanceId is required"})
		return
	}

	metrics, err := h.service.GetByInstance(r.Context(), instanceID)
	if err != nil {
		h.logger.Error("failed to get RDS metrics", "instance_id", instanceID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// List handles GET /api/v1/metrics-server/rds
func (h *RDSHandler) List(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list RDS metrics", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
)

// SageMakerHandler handles SageMaker metrics HTTP requests.
type SageMakerHandler struct {
	logger  *slog.Logger
	service application.SageMakerMetricsService
}

// NewSageMakerHandler creates a new SageMakerHandler.
func NewSageMakerHandler(logger *slog.Logger, service application.SageMakerMetricsService) *SageMakerHandler {
	return &SageMakerHandler{logger: logger, service: service}
}

// Ingest handles POST /api/v1/metrics-server/sagemaker/ingest
func (h *SageMakerHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req application.SageMakerIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid SageMaker ingest payload", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.EndpointID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "endpoint_id is required"})
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ingest metric"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// GetByEndpoint handles GET /api/v1/metrics-server/sagemaker/{endpointId}
func (h *SageMakerHandler) GetByEndpoint(w http.ResponseWriter, r *http.Request) {
	endpointID := r.PathValue("endpointId")
	if endpointID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "endpointId is required"})
		return
	}

	metrics, err := h.service.GetByEndpoint(r.Context(), endpointID)
	if err != nil {
		h.logger.Error("failed to get SageMaker metrics", "endpoint_id", endpointID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// List handles GET /api/v1/metrics-server/sagemaker
func (h *SageMakerHandler) List(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list SageMaker metrics", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

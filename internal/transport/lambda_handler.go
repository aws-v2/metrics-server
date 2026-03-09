package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"fmt"

	"metrics-gateway/application"
)

// LambdaHandler handles Lambda metrics HTTP requests.
type LambdaHandler struct {
	logger  *slog.Logger
	service application.LambdaMetricsService
}

// NewLambdaHandler creates a new LambdaHandler.
func NewLambdaHandler(logger *slog.Logger, service application.LambdaMetricsService) *LambdaHandler {
	return &LambdaHandler{logger: logger, service: service}
}

// Ingest handles POST /api/v1/metrics-server/lambda/ingest
func (h *LambdaHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Lambda ingest request received")
	var req application.LambdaIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid Lambda ingest payload", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.FunctionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "function_id is required"})
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ingest metric"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// GetByFunction handles GET /api/v1/metrics-server/lambda/{functionId}
func (h *LambdaHandler) GetByFunction(w http.ResponseWriter, r *http.Request) {
	functionID := r.PathValue("functionId")
	if functionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "functionId is required"})
		return
	}

	metrics, err := h.service.GetByFunction(r.Context(), functionID)
	if err != nil {
		h.logger.Error("failed to get Lambda metrics", "function_id", functionID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// List handles GET /api/v1/metrics-server/lambda
func (h *LambdaHandler) List(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list Lambda metrics", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

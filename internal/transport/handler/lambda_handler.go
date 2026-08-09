package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
	"metrics-gateway/internal/utils"
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
	requestID := r.Context().Value("requestId")

	var req application.LambdaIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Handler:LambdaIngest] Payload unamarshal, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if req.FunctionID == "" {
		log.Printf("[Handler:LambdaIngest] Function id not found, requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("function_id is required"))
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		log.Printf("[Handler:GetManifest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("failed to ingest metric"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusCreated, "Metrics ingested successfully", map[string]string{"status": "ok"})

}

// GetByFunction handles GET /api/v1/metrics-server/lambda/{functionId}
func (h *LambdaHandler) GetByFunction(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")
	functionID := r.PathValue("functionId")
	if functionID == "" {
		log.Printf("[Handler:LambdaGetByFunction] Function id not found, requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("function_id is required"))
		return
	}

	metrics, err := h.service.GetByFunction(r.Context(), functionID)
	if err != nil {

		log.Printf("[Handler:LambdaGetByFunction] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get metric"))
		return

	}
	utils.WriteJSONSucces(w, http.StatusCreated, "Metrics fetched successfully", metrics)

}

// List handles GET /api/v1/metrics-server/lambda
func (h *LambdaHandler) List(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")

	metrics, err := h.service.List(r.Context())
	if err != nil {

		log.Printf("[Handler:LambdaListByFunction] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to list metrics"))
		return

	}
	utils.WriteJSONSucces(w, http.StatusCreated, "Metrics fetched successfully", metrics)

}

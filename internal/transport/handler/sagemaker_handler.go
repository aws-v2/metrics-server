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
	requestID := r.Context().Value("requestId")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Handler:SGMIngest] Payload unmarshal, bad request for requestID %s, with error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return

	}

	if req.EndpointID == "" {
		log.Printf("[Handler:SGMIngest] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("endpouint id is required"))
		return

	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		log.Printf("[Handler:SGMIngest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to ingest metric"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusCreated, "ingested metrics succesfully", map[string]string{"status": "ok"})

}

// GetByEndpoint handles GET /api/v1/metrics-server/sagemaker/{endpointId}
func (h *SageMakerHandler) GetByEndpoint(w http.ResponseWriter, r *http.Request) {
	endpointID := r.PathValue("endpointId")
	requestID := r.Context().Value("requestId")

	if endpointID == "" {
		log.Printf("[Handler:SGMGetByEndpoint] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance id is required"))
		return
	}

	metrics, err := h.service.GetByEndpoint(r.Context(), endpointID)
	if err != nil {
		log.Printf("[Handler:SGMGetByEndpoint] Service call, requestID %s  error %s", requestID, err.Error())

		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get rds metrics"))

	}

	utils.WriteJSONSucces(w, http.StatusOK, "Fetched instance metrics succesfully", metrics)

}

// List handles GET /api/v1/metrics-server/sagemaker
func (h *SageMakerHandler) List(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")

	metrics, err := h.service.List(r.Context())
	if err != nil {

		log.Printf("[Handler:SGMList] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to list metrics"))

		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched metrics succesfully", metrics)
}

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
	requestID := r.Context().Value("requestId")

	var req application.RDSIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Handler:RDSIngest] Payload unmarshal, bad request for requestID %s, with error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if req.InstanceID == "" {
		log.Printf("[Handler:RDSIngest] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance id is required"))
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		log.Printf("[Handler:RDSIngest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to ingest metric"))
		return
	}

	utils.WriteJSONSucces(w, http.StatusCreated, "ingested metrics succesfully", map[string]string{"status": "ok"})

}

// GetByInstance handles GET /api/v1/metrics-server/rds/{instanceId}
func (h *RDSHandler) GetByInstance(w http.ResponseWriter, r *http.Request) {
	instanceID := r.PathValue("instanceId")
	requestID := r.Context().Value("requestId")

	if instanceID == "" {
		log.Printf("[Handler:RdsGetByInstance] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance id is required"))
		return
	}
	metrics, err := h.service.GetByInstance(r.Context(), instanceID)
	if err != nil {
		log.Printf("[Handler:RdsGetByInstance] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get rds metrics"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched instance metrics succesfully", metrics)

}

// List handles GET /api/v1/metrics-server/rds
func (h *RDSHandler) List(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")

	metrics, err := h.service.List(r.Context())
	if err != nil {

		log.Printf("[Handler:RdsList] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to list metrics"))

		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched metrics succesfully", metrics)

}

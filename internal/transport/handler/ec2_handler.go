package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
	"metrics-gateway/internal/utils"
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
	requestID := r.Context().Value("requestId")

	// Debug logging of the raw body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Handler:EC2Ingest] Payload unmarshal, bad request for requestID %s, with error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance_id is required"))

	} else {
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var req application.EC2IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Handler:EC2Ingest] Payload unmarshal, bad request for requestID %s, with error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if req.InstanceID == "" {
		log.Printf("[Handler:EC2GetByInstance] Instance variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance_id is required"))
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		log.Printf("[Handler:EC2GetByInstance] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to ingest metric"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusCreated, "ingested succesfully", map[string]string{"status": "ok"})

}

// GetByInstance handles GET /api/v1/metrics-server/ec2/{instanceId}
func (h *EC2Handler) GetByInstance(w http.ResponseWriter, r *http.Request) {
	instanceID := r.PathValue("instanceId")
	requestID := r.Context().Value("requestId")

	if instanceID == "" {
		log.Printf("[Handler:EC2GetByInstance] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance id is required"))
		return
	}

	metrics, err := h.service.GetByInstance(r.Context(), instanceID)
	if err != nil {
		log.Printf("[Handler:EC2GetByInstance] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("failed to get EC2 metrics"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched metrics succesfully", metrics)

}

// List handles GET /api/v1/metrics-server/ec2
func (h *EC2Handler) List(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")
	metrics, err := h.service.List(r.Context())
	if err != nil {
		log.Printf("[Handler:EC2ListInstanceAll] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to list metrics"))

		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched metrics succesfully", metrics)

}

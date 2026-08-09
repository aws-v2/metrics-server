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

// S3Handler handles S3 metrics HTTP requests.
type S3Handler struct {
	logger  *slog.Logger
	service application.S3MetricsService
}

// NewS3Handler creates a new S3Handler.
func NewS3Handler(logger *slog.Logger, service application.S3MetricsService) *S3Handler {
	return &S3Handler{logger: logger, service: service}
}

// Ingest handles POST /api/v1/metrics-server/s3/ingest
func (h *S3Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req application.S3IngestRequest
	requestID := r.Context().Value("requestId")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Handler:S3Ingest] Payload unmarshal, bad request for requestID %s, with error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return

	}

	if req.BucketID == "" {
		log.Printf("[Handler:S3Ingest] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("bucket id is required"))
		return

	}

	if err := h.service.Ingest(r.Context(), req); err != nil {

		log.Printf("[Handler:S3Ingest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to ingest metric"))
		return

	}

	utils.WriteJSONSucces(w, http.StatusCreated, "ingested metrics succesfully", map[string]string{"status": "ok"})

}

// GetByBucket handles GET /api/v1/metrics-server/s3/{bucketId}
func (h *S3Handler) GetByBucket(w http.ResponseWriter, r *http.Request) {
	bucketID := r.PathValue("bucketId")
	requestID := r.Context().Value("requestId")

	if bucketID == "" {
		log.Printf("[Handler:RdsGetByInstance] Path variable not found, bad request for requestID %s", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("instance id is required"))
		return
	}

	metrics, err := h.service.GetByBucket(r.Context(), bucketID)
	if err != nil {
		log.Printf("[Handler:S3etByBucket] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get rds metrics"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched instance metrics succesfully", metrics)

}

// List handles GET /api/v1/metrics-server/s3
func (h *S3Handler) List(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")

	metrics, err := h.service.List(r.Context())
	if err != nil {

		log.Printf("[Handler:S3List] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to list metrics"))
		return
	}

	utils.WriteJSONSucces(w, http.StatusOK, "Fetched metrics succesfully", metrics)

}

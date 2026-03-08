package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid S3 ingest payload", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.BucketID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bucket_id is required"})
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ingest metric"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// GetByBucket handles GET /api/v1/metrics-server/s3/{bucketId}
func (h *S3Handler) GetByBucket(w http.ResponseWriter, r *http.Request) {
	bucketID := r.PathValue("bucketId")
	if bucketID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bucketId is required"})
		return
	}

	metrics, err := h.service.GetByBucket(r.Context(), bucketID)
	if err != nil {
		h.logger.Error("failed to get S3 metrics", "bucket_id", bucketID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// List handles GET /api/v1/metrics-server/s3
func (h *S3Handler) List(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list S3 metrics", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

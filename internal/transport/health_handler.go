package transport

import (
	"log/slog"
	"net/http"

	"metrics-gateway/application"
)

// HealthHandler handles health-check HTTP requests.
type HealthHandler struct {
	logger  *slog.Logger
	service application.HealthChecker
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(logger *slog.Logger, service application.HealthChecker) *HealthHandler {
	return &HealthHandler{
		logger:  logger,
		service: service,
	}
}

// Health responds with the service health status.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.Check(r.Context())
	if err != nil {
		h.logger.Error("health check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	h.logger.Info("health check", "status", status)
	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

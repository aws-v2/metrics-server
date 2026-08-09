package handler

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
	"metrics-gateway/internal/utils"
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
	requestID := r.Context().Value("requestId")

	status, err := h.service.Check(r.Context())
	if err != nil {
		log.Printf("[Handler:Health] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusServiceUnavailable, fmt.Errorf("Unhealthy service "))
		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched Public manifest succesfully", map[string]string{"status": status})

}

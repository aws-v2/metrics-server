package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"metrics-gateway/application"
)

// GameLiftHandler handles GameLift metrics HTTP requests.
type GameLiftHandler struct {
	logger  *slog.Logger
	service application.GameLiftMetricsService
}

// NewGameLiftHandler creates a new GameLiftHandler.
func NewGameLiftHandler(logger *slog.Logger, service application.GameLiftMetricsService) *GameLiftHandler {
	return &GameLiftHandler{logger: logger, service: service}
}

// Ingest handles POST /api/v1/metrics-server/gamelift/ingest
func (h *GameLiftHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req application.GameLiftIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid GameLift ingest payload", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.FleetID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "fleet_id is required"})
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ingest metric"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// GetByFleet handles GET /api/v1/metrics-server/gamelift/{fleetId}
func (h *GameLiftHandler) GetByFleet(w http.ResponseWriter, r *http.Request) {
	fleetID := r.PathValue("fleetId")
	if fleetID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "fleetId is required"})
		return
	}

	metrics, err := h.service.GetByFleet(r.Context(), fleetID)
	if err != nil {
		h.logger.Error("failed to get GameLift metrics", "fleet_id", fleetID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// List handles GET /api/v1/metrics-server/gamelift
func (h *GameLiftHandler) List(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list GameLift metrics", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list metrics"})
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

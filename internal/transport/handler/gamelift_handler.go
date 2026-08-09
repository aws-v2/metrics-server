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
	requestID := r.Context().Value("requestId")

	var req application.GameLiftIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Handler:AssignVPC] Payload unmarshal, bad request for requestID %s, with error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("invalid GameLift ingest payload"))
		return
	}

	if req.FleetID == "" {
		log.Printf("[Handler:GLSIngest] fleet_id not found, bad request for requestID %s, ", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("fleet_id is required"))
		return
	}

	if err := h.service.Ingest(r.Context(), req); err != nil {
		log.Printf("[Handler:GetInternalManifest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to ingest metric"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusCreated, "Fetched Public manifest succesfully", map[string]string{"status": "ok"})
}

// GetByFleet handles GET /api/v1/metrics-server/gamelift/{fleetId}
func (h *GameLiftHandler) GetByFleet(w http.ResponseWriter, r *http.Request) {
	fleetID := r.PathValue("fleetId")
	requestID := r.Context().Value("requestId")

	if fleetID == "" {
		log.Printf("[Handler:GLSGetByFleet] fleet_id not found, bad request for requestID %s, ", requestID)
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("fleet_id is required"))
		return
	}

	metrics, err := h.service.GetByFleet(r.Context(), fleetID)
	if err != nil {
		log.Printf("[Handler:GLSGetByFleet] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve metrics"))

		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched Public metrics succesfully", metrics)

}

// List handles GET /api/v1/metrics-server/gamelift
func (h *GameLiftHandler) List(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")

	metrics, err := h.service.List(r.Context())
	if err != nil {
		log.Printf("[Handler:GLSListByFleet] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to list GameLift metrics"))

		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched Public metrics succesfully", metrics)

}

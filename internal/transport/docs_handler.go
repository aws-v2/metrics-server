package transport

import (
	"net/http"

	"metrics-gateway/application"
)

type DocsHandler struct {
	service *application.DocsService
}

func NewDocsHandler(service *application.DocsService) *DocsHandler {
	return &DocsHandler{service: service}
}

 
// ── handlers ──────────────────────────────────────────────────────────────────

func (h *DocsHandler) GetPublicManifest(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetManifest(false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h *DocsHandler) GetInternalManifest(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetManifest(true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h *DocsHandler) GetPublicDoc(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	doc, err := h.service.GetDoc(slug, false)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": doc})
}

func (h *DocsHandler) GetInternalDoc(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	doc, err := h.service.GetDoc(slug, true)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": doc})
}
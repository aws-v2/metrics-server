package handler

import (
	"fmt"
	"log"
	"net/http"

	"metrics-gateway/application"
	"metrics-gateway/internal/utils"
)

type DocsHandler struct {
	service *application.DocsService
}

func NewDocsHandler(service *application.DocsService) *DocsHandler {
	return &DocsHandler{service: service}
}

// ── handlers ──────────────────────────────────────────────────────────────────

func (h *DocsHandler) GetManifest(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")
	role := r.Context().Value("role")

	if role == "USER" {
		data, err := h.service.GetManifest(false)
		if err != nil {
			log.Printf("[Handler:GetManifest] Service call, requestID %s  error %s", requestID, err.Error())
			utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get public docs manifest"))
			return
		}

		utils.WriteJSONSucces(w, http.StatusOK, "Fetched documents successfully", map[string]interface{}{
			"service":    data.Service,
			"apiVersion": data.APIVersion,
			"scope":      "public",
			"internal":   []application.DocCategory{},
			"public":     data.Public,
		})
		return
	}

	// For administrative/internal roles, return both public and internal manifests
	publicData, err := h.service.GetManifest(false)
	if err != nil {
		log.Printf("[Handler:GetManifest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get public docs manifest"))
		return
	}

	internalData, err := h.service.GetManifest(true)
	if err != nil {
		log.Printf("[Handler:GetManifest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to get private docs manifest"))
		return
	}

	utils.WriteJSONSucces(w, http.StatusOK, "Fetched documents successfully", map[string]interface{}{
		"service":    chooseString(publicData, internalData),
		"apiVersion": chooseVersion(publicData, internalData),
		"scope":      "internal",
		"internal":   safeCategories(internalData),
		"public":     safeCategories(publicData),
	})

}

func (h *DocsHandler) GetDoc(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestId")

	slug := r.PathValue("slug")
	doc, err := h.service.GetDoc(slug, false)
	if err != nil {
		log.Printf("[Handler:GetManifest] Service call, requestID %s  error %s", requestID, err.Error())
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Errorf("Could not find public doc"))
		return
	}
	utils.WriteJSONSucces(w, http.StatusOK, "Fetched Public manifest succesfully", doc)

}

// Helpers to safely extract fields when one of the manifests might be nil
func safeCategories(m *application.DocManifest) []application.DocCategory {
	if m == nil {
		return []application.DocCategory{}
	}
	// prefer Public slice if present, otherwise Internal
	if len(m.Public) > 0 {
		return m.Public
	}
	if len(m.Internal) > 0 {
		return m.Internal
	}
	return []application.DocCategory{}
}

func chooseString(a, b *application.DocManifest) string {
	if a != nil && a.Service != "" {
		return a.Service
	}
	if b != nil {
		return b.Service
	}
	return ""
}

func chooseVersion(a, b *application.DocManifest) string {
	if a != nil && a.APIVersion != "" {
		return a.APIVersion
	}
	if b != nil {
		return b.APIVersion
	}
	return ""
}

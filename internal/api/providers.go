package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofciepka/agentic-framework/internal/db"
	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func (s *Server) handleListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := db.GetProviders(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if providers == nil {
		providers = []model.Provider{}
	}
	for i := range providers {
		providers[i].APIKeyEncrypted = nil
	}
	writeJSON(w, http.StatusOK, providers)
}

func (s *Server) handleCreateProvider(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		BaseURL string `json:"base_url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Name == "" || req.BaseURL == "" || req.APIKey == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name, base_url, and api_key are required")
		return
	}

	encKey := encryptAPIKey(req.APIKey)
	p := &model.Provider{
		Name:            req.Name,
		BaseURL:         req.BaseURL,
		APIKeyEncrypted: encKey,
	}
	id, err := db.InsertProvider(s.db, p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	p.ID = id
	p.APIKeyEncrypted = nil
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid provider id")
		return
	}

	existing, err := db.GetProvider(s.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "provider not found")
		return
	}

	var req struct {
		Name   string `json:"name"`
		BaseURL string `json:"base_url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	name := req.Name
	if name == "" {
		name = existing.Name
	}
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = existing.BaseURL
	}
	encKey := existing.APIKeyEncrypted
	if req.APIKey != "" {
		encKey = encryptAPIKey(req.APIKey)
	}

	if err := db.UpdateProvider(s.db, id, name, baseURL, encKey); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid provider id")
		return
	}
	if err := db.DeleteProvider(s.db, id); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

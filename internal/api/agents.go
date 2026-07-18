package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofciepka/agentic-framework/internal/db"
	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := db.GetAgents(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if agents == nil {
		agents = []model.Agent{}
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string       `json:"name"`
		SystemPrompt string       `json:"system_prompt"`
		ProviderID   int64        `json:"provider_id"`
		Model        string       `json:"model"`
		Temperature  float64      `json:"temperature"`
		MaxTokens    int          `json:"max_tokens"`
		Tools        []model.Tool `json:"tools"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Name == "" || req.SystemPrompt == "" || req.Model == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name, system_prompt, and model are required")
		return
	}

	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}
	if req.Tools == nil {
		req.Tools = []model.Tool{}
	}

	ag := &model.Agent{
		Name:         req.Name,
		SystemPrompt: req.SystemPrompt,
		ProviderID:   req.ProviderID,
		Model:        req.Model,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Tools:        req.Tools,
	}
	id, err := db.InsertAgent(s.db, ag)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	ag.ID = id
	writeJSON(w, http.StatusCreated, ag)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid agent id")
		return
	}

	ag, err := db.GetAgent(s.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if ag == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, ag)
}

func (s *Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid agent id")
		return
	}

	existing, err := db.GetAgent(s.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	var req struct {
		Name         string       `json:"name"`
		SystemPrompt string       `json:"system_prompt"`
		ProviderID   int64        `json:"provider_id"`
		Model        string       `json:"model"`
		Temperature  float64      `json:"temperature"`
		MaxTokens    int          `json:"max_tokens"`
		Tools        []model.Tool `json:"tools"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.SystemPrompt != "" {
		existing.SystemPrompt = req.SystemPrompt
	}
	if req.Model != "" {
		existing.Model = req.Model
	}
	if req.ProviderID != 0 {
		existing.ProviderID = req.ProviderID
	}
	if req.Temperature != 0 {
		existing.Temperature = req.Temperature
	}
	if req.MaxTokens != 0 {
		existing.MaxTokens = req.MaxTokens
	}
	if req.Tools != nil {
		existing.Tools = req.Tools
	}

	if existing.Temperature == 0 {
		existing.Temperature = 0.7
	}
	if existing.MaxTokens == 0 {
		existing.MaxTokens = 4096
	}

	if err := db.UpdateAgent(s.db, id, existing); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (s *Server) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid agent id")
		return
	}
	if err := db.DeleteAgent(s.db, id); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofciepka/agentic-framework/internal/db"
	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

type conversationResponse struct {
	model.Conversation
	Messages []model.Message `json:"messages"`
}

func (s *Server) handleListAgentConversations(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := strconv.ParseInt(agentIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid agent id")
		return
	}

	convs, err := db.GetConversationsByAgent(s.db, agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if convs == nil {
		convs = []model.Conversation{}
	}
	writeJSON(w, http.StatusOK, convs)
}

func (s *Server) handleCreateConversation(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := strconv.ParseInt(agentIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid agent id")
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	conv := &model.Conversation{
		AgentID: agentID,
		Title:   req.Title,
	}
	id, err := db.InsertConversation(s.db, conv)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	conv.ID = id
	writeJSON(w, http.StatusCreated, conv)
}

func (s *Server) handleGetConversation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid conversation id")
		return
	}

	conv, err := db.GetConversation(s.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if conv == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "conversation not found")
		return
	}

	messages, err := db.GetMessages(s.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if messages == nil {
		messages = []model.Message{}
	}

	resp := conversationResponse{
		Conversation: *conv,
		Messages:     messages,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDeleteConversation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid conversation id")
		return
	}
	if err := db.DeleteConversation(s.db, id); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

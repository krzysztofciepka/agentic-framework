package api

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofciepka/agentic-framework/internal/agent"
	"github.com/krzysztofciepka/agentic-framework/internal/db"
	"github.com/krzysztofciepka/agentic-framework/internal/llm"
	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

var encryptionKey = []byte("0123456789abcdef0123456789abcdef")

func encryptAPIKey(plaintext string) []byte {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		log.Printf("encrypt: new cipher: %v", err)
		return nil
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("encrypt: new gcm: %v", err)
		return nil
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("encrypt: nonce: %v", err)
		return nil
	}
	return aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
}

func decryptAPIKey(data []byte) string {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		log.Printf("decrypt: new cipher: %v", err)
		return ""
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("decrypt: new gcm: %v", err)
		return ""
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		log.Printf("decrypt: ciphertext too short")
		return ""
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("decrypt: open: %v", err)
		return ""
	}
	return string(plaintext)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	convID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid conversation id")
		return
	}

	var req struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}

	conv, err := db.GetConversation(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if conv == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "conversation not found")
		return
	}

	ag, err := db.GetAgent(s.db, conv.AgentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if ag == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	prov, err := db.GetProvider(s.db, ag.ProviderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if prov == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "provider not found")
		return
	}

	existingMsgs, err := db.GetMessages(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	isFirstMessage := len(existingMsgs) == 0

	userMsg := model.Message{
		ConversationID: convID,
		Role:           req.Role,
		Content:        req.Content,
	}
	if _, err := db.InsertMessage(s.db, &userMsg); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	if isFirstMessage && conv.Title == "" {
		title := req.Content
		if len(title) > 100 {
			title = title[:100]
		}
		db.UpdateConversationTitle(s.db, convID, title)
	}

	allMsgs, err := db.GetMessages(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	llmMsgs := []llm.Message{
		{Role: "system", Content: ag.SystemPrompt},
	}
	llmMsgs = append(llmMsgs, llm.ModelMessagesToLLM(allMsgs)...)

	toolNames := make([]string, len(ag.Tools))
	for i, t := range ag.Tools {
		toolNames[i] = t.Name
	}

	apiKey := decryptAPIKey(prov.APIKeyEncrypted)
	client := llm.NewClient(prov.BaseURL, apiKey, ag.Model)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	originalLen := len(llmMsgs)
	results, err := s.orchestrator.Run(ctx, client, llmMsgs, toolNames, ag.Temperature, ag.MaxTokens)
	if err != nil {
		writeError(w, http.StatusBadGateway, "LLM_ERROR", err.Error())
		return
	}

	toolCallName := make(map[string]string)
	for _, res := range results {
		for _, tc := range res.ToolCalls {
			if tc.Function != nil {
				toolCallName[tc.ID] = tc.Function.Name
			}
		}
	}

	for _, llmMsg := range llmMsgs[originalLen:] {
		dbMsg := model.Message{
			ConversationID: convID,
			Role:           llmMsg.Role,
			Content:        llmMsg.Content,
		}
		if llmMsg.Role == "assistant" && len(llmMsg.ToolCalls) > 0 {
			tcJSON, _ := json.Marshal(llmMsg.ToolCalls)
			dbMsg.Content = string(tcJSON)
		}
		if llmMsg.Role == "tool" {
			dbMsg.ToolCallID = llmMsg.ToolCallID
			dbMsg.ToolName = toolCallName[llmMsg.ToolCallID]
		}
		db.InsertMessage(s.db, &dbMsg)
	}

	db.TouchConversation(s.db, convID)

	if len(results) == 0 {
		writeError(w, http.StatusBadGateway, "LLM_ERROR", "no results from LLM")
		return
	}

	lastAssistant := results[len(results)-1]
	writeJSON(w, http.StatusOK, map[string]string{
		"role":    lastAssistant.Role,
		"content": lastAssistant.Content,
	})
}

func (s *Server) handleStreamMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	convID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid conversation id")
		return
	}

	var req struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}

	conv, err := db.GetConversation(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if conv == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "conversation not found")
		return
	}

	ag, err := db.GetAgent(s.db, conv.AgentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if ag == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	prov, err := db.GetProvider(s.db, ag.ProviderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if prov == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "provider not found")
		return
	}

	existingMsgs, err := db.GetMessages(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	isFirstMessage := len(existingMsgs) == 0

	userMsg := model.Message{
		ConversationID: convID,
		Role:           req.Role,
		Content:        req.Content,
	}
	if _, err := db.InsertMessage(s.db, &userMsg); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	if isFirstMessage && conv.Title == "" {
		title := req.Content
		if len(title) > 100 {
			title = title[:100]
		}
		db.UpdateConversationTitle(s.db, convID, title)
	}

	allMsgs, err := db.GetMessages(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	llmMsgs := []llm.Message{
		{Role: "system", Content: ag.SystemPrompt},
	}
	llmMsgs = append(llmMsgs, llm.ModelMessagesToLLM(allMsgs)...)

	toolNames := make([]string, len(ag.Tools))
	for i, t := range ag.Tools {
		toolNames[i] = t.Name
	}

	apiKey := decryptAPIKey(prov.APIKeyEncrypted)
	client := llm.NewClient(prov.BaseURL, apiKey, ag.Model)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusNotImplemented, "STREAMING_NOT_SUPPORTED", "streaming not supported")
		return
	}

	eventCh := make(chan agent.StreamEvent)
	go func() {
		s.orchestrator.RunStream(ctx, client, llmMsgs, toolNames, ag.Temperature, ag.MaxTokens, eventCh)
	}()

	var accumulatedContent string
	for event := range eventCh {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		if event.Type == "content" {
			accumulatedContent += event.Content
		}
	}

	if accumulatedContent != "" {
		assistantMsg := model.Message{
			ConversationID: convID,
			Role:           "assistant",
			Content:        accumulatedContent,
		}
		db.InsertMessage(s.db, &assistantMsg)
	}

	db.TouchConversation(s.db, convID)
}

func (s *Server) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	convID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid conversation id")
		return
	}

	messages, err := db.GetMessages(s.db, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if messages == nil {
		messages = []model.Message{}
	}
	writeJSON(w, http.StatusOK, messages)
}

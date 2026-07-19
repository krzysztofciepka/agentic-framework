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
		Role    string   `json:"role"`
		Content string   `json:"content"`
		Images  []string `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}
	if req.Role != "user" {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE", "only 'user' role is accepted")
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

	if len(req.Images) > 0 {
		blocks := []llm.ContentBlock{{Type: "text", Text: req.Content}}
		for _, imgURL := range req.Images {
			blocks = append(blocks, llm.ContentBlock{
				Type:     "image_url",
				ImageURL: &llm.ImageURLBlock{URL: "https://" + r.Host + imgURL, Detail: "auto"},
			})
		}
		llmMsgs[len(llmMsgs)-1].Content = blocks
	}

	toolNames := make([]string, len(ag.Tools))
	for i, t := range ag.Tools {
		toolNames[i] = t.Name
	}

	apiKey := decryptAPIKey(prov.APIKeyEncrypted)
	client := llm.NewClient(prov.BaseURL, apiKey, ag.Model)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	originalLen := len(llmMsgs)
	results, fullMsgs, err := s.orchestrator.Run(ctx, client, llmMsgs, toolNames, ag.Temperature, ag.MaxTokens)
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

	toolCallNameBuild := make(map[string]string)
	for _, llmMsg := range fullMsgs[originalLen:] {
		if llmMsg.Role == "assistant" {
			for _, tc := range llmMsg.ToolCalls {
				if tc.Function != nil {
					toolCallNameBuild[tc.ID] = tc.Function.Name
				}
			}
		}
	}
	for k, v := range toolCallNameBuild {
		toolCallName[k] = v
	}

	for _, llmMsg := range fullMsgs[originalLen:] {
		contentStr, _ := llmMsg.Content.(string)
		dbMsg := model.Message{
			ConversationID: convID,
			Role:           llmMsg.Role,
			Content:        contentStr,
		}
		if llmMsg.Role == "tool" {
			dbMsg.ToolCallID = llmMsg.ToolCallID
			dbMsg.ToolName = toolCallName[llmMsg.ToolCallID]
		}
		if _, err := db.InsertMessage(s.db, &dbMsg); err != nil {
			log.Printf("insert message: %v", err)
		}
	}

	if err := db.TouchConversation(s.db, convID); err != nil {
		log.Printf("touch conversation: %v", err)
	}

	if len(results) == 0 {
		writeError(w, http.StatusBadGateway, "LLM_ERROR", "no results from LLM")
		return
	}

	lastAssistant := results[len(results)-1]
	lastContent, _ := lastAssistant.Content.(string)
	writeJSON(w, http.StatusOK, map[string]string{
		"role":    lastAssistant.Role,
		"content": lastContent,
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
		Role    string   `json:"role"`
		Content string   `json:"content"`
		Images  []string `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}
	if req.Role != "user" {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE", "only 'user' role is accepted")
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

	if len(req.Images) > 0 {
		blocks := []llm.ContentBlock{{Type: "text", Text: req.Content}}
		for _, imgURL := range req.Images {
			blocks = append(blocks, llm.ContentBlock{
				Type:     "image_url",
				ImageURL: &llm.ImageURLBlock{URL: "https://" + r.Host + imgURL, Detail: "auto"},
			})
		}
		llmMsgs[len(llmMsgs)-1].Content = blocks
	}

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

	originalLen := len(llmMsgs)

	eventCh := make(chan agent.StreamEvent)
	resCh := make(chan struct {
		msgs []llm.Message
		err  error
	}, 1)

	go func() {
		msgs, err := s.orchestrator.RunStream(ctx, client, llmMsgs, toolNames, ag.Temperature, ag.MaxTokens, eventCh)
		resCh <- struct {
			msgs []llm.Message
			err  error
		}{msgs, err}
	}()

	for event := range eventCh {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	streamResult := <-resCh

	if streamResult.err != nil {
		writeError(w, http.StatusBadGateway, "LLM_ERROR", streamResult.err.Error())
		return
	}

	toolCallName := make(map[string]string)
	for _, llmMsg := range streamResult.msgs[originalLen:] {
		if llmMsg.Role == "assistant" {
			for _, tc := range llmMsg.ToolCalls {
				if tc.Function != nil {
					toolCallName[tc.ID] = tc.Function.Name
				}
			}
		}
	}

	for _, llmMsg := range streamResult.msgs[originalLen:] {
		contentStr, _ := llmMsg.Content.(string)
		dbMsg := model.Message{
			ConversationID: convID,
			Role:           llmMsg.Role,
			Content:        contentStr,
		}
		if llmMsg.Role == "tool" {
			dbMsg.ToolCallID = llmMsg.ToolCallID
			dbMsg.ToolName = toolCallName[llmMsg.ToolCallID]
		}
		if _, err := db.InsertMessage(s.db, &dbMsg); err != nil {
			log.Printf("insert message: %v", err)
		}
	}

	if err := db.TouchConversation(s.db, convID); err != nil {
		log.Printf("touch conversation: %v", err)
	}
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

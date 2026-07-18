package db

import (
	"database/sql"
	"testing"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&_foreign_keys=on")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func TestProviders(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	p := &model.Provider{
		Name:    "OpenAI",
		BaseURL: "https://api.openai.com",
		APIKeyEncrypted: []byte("encrypted-key"),
	}

	id, err := InsertProvider(db, p)
	if err != nil {
		t.Fatalf("InsertProvider: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}

	got, err := GetProvider(db, id)
	if err != nil {
		t.Fatalf("GetProvider: %v", err)
	}
	if got == nil {
		t.Fatal("expected provider, got nil")
	}
	if got.Name != "OpenAI" {
		t.Errorf("expected name OpenAPI, got %s", got.Name)
	}

	providers, err := GetProviders(db)
	if err != nil {
		t.Fatalf("GetProviders: %v", err)
	}
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}

	err = UpdateProvider(db, id, "Anthropic", "https://api.anthropic.com", []byte("new-enc-key"))
	if err != nil {
		t.Fatalf("UpdateProvider: %v", err)
	}

	got, err = GetProvider(db, id)
	if err != nil {
		t.Fatalf("GetProvider after update: %v", err)
	}
	if got.Name != "Anthropic" {
		t.Errorf("expected name Anthropic, got %s", got.Name)
	}

	err = DeleteProvider(db, id)
	if err != nil {
		t.Fatalf("DeleteProvider: %v", err)
	}

	got, err = GetProvider(db, id)
	if err != nil {
		t.Fatalf("GetProvider after delete: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after delete, got %v", got)
	}
}

func TestAgentsWithTools(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pid, err := InsertProvider(db, &model.Provider{
		Name: "OpenAI", BaseURL: "https://api.openai.com", APIKeyEncrypted: []byte("key"),
	})
	if err != nil {
		t.Fatalf("InsertProvider: %v", err)
	}

	err = UpsertTool(db, &model.Tool{Name: "web_search", Description: "Search the web", Category: "search"})
	if err != nil {
		t.Fatalf("UpsertTool 1: %v", err)
	}
	err = UpsertTool(db, &model.Tool{Name: "calculator", Description: "Do math", Category: "math"})
	if err != nil {
		t.Fatalf("UpsertTool 2: %v", err)
	}

	tools, err := GetTools(db)
	if err != nil {
		t.Fatalf("GetTools: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	agent := &model.Agent{
		Name:        "TestAgent",
		SystemPrompt: "You are helpful.",
		ProviderID:  pid,
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		Tools:       tools,
	}

	aid, err := InsertAgent(db, agent)
	if err != nil {
		t.Fatalf("InsertAgent: %v", err)
	}

	got, err := GetAgent(db, aid)
	if err != nil {
		t.Fatalf("GetAgent: %v", err)
	}
	if got == nil {
		t.Fatal("expected agent, got nil")
	}
	if len(got.Tools) != 2 {
		t.Errorf("expected 2 tools on agent, got %d", len(got.Tools))
	}

	agents, err := GetAgents(db)
	if err != nil {
		t.Fatalf("GetAgents: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
	if len(agents[0].Tools) != 2 {
		t.Errorf("expected 2 tools on listed agent, got %d", len(agents[0].Tools))
	}
}

func TestConversationsAndMessages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pid, err := InsertProvider(db, &model.Provider{
		Name: "OpenAI", BaseURL: "https://api.openai.com", APIKeyEncrypted: []byte("key"),
	})
	if err != nil {
		t.Fatalf("InsertProvider: %v", err)
	}

	aid, err := InsertAgent(db, &model.Agent{
		Name: "Agent", SystemPrompt: "sys", ProviderID: pid, Model: "gpt-4", Temperature: 0.5, MaxTokens: 2048,
	})
	if err != nil {
		t.Fatalf("InsertAgent: %v", err)
	}

	cid, err := InsertConversation(db, &model.Conversation{
		AgentID: aid, Title: "Test Chat",
	})
	if err != nil {
		t.Fatalf("InsertConversation: %v", err)
	}

	msg1 := &model.Message{
		ConversationID: cid, Role: "user", Content: "Hello",
	}
	mid1, err := InsertMessage(db, msg1)
	if err != nil {
		t.Fatalf("InsertMessage 1: %v", err)
	}
	if mid1 != 1 {
		t.Errorf("expected message id 1, got %d", mid1)
	}

	msg2 := &model.Message{
		ConversationID: cid, Role: "assistant", Content: "Hi there!",
	}
	_, err = InsertMessage(db, msg2)
	if err != nil {
		t.Fatalf("InsertMessage 2: %v", err)
	}

	messages, err := GetMessages(db, cid)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("expected first message role 'user', got %s", messages[0].Role)
	}

	convs, err := GetConversationsByAgent(db, aid)
	if err != nil {
		t.Fatalf("GetConversationsByAgent: %v", err)
	}
	if len(convs) != 1 {
		t.Errorf("expected 1 conversation, got %d", len(convs))
	}

	conv, err := GetConversation(db, cid)
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if conv.Title != "Test Chat" {
		t.Errorf("expected title 'Test Chat', got %s", conv.Title)
	}

	err = DeleteConversation(db, cid)
	if err != nil {
		t.Fatalf("DeleteConversation: %v", err)
	}

	conv, err = GetConversation(db, cid)
	if err != nil {
		t.Fatalf("GetConversation after delete: %v", err)
	}
	if conv != nil {
		t.Errorf("expected nil after delete, got %v", conv)
	}
}

func TestSettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := UpsertSetting(db, "theme", "dark")
	if err != nil {
		t.Fatalf("UpsertSetting insert: %v", err)
	}

	val, err := GetSetting(db, "theme")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if val != "dark" {
		t.Errorf("expected 'dark', got '%s'", val)
	}

	err = UpsertSetting(db, "theme", "light")
	if err != nil {
		t.Fatalf("UpsertSetting update: %v", err)
	}

	val, err = GetSetting(db, "theme")
	if err != nil {
		t.Fatalf("GetSetting after update: %v", err)
	}
	if val != "light" {
		t.Errorf("expected 'light', got '%s'", val)
	}

	settings, err := GetSettings(db)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if len(settings) != 1 {
		t.Errorf("expected 1 setting, got %d", len(settings))
	}
}

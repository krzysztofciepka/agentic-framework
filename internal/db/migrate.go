package db

import (
	"database/sql"
	"fmt"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS providers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		base_url TEXT NOT NULL,
		api_key_encrypted BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		system_prompt TEXT NOT NULL,
		provider_id INTEGER REFERENCES providers(id),
		model TEXT NOT NULL,
		temperature REAL DEFAULT 0.7,
		max_tokens INTEGER DEFAULT 4096,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS tools (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		category TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS agent_tools (
		agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
		tool_id INTEGER REFERENCES tools(id) ON DELETE CASCADE,
		PRIMARY KEY (agent_id, tool_id)
	)`,
	`CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
		title TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		tool_call_id TEXT,
		tool_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	)`,
	`CREATE INDEX IF NOT EXISTS idx_agents_provider ON agents(provider_id)`,
	`CREATE INDEX IF NOT EXISTS idx_conversations_agent ON conversations(agent_id)`,
	`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id)`,
}

func Migrate(db *sql.DB) error {
	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}

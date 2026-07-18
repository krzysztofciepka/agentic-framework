package db

import (
	"database/sql"
	"fmt"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func InsertMessage(db *sql.DB, m *model.Message) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO messages (conversation_id, role, content, tool_call_id, tool_name) VALUES (?, ?, ?, ?, ?)",
		m.ConversationID, m.Role, m.Content, nullableString(m.ToolCallID), nullableString(m.ToolName),
	)
	if err != nil {
		return 0, fmt.Errorf("insert message: %w", err)
	}
	return res.LastInsertId()
}

func GetMessages(db *sql.DB, conversationID int64) ([]model.Message, error) {
	rows, err := db.Query(
		"SELECT id, conversation_id, role, content, COALESCE(tool_call_id, '') AS tool_call_id, COALESCE(tool_name, '') AS tool_name, created_at FROM messages WHERE conversation_id = ? ORDER BY id ASC",
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.ToolCallID, &m.ToolName, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

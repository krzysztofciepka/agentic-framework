package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func InsertConversation(db *sql.DB, c *model.Conversation) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO conversations (agent_id, title) VALUES (?, ?)",
		c.AgentID, c.Title,
	)
	if err != nil {
		return 0, fmt.Errorf("insert conversation: %w", err)
	}
	return res.LastInsertId()
}

func GetConversationsByAgent(db *sql.DB, agentID int64) ([]model.Conversation, error) {
	rows, err := db.Query(
		"SELECT id, agent_id, title, created_at, updated_at FROM conversations WHERE agent_id = ? ORDER BY updated_at DESC",
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("get conversations by agent: %w", err)
	}
	defer rows.Close()

	var conversations []model.Conversation
	for rows.Next() {
		var c model.Conversation
		if err := rows.Scan(&c.ID, &c.AgentID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func GetConversation(db *sql.DB, id int64) (*model.Conversation, error) {
	var c model.Conversation
	err := db.QueryRow(
		"SELECT id, agent_id, title, created_at, updated_at FROM conversations WHERE id = ?",
		id,
	).Scan(&c.ID, &c.AgentID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return &c, nil
}

func UpdateConversationTitle(db *sql.DB, id int64, title string) error {
	_, err := db.Exec(
		"UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?",
		title, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update conversation title: %w", err)
	}
	return nil
}

func TouchConversation(db *sql.DB, id int64) error {
	_, err := db.Exec(
		"UPDATE conversations SET updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("touch conversation: %w", err)
	}
	return nil
}

func DeleteConversation(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}
	return nil
}

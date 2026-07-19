package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func GetAgentTools(db *sql.DB, agentID int64) ([]model.Tool, error) {
	rows, err := db.Query(
		`SELECT t.id, t.name, t.description, t.category
		 FROM tools t
		 INNER JOIN agent_tools at ON t.id = at.tool_id
		 WHERE at.agent_id = ?`, agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("get agent tools: %w", err)
	}
	defer rows.Close()

	var tools []model.Tool
	for rows.Next() {
		var t model.Tool
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Category); err != nil {
			return nil, fmt.Errorf("scan agent tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func InsertAgent(db *sql.DB, a *model.Agent) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO agents (name, system_prompt, provider_id, model, temperature, max_tokens)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		a.Name, a.SystemPrompt, a.ProviderID, a.Model, a.Temperature, a.MaxTokens,
	)
	if err != nil {
		return 0, fmt.Errorf("insert agent: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	for _, tool := range a.Tools {
		_, err := tx.Exec(
			"INSERT OR IGNORE INTO agent_tools (agent_id, tool_id) VALUES (?, ?)",
			id, tool.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("insert agent tool: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}
	return id, nil
}

func GetAgents(db *sql.DB) ([]model.Agent, error) {
	rows, err := db.Query(
		"SELECT id, name, system_prompt, provider_id, model, temperature, max_tokens, created_at, updated_at FROM agents ORDER BY id DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("get agents: %w", err)
	}
	defer rows.Close()

	var agents []model.Agent
	for rows.Next() {
		var a model.Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.SystemPrompt, &a.ProviderID, &a.Model, &a.Temperature, &a.MaxTokens, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		tools, err := GetAgentTools(db, a.ID)
		if err != nil {
			return nil, err
		}
		if tools == nil {
			tools = []model.Tool{}
		}
		a.Tools = tools
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

func GetAgent(db *sql.DB, id int64) (*model.Agent, error) {
	var a model.Agent
	err := db.QueryRow(
		"SELECT id, name, system_prompt, provider_id, model, temperature, max_tokens, created_at, updated_at FROM agents WHERE id = ?",
		id,
	).Scan(&a.ID, &a.Name, &a.SystemPrompt, &a.ProviderID, &a.Model, &a.Temperature, &a.MaxTokens, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get agent: %w", err)
	}
	tools, err := GetAgentTools(db, a.ID)
	if err != nil {
		return nil, err
	}
	if tools == nil {
		tools = []model.Tool{}
	}
	a.Tools = tools
	return &a, nil
}

func UpdateAgent(db *sql.DB, id int64, a *model.Agent) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE agents SET name = ?, system_prompt = ?, provider_id = ?, model = ?, temperature = ?, max_tokens = ?, updated_at = ?
		 WHERE id = ?`,
		a.Name, a.SystemPrompt, a.ProviderID, a.Model, a.Temperature, a.MaxTokens, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update agent: %w", err)
	}

	_, err = tx.Exec("DELETE FROM agent_tools WHERE agent_id = ?", id)
	if err != nil {
		return fmt.Errorf("delete agent tools: %w", err)
	}

	for _, tool := range a.Tools {
		_, err := tx.Exec(
			"INSERT OR IGNORE INTO agent_tools (agent_id, tool_id) VALUES (?, ?)",
			id, tool.ID,
		)
		if err != nil {
			return fmt.Errorf("reinsert agent tool: %w", err)
		}
	}

	return tx.Commit()
}

func DeleteAgent(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	return nil
}

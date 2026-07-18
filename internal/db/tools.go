package db

import (
	"database/sql"
	"fmt"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func GetTools(db *sql.DB) ([]model.Tool, error) {
	rows, err := db.Query("SELECT id, name, description, category FROM tools ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("get tools: %w", err)
	}
	defer rows.Close()

	var tools []model.Tool
	for rows.Next() {
		var t model.Tool
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Category); err != nil {
			return nil, fmt.Errorf("scan tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func UpsertTool(db *sql.DB, t *model.Tool) error {
	_, err := db.Exec(
		`INSERT INTO tools (name, description, category) VALUES (?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET description = excluded.description, category = excluded.category`,
		t.Name, t.Description, t.Category,
	)
	if err != nil {
		return fmt.Errorf("upsert tool: %w", err)
	}
	return nil
}

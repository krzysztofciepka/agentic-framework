package db

import (
	"database/sql"
	"fmt"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
)

func InsertProvider(db *sql.DB, p *model.Provider) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO providers (name, base_url, api_key_encrypted) VALUES (?, ?, ?)",
		p.Name, p.BaseURL, p.APIKeyEncrypted,
	)
	if err != nil {
		return 0, fmt.Errorf("insert provider: %w", err)
	}
	return res.LastInsertId()
}

func GetProviders(db *sql.DB) ([]model.Provider, error) {
	rows, err := db.Query("SELECT id, name, base_url, api_key_encrypted, created_at FROM providers ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("get providers: %w", err)
	}
	defer rows.Close()

	var providers []model.Provider
	for rows.Next() {
		var p model.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKeyEncrypted, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

func GetProvider(db *sql.DB, id int64) (*model.Provider, error) {
	var p model.Provider
	err := db.QueryRow(
		"SELECT id, name, base_url, api_key_encrypted, created_at FROM providers WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKeyEncrypted, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get provider: %w", err)
	}
	return &p, nil
}

func UpdateProvider(db *sql.DB, id int64, name, baseURL string, encKey []byte) error {
	_, err := db.Exec(
		"UPDATE providers SET name = ?, base_url = ?, api_key_encrypted = ? WHERE id = ?",
		name, baseURL, encKey, id,
	)
	if err != nil {
		return fmt.Errorf("update provider: %w", err)
	}
	return nil
}

func DeleteProvider(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM providers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete provider: %w", err)
	}
	return nil
}

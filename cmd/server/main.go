package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/krzysztofciepka/agentic-framework"
	"github.com/krzysztofciepka/agentic-framework/internal/api"
	"github.com/krzysztofciepka/agentic-framework/internal/config"
	"github.com/krzysztofciepka/agentic-framework/internal/db"
	"github.com/krzysztofciepka/agentic-framework/internal/model"
	"github.com/krzysztofciepka/agentic-framework/internal/tool"
)

func main() {
	cfg := config.Load()

	os.MkdirAll(filepath.Dir(cfg.DBPath), 0755)

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	registry := tool.DefaultRegistry()
	for _, t := range registry.All() {
		db.UpsertTool(database, &model.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			Category:    t.Category(),
		})
	}

	server := api.NewServer(database, registry, agentic.StaticFiles)

	log.Printf("Server starting on :%s", cfg.Port)
	log.Printf("API docs: http://localhost:%s/api/docs", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, server.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

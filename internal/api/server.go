package api

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/krzysztofciepka/agentic-framework/internal/agent"
	"github.com/krzysztofciepka/agentic-framework/internal/tool"
	"github.com/krzysztofciepka/agentic-framework/internal/ws"
)

type Server struct {
	db           *sql.DB
	router       chi.Router
	toolRegistry *tool.Registry
	orchestrator *agent.Orchestrator
	wsHub        *ws.Hub
}

func NewServer(db *sql.DB, toolRegistry *tool.Registry, staticFS embed.FS) *Server {
	s := &Server{
		db:           db,
		router:       chi.NewRouter(),
		toolRegistry: toolRegistry,
		orchestrator: agent.NewOrchestrator(toolRegistry),
		wsHub:        ws.NewHub(),
	}

	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(loggingMiddleware)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	s.router.Route("/api", func(r chi.Router) {
		r.Get("/docs", s.handleDocs)
		r.Get("/agents", s.handleListAgents)
		r.Post("/agents", s.handleCreateAgent)
		r.Get("/agents/{id}", s.handleGetAgent)
		r.Put("/agents/{id}", s.handleUpdateAgent)
		r.Delete("/agents/{id}", s.handleDeleteAgent)
		r.Get("/providers", s.handleListProviders)
		r.Post("/providers", s.handleCreateProvider)
		r.Put("/providers/{id}", s.handleUpdateProvider)
		r.Delete("/providers/{id}", s.handleDeleteProvider)
		r.Get("/tools", s.handleListTools)
		r.Get("/conversations/agents/{agentID}", s.handleListAgentConversations)
		r.Post("/conversations/agents/{agentID}", s.handleCreateConversation)
		r.Get("/conversations/{id}", s.handleGetConversation)
		r.Delete("/conversations/{id}", s.handleDeleteConversation)
		r.Post("/conversations/{id}/messages", s.handleSendMessage)
		r.Get("/conversations/{id}/messages", s.handleGetMessages)
		r.Post("/conversations/{id}/stream", s.handleStreamMessage)
		r.Get("/settings", s.handleGetSettings)
		r.Put("/settings", s.handleUpdateSettings)
		r.Get("/ws", s.wsHub.HandleConnection)
	})

	content, err := fs.Sub(staticFS, "web/dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(content))
		s.router.Handle("/*", fileServer)
	}

	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

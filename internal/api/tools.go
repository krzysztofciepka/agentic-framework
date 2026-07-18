package api

import (
	"net/http"

	"github.com/krzysztofciepka/agentic-framework/internal/db"
	"github.com/krzysztofciepka/agentic-framework/internal/model"
	"github.com/krzysztofciepka/agentic-framework/internal/tool"
)

type toolResponse struct {
	model.Tool
	Parameters tool.ToolSchema `json:"parameters"`
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools, err := db.GetTools(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if tools == nil {
		tools = []model.Tool{}
	}

	result := make([]toolResponse, len(tools))
	for i, t := range tools {
		tr := toolResponse{Tool: t}
		if regTool, err := s.toolRegistry.Get(t.Name); err == nil {
			tr.Parameters = regTool.Parameters()
		}
		result[i] = tr
	}
	writeJSON(w, http.StatusOK, result)
}

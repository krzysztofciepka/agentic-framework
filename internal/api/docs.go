package api

import (
	"fmt"
	"net/http"
	"strings"
)

type endpoint struct {
	Method string
	Path   string
	Desc   string
}

var endpoints = []endpoint{
	{"GET", "/api/docs", "API documentation (this page)"},
	{"GET", "/api/agents", "List all agents"},
	{"POST", "/api/agents", "Create a new agent"},
	{"GET", "/api/agents/{id}", "Get agent by ID"},
	{"PUT", "/api/agents/{id}", "Update agent by ID"},
	{"DELETE", "/api/agents/{id}", "Delete agent by ID"},
	{"GET", "/api/providers", "List all LLM providers"},
	{"POST", "/api/providers", "Create a new provider"},
	{"PUT", "/api/providers/{id}", "Update provider by ID"},
	{"DELETE", "/api/providers/{id}", "Delete provider by ID"},
	{"GET", "/api/tools", "List available tools"},
	{"GET", "/api/conversations/agents/{agentID}", "List conversations for an agent"},
	{"POST", "/api/conversations/agents/{agentID}", "Create conversation for an agent"},
	{"GET", "/api/conversations/{id}", "Get conversation by ID"},
	{"DELETE", "/api/conversations/{id}", "Delete conversation by ID"},
	{"POST", "/api/conversations/{id}/messages", "Send a message in a conversation"},
	{"GET", "/api/conversations/{id}/messages", "Get messages for a conversation"},
	{"POST", "/api/conversations/{id}/stream", "Stream an LLM response via SSE"},
	{"GET", "/api/settings", "Get application settings"},
	{"PUT", "/api/settings", "Update application settings"},
	{"GET", "/api/ws", "WebSocket connection (query: conversation_id)"},
}

var methodColors = map[string]string{
	"GET":    "#61affe",
	"POST":   "#49cc90",
	"PUT":    "#fca130",
	"DELETE": "#f93e3e",
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Agentic Framework — API Docs</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
    background: #0d1117;
    color: #c9d1d9;
    padding: 2rem;
  }
  h1 { font-size: 1.5rem; color: #58a6ff; margin-bottom: 0.5rem; }
  p.subtitle { color: #8b949e; margin-bottom: 2rem; }
  .endpoint {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.6rem 0.8rem;
    border-bottom: 1px solid #21262d;
  }
  .endpoint:hover { background: #161b22; }
  .method {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    min-width: 4.2rem;
    text-align: center;
    color: #000;
  }
  .path {
    font-size: 0.85rem;
    color: #c9d1d9;
    min-width: 24rem;
  }
  .desc { font-size: 0.8rem; color: #8b949e; }
</style>
</head>
<body>
<h1>Agentic Framework — API v1</h1>
<p class="subtitle">Base URL: <code style="color:#e6edf3;">/api</code></p>
`)

	for _, e := range endpoints {
		color := methodColors[e.Method]
		method := e.Method
		for len(method) < 6 {
			method = " " + method
		}
		fmt.Fprintf(&b,
			`<div class="endpoint">`+
				`<span class="method" style="background:%s">%s</span>`+
				`<span class="path"><code>%s</code></span>`+
				`<span class="desc">%s</span>`+
				`</div>`+"\n",
			color, method, e.Path, e.Desc,
		)
	}

	b.WriteString(`
</body>
</html>
`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b.String()))
}

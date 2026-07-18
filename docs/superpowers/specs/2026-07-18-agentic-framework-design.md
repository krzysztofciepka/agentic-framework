# Agentic Framework — Design Spec

## Overview

A single-binary Go application for spawning, managing, and chatting with AI agents. Each agent has a system prompt, a model/provider, and a subset of tools from an extensible tool library. A Svelte SPA provides the chat interface, agent management, and settings. Agents are available via REST API for integration with n8n workflows.

## Architecture

```
┌──────────────────────────────────────────────────┐
│                   Go Binary                       │
│                                                   │
│  ┌──────────┐   ┌───────────┐   ┌────────────┐  │
│  │  SPA      │   │  REST API  │   │  WebSocket │  │
│  │  (embed)  │   │  (chi)    │   │  handler   │  │
│  └──────────┘   └─────┬─────┘   └──────┬─────┘  │
│                       │                 │         │
│              ┌────────┴─────────┐       │         │
│              │   Agent Manager  │◄──────┘         │
│              │   Chat Manager   │                 │
│              │   Tool Registry  │                 │
│              └────────┬─────────┘                 │
│                       │                           │
│              ┌────────┴─────────┐                 │
│              │   LLM Client     │                 │
│              │   (OpenAI API)   │                 │
│              └──────────────────┘                 │
│                       │                           │
│              ┌────────┴─────────┐                 │
│              │     SQLite       │                 │
│              └──────────────────┘                 │
└──────────────────────────────────────────────────┘
```

Single Go binary. SPA built with Vite + Svelte, embedded via `//go:embed`. HTTP router (chi) serves both the API and the embedded SPA. WebSocket for streaming LLM responses. Tools execute in-process; dangerous tools (`run_cmd`, `file_system_write`, `execute`) run via `os/exec` with timeout, output limits, and directory confinement.

## Data Model

```
providers
  id INTEGER PRIMARY KEY
  name TEXT NOT NULL                  -- e.g. "opencode"
  base_url TEXT NOT NULL              -- OpenAI-compatible API URL
  api_key_encrypted BLOB NOT NULL     -- AES-GCM encrypted
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP

agents
  id INTEGER PRIMARY KEY
  name TEXT NOT NULL
  system_prompt TEXT NOT NULL
  provider_id INTEGER REFERENCES providers(id)
  model TEXT NOT NULL                 -- e.g. "gpt-4o"
  temperature REAL DEFAULT 0.7
  max_tokens INTEGER DEFAULT 4096
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP

tools
  id INTEGER PRIMARY KEY
  name TEXT NOT NULL UNIQUE           -- e.g. "web_search"
  description TEXT NOT NULL
  category TEXT NOT NULL              -- "web", "system", "file"

agent_tools
  agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE
  tool_id INTEGER REFERENCES tools(id) ON DELETE CASCADE
  PRIMARY KEY (agent_id, tool_id)

conversations
  id INTEGER PRIMARY KEY
  agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE
  title TEXT                          -- first user message or custom
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP

messages
  id INTEGER PRIMARY KEY
  conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE
  role TEXT NOT NULL                  -- "system", "user", "assistant", "tool"
  content TEXT NOT NULL
  tool_call_id TEXT                   -- for tool role messages
  tool_name TEXT                      -- for tool role messages
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP

settings
  key TEXT PRIMARY KEY
  value TEXT
```

- Tools are seeded into the `tools` table on startup from registered Go tool implementations
- `agent_tools` is the subset of tools assigned to each agent
- API keys encrypted at rest with AES-GCM; encryption key stored in a config file or env var
- `settings` table for global app config (default provider, port, etc.)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/agents` | List all agents |
| POST | `/api/agents` | Create agent |
| GET | `/api/agents/:id` | Get agent with tools |
| PUT | `/api/agents/:id` | Update agent |
| DELETE | `/api/agents/:id` | Delete agent |
| GET | `/api/providers` | List providers |
| POST | `/api/providers` | Create provider |
| PUT | `/api/providers/:id` | Update provider |
| DELETE | `/api/providers/:id` | Delete provider |
| GET | `/api/tools` | List all tools in library |
| GET | `/api/agents/:id/conversations` | List conversations for agent |
| POST | `/api/agents/:id/conversations` | Create conversation |
| GET | `/api/conversations/:id` | Get conversation with messages |
| DELETE | `/api/conversations/:id` | Delete conversation |
| POST | `/api/conversations/:id/messages` | Send message, get LLM response |
| GET | `/api/conversations/:id/messages` | Get messages (paginated) |
| WS | `/api/conversations/:id/stream` | Stream LLM response |
| GET | `/api/settings` | Get global settings |
| PUT | `/api/settings` | Update global settings |
| GET | `/api/docs` | Interactive API documentation |

### Chat Flow

1. Client sends `POST /api/conversations/:id/messages` with `{ "role": "user", "content": "..." }`
2. Server loads conversation history + agent config (system prompt, model, tools)
3. Sends to OpenAI-compatible API with tools definitions
4. If LLM returns `tool_calls`, server executes each tool, appends tool result messages, loops back
5. Returns final assistant response (or streams via WebSocket)
6. All messages persisted to SQLite

## Tool System

```go
type Tool interface {
    Name() string
    Description() string
    Category() string
    Parameters() ToolSchema  // JSON Schema for function calling
    Execute(ctx context.Context, args map[string]any) (string, error)
}
```

### Built-in Tools

| Tool | Category | Description |
|------|----------|-------------|
| `web_search` | web | Search the web using Brave Search API |
| `web_fetch` | web | Fetch URL content (text/markdown/html) |
| `run_cmd` | system | Execute shell command |
| `file_system_read` | file | Read file contents |
| `file_system_write` | file | Write content to file |
| `execute` | system | Run a binary with arguments |

### Danger Tool Safety

`run_cmd`, `execute`, and `file_system_write` execute via `os/exec` with:
- 30s timeout (configurable)
- 1MB max output (configurable)
- Working directory confined to `ALLOWED_DIR`
- All tool executions logged

### Extending

New tools implement the `Tool` interface and register at init:

```go
func init() {
    registry.Register(&MyTool{})
}
```

## Frontend (SPA)

- **Framework:** Svelte 5 with SvelteKit, `shadcn-svelte` component library
- Built with Vite, output embedded via `//go:embed web/dist`
- Three main pages: Agents, Chat, Settings

### Pages

**Agents page:**
- Table/list of all agents with name, model, tool count
- Create/edit form: name, system prompt, provider dropdown, model, temperature, max tokens
- Tool assignment: checkboxes from tools library
- Delete with confirmation

**Chat page:**
- Left sidebar: conversation list for selected agent, "New conversation" button
- Main area: chat messages with role labels, streaming indicator for in-progress responses
- Tool calls rendered as expandable cards showing tool name, arguments, result
- Message input at bottom, send button
- Agent selector dropdown at top

**Settings page:**
- Providers CRUD: name, base URL, API key (masked on display)
- Global defaults: default provider, default model
- Tool library overview (read-only list)

## Project Structure

```
agentic-framework/
├── cmd/server/main.go
├── internal/
│   ├── api/          — HTTP handlers, middleware, router
│   ├── agent/        — agent orchestrator (chat loop with tool calling)
│   ├── db/           — SQLite migrations, query functions
│   ├── llm/          — OpenAI-compatible HTTP client
│   ├── tool/         — tool interface, registry, built-in implementations
│   ├── model/        — domain types (Agent, Provider, Message, etc.)
│   └── ws/           — WebSocket hub for streaming
├── web/
│   ├── src/          — SvelteKit app
│   ├── static/
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
├── Makefile
├── go.mod
└── docs/superpowers/specs/
```

## Error Handling

- API returns JSON errors: `{ "error": "message", "code": "ERROR_CODE" }`
- LLM failures result in 502 with the upstream error message
- Tool execution failures are returned as tool result messages to the LLM (not as API errors)
- SQLite constraint violations return 409 Conflict
- Not found returns 404

## Testing

- Unit tests for tool implementations, LLM client, agent loop
- Integration tests for API endpoints with in-memory SQLite
- Frontend: component tests with Vitest, e2e with Playwright (post-MVP)

## Scope for MVP (v0.1.0)

Must have:
- Go server with all API endpoints + WebSocket streaming
- SQLite persistence with migrations
- Tool system with all 6 built-in tools
- Svelte SPA with Agents, Chat, and Settings pages
- API documentation page
- Single binary build with embedded frontend

Out of scope for v0.1.0:
- Frontend tests (e2e/component)
- Advanced tool sandboxing (chroot, containers)
- User authentication / multi-user
- Tool execution history/audit UI

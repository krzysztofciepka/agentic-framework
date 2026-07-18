# Agentic Framework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single-binary Go application with embedded Svelte SPA for spawning and chatting with AI agents that have configurable system prompts, models, providers, and tools.

**Architecture:** Monolithic Go server using chi router, SQLite for persistence, OpenAI-compatible LLM client, pluggable tool system with 6 built-in tools, WebSocket for streaming, embedded SvelteKit SPA.

**Tech Stack:** Go 1.22+, chi, mattn/go-sqlite3, gorilla/websocket, Svelte 5 + SvelteKit + shadcn-svelte, Vite

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `internal/model/models.go`
- Create: `Makefile`
- Create: `.gitignore`

- [ ] **Step 1: Initialize Go module**

Run: `go mod init github.com/krzysztofciepka/agentic-framework`
Expected: creates `go.mod`

- [ ] **Step 2: Define domain models**

Write `internal/model/models.go` — all domain types (Provider, Agent, Tool, Conversation, Message, Setting) with JSON tags. See spec data model section for fields. Agent has `Tools []Tool` for the assigned subset, API key fields tagged `json:"-"` to prevent leaks.

- [ ] **Step 3: Write empty main entrypoint** — `cmd/server/main.go` with empty `func main() {}`

- [ ] **Step 4: Write .gitignore** — ignore `bin/`, `web/dist/`, `web/node_modules/`, `*.db`, `.env`

- [ ] **Step 5: Tidy and build** — `go mod tidy && go build ./cmd/server`

- [ ] **Step 6: Commit** — `feat: project scaffolding with domain models`

---

### Task 2: Database Layer

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrate.go`

- [ ] **Step 1: Add sqlite3 dependency** — `go get github.com/mattn/go-sqlite3`

- [ ] **Step 2: Write database connection** (`internal/db/db.go`) — `Open(path)` function using `sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")` with ping.

- [ ] **Step 3: Write migrations** (`internal/db/migrate.go`) — `Migrate(db)` function with all 7 CREATE TABLE IF NOT EXISTS statements matching the spec data model (providers, agents, tools, agent_tools, conversations, messages, settings).

- [ ] **Step 4: Commit** — `feat: database layer with SQLite and migrations`

---

### Task 3: Database Query Layer

**Files:**
- Create: `internal/db/providers.go`
- Create: `internal/db/agents.go`
- Create: `internal/db/tools.go`
- Create: `internal/db/conversations.go`
- Create: `internal/db/messages.go`
- Create: `internal/db/settings.go`
- Create: `internal/db/queries_test.go`

- [ ] **Step 1: Write providers CRUD** (`internal/db/providers.go`) — InsertProvider, GetProviders, GetProvider, UpdateProvider, DeleteProvider. UpdateProvider takes name, baseURL, encKey params. All use parameterized queries.

- [ ] **Step 2: Write agents CRUD with tools join** (`internal/db/agents.go`) — InsertAgent (inserts agent + agent_tools rows), GetAgents (joins tools), GetAgent (single + tools), UpdateAgent (replaces tool assignments), DeleteAgent (cascades), GetAgentTools (join query).

- [ ] **Step 3: Write tools queries** (`internal/db/tools.go`) — GetTools (list all), UpsertTool (INSERT ON CONFLICT for seeding).

- [ ] **Step 4: Write conversations CRUD** (`internal/db/conversations.go`) — InsertConversation, GetConversationsByAgent, GetConversation, UpdateConversationTitle, TouchConversation (update updated_at), DeleteConversation.

- [ ] **Step 5: Write messages queries** (`internal/db/messages.go`) — InsertMessage (with nullable tool_call_id/tool_name via helper), GetMessages (ordered ASC, COALESCE for nullable fields).

- [ ] **Step 6: Write settings queries** (`internal/db/settings.go`) — GetSettings, GetSetting, UpsertSetting (INSERT ON CONFLICT).

- [ ] **Step 7: Write tests** (`internal/db/queries_test.go`) — use `:memory:` SQLite, test all CRUD operations: providers, agents with tools, conversations with messages, settings.

- [ ] **Step 8: Run tests** — `go test ./internal/db/ -v`

- [ ] **Step 9: Commit** — `feat: database query layer with CRUD for all entities`

---

### Task 4: Tool System

**Files:**
- Create: `internal/tool/registry.go`
- Create: `internal/tool/web_search.go`
- Create: `internal/tool/web_fetch.go`
- Create: `internal/tool/run_cmd.go`
- Create: `internal/tool/file_read.go`
- Create: `internal/tool/file_write.go`
- Create: `internal/tool/execute.go`
- Create: `internal/tool/registry_test.go`

- [ ] **Step 1: Write tool interface and registry** (`internal/tool/registry.go`) — Tool interface with Name(), Description(), Category(), Parameters() (returns ToolSchema with Type, Properties map of ParamProperty, Required slice), Execute(ctx, args). Registry struct with thread-safe Register/Get/All. DefaultRegistry() that registers all 6 tools.

- [ ] **Step 2: Write web_search tool** (`internal/tool/web_search.go`) — uses Brave Search API via BRAVE_API_KEY env var. Parameters: query (required), count (optional). Returns formatted search results.

- [ ] **Step 3: Write web_fetch tool** (`internal/tool/web_fetch.go`) — HTTP GET with User-Agent header, reads body up to 1MB, returns trimmed string.

- [ ] **Step 4: Write run_cmd tool** (`internal/tool/run_cmd.go`) — executes via `sh -c`, 30s timeout, 1MB output limit, optional workdir.

- [ ] **Step 5: Write file_system_read tool** (`internal/tool/file_read.go`) — os.ReadFile, 1MB output limit.

- [ ] **Step 6: Write file_system_write tool** (`internal/tool/file_write.go`) — os.WriteFile, creates parent dirs with MkdirAll.

- [ ] **Step 7: Write execute tool** (`internal/tool/execute.go`) — exec.CommandContext with args parsing, 30s timeout, 1MB output limit.

- [ ] **Step 8: Write tests** (`internal/tool/registry_test.go`) — test registry operations, test web_fetch with httpbin.org/get, test run_cmd with echo, test file read/write in temp dir.

- [ ] **Step 9: Run tests** — `go test ./internal/tool/ -v -timeout 30s`

- [ ] **Step 10: Commit** — `feat: tool system with registry and 6 built-in tools`

---

### Task 5: LLM Client

**Files:**
- Create: `internal/llm/client.go`

- [ ] **Step 1: Get gorilla/websocket dependency** — `go get github.com/gorilla/websocket`

- [ ] **Step 2: Write LLM client** (`internal/llm/client.go`) — types: Message, ToolCall, FunctionCall, ToolDef, FunctionDef, ChatRequest, ChatResponse, Choice, StreamChunk. Client struct with baseURL, apiKey, model. Chat() method: POST to /chat/completions, builds request with tools as ToolDef array. ChatStream() method: same request with stream=true, reads SSE with bufio.Scanner, sends chunks through channel. Helper `ModelMessagesToLLM()` converts model.Message to llm.Message.

- [ ] **Step 3: Compile** — `go build ./...`

- [ ] **Step 4: Commit** — `feat: OpenAI-compatible LLM client with streaming support`

---

### Task 6: Agent Orchestrator

**Files:**
- Create: `internal/agent/agent.go`

- [ ] **Step 1: Write agent orchestrator** (`internal/agent/agent.go`) — Orchestrator struct with registry. Run() method: loops up to 10 iterations, calls LLM, checks for tool_calls, executes tools via registry, appends results, returns all assistant messages. RunStream() method: uses ChatStream, accumulates content/tool calls from SSE chunks, sends StreamEvent through channel. executeTool() resolves + executes + formats tool results as tool-role messages.

- [ ] **Step 2: Compile** — `go build ./...`

- [ ] **Step 3: Commit** — `feat: agent orchestrator with tool-calling loop and streaming`

---

### Task 7: WebSocket Hub

**Files:**
- Create: `internal/ws/hub.go`

- [ ] **Step 1: Write WebSocket hub** (`internal/ws/hub.go`) — Hub struct with thread-safe client map. HandleConnection upgrades HTTP to WS, subscribes client to conversation_id query param. SendEvent broadcasts JSON events to all subscribers of a conversation. ReadMessages loop for client disconnect detection.

- [ ] **Step 2: Compile** — `go build ./...`

- [ ] **Step 3: Commit** — `feat: WebSocket hub for streaming chat events`

---

### Task 8: API Server — Middleware & Router

**Files:**
- Create: `internal/api/server.go`
- Create: `internal/api/middleware.go`

- [ ] **Step 1: Get chi dependencies** — `go get github.com/go-chi/chi/v5 && go get github.com/go-chi/cors`

- [ ] **Step 2: Write middleware** (`internal/api/middleware.go`) — errorResponse struct, writeJSON(), writeError(), loggingMiddleware (logs method/path/duration).

- [ ] **Step 3: Write server setup** (`internal/api/server.go`) — Server struct with db, router, toolRegistry, orchestrator, wsHub. NewServer() configures chi router with Recoverer, RequestID, logging, CORS. Mounts /api routes: agents, providers, tools, conversations, settings, ws, docs. Serves embedded SPA via fs.Sub + http.FileServer on "/*".

- [ ] **Step 4: Compile only (route implementations in next task)** — `go build ./...`

- [ ] **Step 5: Commit** — `feat: API server skeleton with middleware and routing`

---

### Task 9: API Server — Route Handlers

**Files:**
- Create: `internal/api/providers.go`
- Create: `internal/api/agents.go`
- Create: `internal/api/tools.go`
- Create: `internal/api/conversations.go`
- Create: `internal/api/messages.go`
- Create: `internal/api/settings.go`
- Create: `internal/api/docs.go`

- [ ] **Step 1: Write providers handlers** (`internal/api/providers.go`) — CRUD with JSON input (name, base_url, api_key). API key encrypted before storage. Listing returns providers without encrypted keys. HTTP codes: 200, 201, 404, 204.

- [ ] **Step 2: Write agents handlers** (`internal/api/agents.go`) — CRUD with agentInput struct. Creates/updates agents with tool assignments via the join table. Get returns agent with tools.

- [ ] **Step 3: Write tools handler** (`internal/api/tools.go`) — list all tools, enriched with parameter schemas from the tool registry.

- [ ] **Step 4: Write conversations handlers** (`internal/api/conversations.go`) — list by agent, create, get with messages, delete. Conversation routes nested under both /conversations and /conversations/agents/:agentID.

- [ ] **Step 5: Write messages handler** (`internal/api/messages.go`) — handleSendMessage: loads conversation + agent + provider, saves user message, builds LLM message list (system prompt + history), creates LLM client (decrypting API key), calls orchestrator.Run(), saves results. handleStreamMessage: same setup but SSE streaming with chan. Encryption helpers: encryptAPIKey/decryptAPIKey using AES-GCM with fixed 32-byte key. Also handleGetMessages for listing.

- [ ] **Step 6: Write settings handlers** (`internal/api/settings.go`) — list all, batch update.

- [ ] **Step 7: Write docs page** (`internal/api/docs.go`) — inline HTML with dark theme, lists all endpoints with method badges, example payloads, SSE event format, error codes.

- [ ] **Step 8: Compile** — `go build ./...`

- [ ] **Step 9: Commit** — `feat: REST API server with all endpoints, docs, and SSE streaming`

---

### Task 10: Config & Main Entrypoint

**Files:**
- Modify: `cmd/server/main.go`
- Create: `internal/config/config.go`

- [ ] **Step 1: Write config** (`internal/config/config.go`) — Config struct with Port (default "8080"), DBPath (default "data/agentic.db"). Load() uses getEnv() helper.

- [ ] **Step 2: Write main.go** — `cmd/server/main.go`:
  - `//go:embed web/dist/*` directive to embed SPA
  - Load config, create data dir
  - Open DB, run Migrate
  - Create DefaultRegistry, seed tools into DB (UpsertTool for each)
  - Create API server passing DB, registry, staticFiles
  - http.ListenAndServe on configured port. Log docs URL.

- [ ] **Step 3: Build** — `go build ./cmd/server` (will fail if web/dist doesn't exist yet, but syntax checks pass)

- [ ] **Step 4: Commit** — `feat: main entrypoint with config, DB init, tool seeding`

---

### Task 11: Frontend Scaffolding

**Files:**
- Create: `web/package.json`
- Create: `web/svelte.config.js`
- Create: `web/vite.config.ts`
- Create: `web/src/app.html`
- Create: `web/src/app.css`
- Create: `web/src/routes/+layout.svelte`
- Create: `web/src/routes/chat/+page.svelte` (placeholder)
- Create: `web/src/routes/settings/+page.svelte` (placeholder)

- [ ] **Step 1: Write package.json** — dependencies: @sveltejs/adapter-static, @sveltejs/kit, @sveltejs/vite-plugin-svelte, svelte 5, vite 6. Scripts: dev, build, preview.

- [ ] **Step 2: Write svelte.config.js** — adapter-static to dist/, fallback index.html.

- [ ] **Step 3: Write vite.config.ts** — sveltekit plugin, proxy /api to localhost:8080 for dev.

- [ ] **Step 4: Write app.html** — basic HTML5 with %sveltekit.head% and %sveltekit.body%.

- [ ] **Step 5: Write app.css** — CSS custom properties for dark theme (--bg, --bg-surface, --bg-elevated, --border, --text, --text-muted, --accent, --danger, --success). Base styles for body, a, button, input, textarea, select.

- [ ] **Step 6: Write layout** (`+layout.svelte`) — App shell with sidebar navigation (Agents, Chat, Settings links) and content slot. Sidebar 220px, dark surface background.

- [ ] **Step 7: Write placeholder pages** for chat and settings — empty pages with titles.

- [ ] **Step 8: Build frontend** — `cd web && npm install && npm run build`

- [ ] **Step 9: Commit** — `feat: SvelteKit frontend scaffolding with layout and navigation`

---

### Task 12: Frontend — API Client

**Files:**
- Create: `web/src/lib/api.ts`

- [ ] **Step 1: Write API client** (`web/src/lib/api.ts`) — TypeScript types for all domain models (Provider, Tool, Agent, Conversation, Message, Setting). Generic request() helper with error handling. Functions: getProviders, createProvider, updateProvider, deleteProvider, getAgents, getAgent, createAgent, updateAgent, deleteAgent, getTools, getConversations, createConversation, getConversation, deleteConversation, getMessages, sendMessage, getSettings, updateSettings.

- [ ] **Step 2: Verify build** — `cd web && npm run build`

- [ ] **Step 3: Commit** — `feat: TypeScript API client for all endpoints`

---

### Task 13: Frontend — Agents Page

**Files:**
- Modify: `web/src/routes/+page.svelte`

- [ ] **Step 1: Write Agents page** — Full CRUD UI:
  - Header: "Agents" title + "New Agent" button
  - Form card (toggled): name input, system prompt textarea, provider select, model input, temperature slider (0-2, step 0.1), max tokens number input (100-128000), tool checkboxes from library. Submit button creates or updates.
  - Agent list: cards showing name, model, tool count. Edit and Delete buttons. Delete with confirm().

- [ ] **Step 2: Build & verify** — `cd web && npm run build`

- [ ] **Step 3: Commit** — `feat: agents page with CRUD form and tool assignment`

---

### Task 14: Frontend — Chat Page

**Files:**
- Modify: `web/src/routes/chat/+page.svelte`

- [ ] **Step 1: Write Chat page** — Two-panel layout:
  - Left sidebar: Agent selector dropdown, "New Chat" button, conversation list (clickable items)
  - Main area: Messages panel (scrollable), each message shows role label and content. Role-specific styling (user=accent bg, assistant=elevated bg, tool=muted). Input area at bottom with textarea + Send button, Enter to send.
  - State management: load agents on mount, select agent loads conversations, select conversation loads messages, send message appends user msg, calls API, appends assistant response.

- [ ] **Step 2: Build & verify** — `cd web && npm run build`

- [ ] **Step 3: Commit** — `feat: chat page with conversation management and messaging`

---

### Task 15: Frontend — Settings Page

**Files:**
- Modify: `web/src/routes/settings/+page.svelte`

- [ ] **Step 1: Write Settings page** — Two sections:
  - Providers: list with name/base_url, Add/Edit form (name, base_url, password-masked API key), delete with confirm.
  - Available tools: grid of tool cards showing name, category badge, description. Read-only view from GET /tools.

- [ ] **Step 2: Build & verify** — `cd web && npm run build`

- [ ] **Step 3: Commit** — `feat: settings page with provider CRUD and tool library overview`

---

### Task 16: Integration Build

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Update Makefile** — targets: `build` (builds frontend then go binary), `run` (builds then runs), `dev` (go run), `frontend` (npm install + build), `test` (go test), `clean` (remove bin/, web/dist/, web/node_modules/). DEFAULT_GOAL = build.

- [ ] **Step 2: Full build** — `make build` (builds frontend then compiles Go with embedded SPA)

- [ ] **Step 3: Verify binary** — `ls -lh bin/server` exists, `./bin/server` starts (Ctrl+C to stop)

- [ ] **Step 4: Run tests** — `make test`

- [ ] **Step 5: Commit** — `feat: integration build with Makefile and embedded SPA`

---

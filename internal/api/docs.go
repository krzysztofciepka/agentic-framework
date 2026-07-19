package api

import "net/http"

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(docsHTML))
}

const docsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Agentic Framework — API Docs</title>
<style>
  :root { --bg:#fdfcfc; --surface:#f8f7f7; --border:rgba(15,0,0,0.12); --ink:#201d1d; --mute:#646262; --ash:#9a9898; --accent:#007aff; --green:#30d158; --orange:#ff9f0a; --red:#ff3b30; --purple:#af52de; }
  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family:'Berkeley Mono','IBM Plex Mono','ui-monospace','SFMono-Regular',Menlo,Monaco,Consolas,monospace; background:var(--bg); color:var(--ink); font-size:14px; line-height:1.6; padding:32px; max-width:1000px; margin:0 auto; }
  h1 { font-size:16px; font-weight:700; margin-bottom:4px; }
  .base { color:var(--mute); margin-bottom:32px; font-size:14px; }
  .section { margin-bottom:48px; }
  .section h2 { font-size:14px; font-weight:700; padding-bottom:8px; border-bottom:1px solid var(--border); margin-bottom:12px; }
  .endpoint { border:1px solid var(--border); border-radius:0; padding:16px; margin-bottom:12px; background:var(--bg); }
  .ep-header { display:flex; align-items:center; gap:8px; margin-bottom:8px; }
  .method { font-size:12px; font-weight:700; padding:2px 8px; border-radius:4px; color:#fff; min-width:52px; text-align:center; }
  .method.get { background:var(--accent); }
  .method.post { background:var(--green); }
  .method.put { background:var(--orange); }
  .method.delete { background:var(--red); }
  .ep-path { font-size:14px; font-weight:500; font-family:inherit; }
  .ep-desc { font-size:14px; color:var(--mute); margin-bottom:8px; }
  pre { background:var(--surface); border:1px solid var(--border); padding:12px; border-radius:0; overflow-x:auto; font-family:inherit; font-size:13px; margin:8px 0; color:var(--ink); }
  code { font-family:inherit; font-size:13px; }
  .label { font-size:12px; color:var(--mute); font-weight:500; margin:8px 0 4px; text-transform:uppercase; letter-spacing:0.05em; }
  hr { border:none; border-top:1px solid var(--border); margin:8px 0; }
  .auth-note { background:var(--surface); border:1px solid var(--border); padding:12px; margin-bottom:32px; font-size:14px; }
  .auth-note code { color:var(--ink); font-weight:500; }
</style>
</head>
<body>
<h1>[+] Agentic Framework — API v1</h1>
<p class="base">Base URL: <code>/api</code></p>

<div class="auth-note">
  <strong>Authentication:</strong> All endpoints (except <code>POST /api/login</code>) require either:<br>
  &nbsp;&nbsp;&bull; Cookie: set via <code>POST /api/login</code> (for browser)<br>
  &nbsp;&nbsp;&bull; Header: <code>Authorization: Bearer &lt;APP_PASSWORD&gt;</code> (for API clients / n8n)
</div>

<div class="section"><h2>[x] Providers</h2>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/providers</span></div>
<p class="ep-desc">List all configured LLM providers. API keys are never returned.</p>
<p class="label">Response 200</p>
<pre>[
  { "id": 1, "name": "OpenCode Go", "base_url": "https://opencode.ai/zen/go/v1", "created_at": "2026-07-18T23:03:17Z" }
]</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method post">POST</span><span class="ep-path">/api/providers</span></div>
<p class="ep-desc">Create a new provider. API key is encrypted at rest.</p>
<p class="label">Request body</p>
<pre>{
  "name": "OpenCode Go",
  "base_url": "https://opencode.ai/zen/go/v1",
  "api_key": "sk-..."
}</pre>
<p class="label">Response 201</p>
<pre>{ "id": 1, "name": "OpenCode Go", "base_url": "https://opencode.ai/zen/go/v1", "created_at": "..." }</pre>
<p class="label">Errors</p>
<pre>400 { "error": "...", "code": "MISSING_FIELDS" }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method put">PUT</span><span class="ep-path">/api/providers/{id}</span></div>
<p class="ep-desc">Update provider name, URL, or API key. Omit fields to leave unchanged.</p>
<p class="label">Request body</p>
<pre>{ "name": "New Name", "base_url": "https://new.url/v1", "api_key": "sk-..." }</pre>
<p class="label">Response 200</p>
<pre>{ "status": "ok" }</pre>
<p class="label">Errors</p>
<pre>404 { "error": "...", "code": "NOT_FOUND" }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method delete">DELETE</span><span class="ep-path">/api/providers/{id}</span></div>
<p class="ep-desc">Delete a provider. Fails if agents reference it (FK constraint).</p>
<p class="label">Response 204 (no body)</p>
</div>
</div>

<div class="section"><h2>[x] Agents</h2>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/agents</span></div>
<p class="ep-desc">List all agents with their assigned tools.</p>
<p class="label">Response 200</p>
<pre>[
  {
    "id": 1, "name": "Helper", "system_prompt": "You are helpful.",
    "provider_id": 1, "model": "deepseek-v4-flash",
    "temperature": 0.7, "max_tokens": 4096,
    "tools": [{ "id": 1, "name": "web_search", "description": "...", "category": "web" }],
    "created_at": "...", "updated_at": "..."
  }
]</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method post">POST</span><span class="ep-path">/api/agents</span></div>
<p class="ep-desc">Create a new agent with optional tool assignments.</p>
<p class="label">Request body</p>
<pre>{
  "name": "Helper",
  "system_prompt": "You are a helpful assistant.",
  "provider_id": 1,
  "model": "deepseek-v4-flash",
  "temperature": 0.7,
  "max_tokens": 4096,
  "tool_ids": [1, 3]
}</pre>
<p class="label">Response 201</p>
<pre>{ "id": 1, "name": "Helper", ... }</pre>
<p class="label">Errors</p>
<pre>400 { "error": "...", "code": "MISSING_FIELDS" }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/agents/{id}</span></div>
<p class="ep-desc">Get a single agent with its tool assignments.</p>
<p class="label">Response 200</p>
<pre>{ "id": 1, "name": "Helper", "tools": [...], ... }</pre>
<p class="label">Errors</p>
<pre>404 { "error": "...", "code": "NOT_FOUND" }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method put">PUT</span><span class="ep-path">/api/agents/{id}</span></div>
<p class="ep-desc">Update agent. Omitted fields keep their current value. Tools are fully replaced.</p>
<p class="label">Request body</p>
<pre>{ "name": "New Name", "temperature": 0.5, "tool_ids": [1] }</pre>
<p class="label">Response 200</p>
<pre>{ "id": 1, "name": "New Name", "tools": [...], ... }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method delete">DELETE</span><span class="ep-path">/api/agents/{id}</span></div>
<p class="ep-desc">Delete agent and all associated conversations/messages (cascade).</p>
<p class="label">Response 204 (no body)</p>
</div>
</div>

<div class="section"><h2>[x] Tools</h2>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/tools</span></div>
<p class="ep-desc">List all available tools with their parameter schemas (JSON Schema for function calling).</p>
<p class="label">Response 200</p>
<pre>[
  {
    "id": 1, "name": "web_search", "description": "Search the web using Brave Search API",
    "category": "web",
    "parameters": {
      "type": "object",
      "properties": {
        "query": { "type": "string", "description": "The search query" },
        "count": { "type": "integer", "description": "Number of results (1-20, default 10)" }
      },
      "required": ["query"]
    }
  }
]</pre>
</div>
</div>

<div class="section"><h2>[x] Conversations</h2>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/conversations/agents/{agentID}</span></div>
<p class="ep-desc">List all conversations for an agent, newest first.</p>
<p class="label">Response 200</p>
<pre>[{ "id":1, "agent_id":1, "title":"Hello", "created_at":"...", "updated_at":"..." }]</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method post">POST</span><span class="ep-path">/api/conversations/agents/{agentID}</span></div>
<p class="ep-desc">Create a new conversation for an agent.</p>
<p class="label">Request body</p>
<pre>{ "title": "Optional custom title" }</pre>
<p class="label">Response 201</p>
<pre>{ "id": 1, "agent_id": 1, "title": "Optional custom title", ... }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/conversations/{id}</span></div>
<p class="ep-desc">Get conversation with all messages.</p>
<p class="label">Response 200</p>
<pre>{
  "id": 1, "agent_id": 1, "title": "Hello",
  "created_at": "...", "updated_at": "...",
  "messages": [
    { "id": 1, "role": "user", "content": "Hello", "created_at": "..." },
    { "id": 2, "role": "assistant", "content": "Hi!", "created_at": "..." }
  ]
}</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method delete">DELETE</span><span class="ep-path">/api/conversations/{id}</span></div>
<p class="ep-desc">Delete conversation and all its messages (cascade).</p>
<p class="label">Response 204 (no body)</p>
</div>
</div>

<div class="section"><h2>[x] Messages</h2>

<div class="endpoint">
<div class="ep-header"><span class="method post">POST</span><span class="ep-path">/api/conversations/{id}/messages</span></div>
<p class="ep-desc">Send a message and get the assistant response. This is a synchronous call — the LLM runs (including tool calls) and returns the final answer. All messages are persisted to the database.</p>
<p class="label">Request body</p>
<pre>{ "role": "user", "content": "Search for latest Go release" }</pre>
<p class="label">Response 200</p>
<pre>{ "role": "assistant", "content": "The latest Go release is 1.26.4..." }</pre>
<p class="label">Errors</p>
<pre>400 { "error": "...", "code": "MISSING_FIELDS" }
400 { "error": "...", "code": "INVALID_ROLE" }
502 { "error": "...", "code": "LLM_ERROR" }</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method post">POST</span><span class="ep-path">/api/conversations/{id}/stream</span></div>
<p class="ep-desc">Send a message and stream the response via Server-Sent Events (SSE). Real-time output for content and tool calls.</p>
<p class="label">Request body</p>
<pre>{ "role": "user", "content": "Write a blog post about Go" }</pre>
<p class="label">SSE events</p>
<pre>data: {"type":"content","content":"# "}
data: {"type":"content","content":"Go in "}
data: {"type":"content","content":"2026\n\n"}
data: {"type":"tool_start","tool":"web_search"}
data: {"type":"tool_end","tool":"web_search","content":"1. Result..."}
data: {"type":"done"}</pre>
<p class="label">Event types</p>
<pre>content     — text chunk from LLM (streamed in real-time)
tool_start  — LLM requested a tool call (tool name included)
tool_end    — tool execution finished (result content included)
done        — streaming complete
error       — error occurred (message in content)</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/conversations/{id}/messages</span></div>
<p class="ep-desc">Get all messages for a conversation, ordered by creation time.</p>
<p class="label">Response 200</p>
<pre>[
  { "id": 1, "conversation_id": 1, "role": "user", "content": "Hello", "created_at": "..." },
  { "id": 2, "conversation_id": 1, "role": "assistant", "content": "Hi!", "created_at": "..." },
  { "id": 3, "conversation_id": 1, "role": "tool", "content": "...", "tool_call_id": "...", "tool_name": "web_search", "created_at": "..." }
]</pre>
</div>
</div>

<div class="section"><h2>[x] Settings</h2>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/settings</span></div>
<p class="ep-desc">Get all application settings as key-value pairs.</p>
<p class="label">Response 200</p>
<pre>[
  { "key": "openode_go_models", "value": "deepseek-v4-flash,deepseek-v4-pro,..." }
]</pre>
</div>

<div class="endpoint">
<div class="ep-header"><span class="method put">PUT</span><span class="ep-path">/api/settings</span></div>
<p class="ep-desc">Upsert settings. Send a flat object with key-value pairs.</p>
<p class="label">Request body</p>
<pre>{
  "openode_go_models": "deepseek-v4-flash,deepseek-v4-pro,kimi-k3",
  "default_model": "deepseek-v4-flash"
}</pre>
<p class="label">Response 200</p>
<pre>{ "status": "ok" }</pre>
</div>
</div>

<div class="section"><h2>[x] WebSocket</h2>

<div class="endpoint">
<div class="ep-header"><span class="method get">GET</span><span class="ep-path">/api/ws?conversation_id={id}</span></div>
<p class="ep-desc">WebSocket connection for receiving real-time stream events from a conversation. Connect before calling POST /stream to receive events in the browser.</p>
<p class="label">Query params</p>
<pre>conversation_id (required) — subscribe to events for this conversation</pre>
<p class="label">Events (JSON over WebSocket)</p>
<pre>{"type":"content","content":"Hello"}
{"type":"tool_start","tool":"web_search"}
{"type":"tool_end","tool":"web_search","content":"..."}
{"type":"done"}</pre>
</div>
</div>

<div class="section"><h2>[x] Auth</h2>

<div class="endpoint">
<div class="ep-header"><span class="method post">POST</span><span class="ep-path">/api/login</span></div>
<p class="ep-desc">Authenticate via password. Sets a session cookie (30 days) for browser access.</p>
<p class="label">Request body</p>
<pre>{ "password": "your-password" }</pre>
<p class="label">Response 200</p>
<pre>{ "status": "ok" }</pre>
<p class="label">Errors</p>
<pre>401 { "error": "...", "code": "INVALID_PASSWORD" }</pre>
</div>
</div>

<div class="section"><h2>[x] Common Errors</h2>
<pre>400 INVALID_JSON     — malformed request body
400 MISSING_FIELDS   — required fields missing
400 INVALID_ROLE     — only "user" role accepted for messages
404 NOT_FOUND        — requested resource doesn't exist
401 UNAUTHORIZED     — missing or invalid auth
502 LLM_ERROR        — upstream LLM API returned an error
500 DB_ERROR         — database operation failed</pre>
</div>

</body>
</html>`

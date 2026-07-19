# Agentic Framework v0.1.1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Five improvements: public docs, correct defaults, shareable chat URLs, image uploads with vision support, docs link in sidebar.

**Architecture:** All changes are isolated modifications to existing files. Image uploads add a new upload handler + file server, and the LLM client gains vision message format support.

**Tech Stack:** Go, Svelte 5, SQLite, OpenAI-compatible API (with vision)

---

### Task 1: Public `/api/docs` + Sidebar Docs Link

**Files:**
- Modify: `internal/api/auth.go` (exempt /api/docs)
- Modify: `web/src/routes/+layout.svelte` (add docs nav item)

- [ ] **Step 1: Exempt /api/docs from auth**

In `internal/api/auth.go`, the auth middleware has an exemption for `/api/login`. Add `/api/docs` to the same condition:

```go
if r.URL.Path == "/api/login" || r.URL.Path == "/api/docs" || isAuthenticated(r, password) {
```

- [ ] **Step 2: Add docs link to sidebar**

In `web/src/routes/+layout.svelte`, add a nav item after Settings:

```svelte
<a href="/api/docs" target="_blank" class="nav-item">[+] Docs</a>
```

- [ ] **Step 3: Build frontend and Go**

Run: `cd web && npm run build && cd .. && go build ./cmd/server`

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: public docs + sidebar docs link"
```

---

### Task 2: Default Model

**Files:**
- Modify: `web/src/routes/+page.svelte` (form defaults)

- [ ] **Step 1: Change default model**

In `+page.svelte`, change initial `formModel` state from `'gpt-4o'` to `'deepseek-v4-pro'`. Also change the `resetForm` function and the placeholder on the input.

- [ ] **Step 2: Build frontend**

Run: `cd web && npm run build`

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "fix: default model deepseek-v4-pro"
```

---

### Task 3: Chat URL with Conversation ID

**Files:**
- Modify: `web/src/routes/chat/+page.svelte` (URL reading + pushState)

- [ ] **Step 1: Read conv from URL on mount**

In the `onMount` callback (or after loadAgents), read `?conv=` from `window.location.search`:

```typescript
onMount(async () => {
  await loadAgents();
  const params = new URLSearchParams(window.location.search);
  const convId = params.get('conv');
  if (convId) {
    const agentId = params.get('agent');
    if (agentId) {
      const agent = agents.find(a => a.id === +agentId);
      if (agent) await selectAgent(agent);
    }
    // try to load conversation after agent is selected
    // or just try loading it directly
  }
});
```

Actually, the simpler approach: after loading agents, check URL for `conv` param, load that conversation directly via `getConversation()`. The agent is found from the conversation's `agent_id`.

In `selectConversation()`, add URL update:

```typescript
async function selectConversation(conv: Conversation) {
  // ... existing load code ...
  const url = new URL(window.location.href);
  url.searchParams.set('conv', String(conv.id));
  url.searchParams.set('agent', String(conv.agent_id));
  window.history.pushState({}, '', url.toString());
}
```

On mount, restore from URL:

```typescript
const urlParams = new URLSearchParams(window.location.search);
const convId = urlParams.get('conv');
if (convId) {
  try {
    const full = await getConversation(+convId);
    if (full) {
      selectedAgent = agents.find(a => a.id === full.agent_id) ?? null;
      if (selectedAgent) conversations = await getConversations(selectedAgent.id);
      selectedConv = full;
      messages = full.messages ?? [];
    }
  } catch (_) {}
}
```

- [ ] **Step 2: Build frontend**

Run: `cd web && npm run build`

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: chat URL with conversation ID for refresh/share"
```

---

### Task 4: Image Uploads

**Files:**
- Create: `internal/api/upload.go` (upload handler)
- Modify: `internal/api/server.go` (mount upload routes)
- Modify: `internal/api/messages.go` (vision format for images)
- Modify: `internal/llm/client.go` (support array content for vision)
- Modify: `web/src/routes/chat/+page.svelte` (image picker + previews)
- Modify: `web/src/lib/api.ts` (upload function)

- [ ] **Step 1: Write upload handler**

Create `internal/api/upload.go`:

```go
package api

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

var imageExts = map[string]string{
    "image/png":  ".png",
    "image/jpeg": ".jpg",
    "image/gif":  ".gif",
    "image/webp": ".webp",
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        writeError(w, http.StatusBadRequest, "TOO_LARGE", "max 10MB")
        return
    }

    file, header, err := r.FormFile("image")
    if err != nil {
        writeError(w, http.StatusBadRequest, "MISSING_FILE", "image field required")
        return
    }
    defer file.Close()

    contentType := header.Header.Get("Content-Type")
    ext, ok := imageExts[contentType]
    if !ok {
        writeError(w, http.StatusBadRequest, "INVALID_TYPE", "only PNG, JPEG, GIF, WebP")
        return
    }

    id := make([]byte, 16)
    rand.Read(id)
    filename := hex.EncodeToString(id) + ext

    uploadDir := os.Getenv("UPLOAD_DIR")
    if uploadDir == "" {
        uploadDir = "data/uploads"
    }
    os.MkdirAll(uploadDir, 0755)

    dst, err := os.Create(filepath.Join(uploadDir, filename))
    if err != nil {
        writeError(w, http.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        writeError(w, http.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
        return
    }

    writeJSON(w, http.StatusCreated, map[string]string{
        "id":  filename,
        "url": "/uploads/" + filename,
    })
}
```

- [ ] **Step 2: Add upload route and file server to server.go**

Add in the `/api` route group before auth middleware applies:
- `r.Post("/upload", s.handleUpload)` inside the `/api` Route block

Add in `NewServer()`, serve uploads directory from host filesystem:

```go
uploadDir := os.Getenv("UPLOAD_DIR")
if uploadDir == "" {
    uploadDir = "data/uploads"
}
s.router.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
```

Note: this should be added BEFORE the spaFileServer `/*` catch-all route.

- [ ] **Step 3: Support vision format in LLM client**

In `internal/llm/client.go`, the `ChatRequest.Messages` currently uses `Message` with `Content string`. For vision, `Content` can be either a string (text) or an array of content blocks.

Change the `Message` type to support both:

```go
type Message struct {
    Role       string      `json:"role"`
    Content    any         `json:"content,omitempty"`
    ToolCallID string      `json:"tool_call_id,omitempty"`
    ToolCalls  []*ToolCall `json:"tool_calls,omitempty"`
}

type ContentBlock struct {
    Type     string          `json:"type"`
    Text     string          `json:"text,omitempty"`
    ImageURL *ImageURLBlock  `json:"image_url,omitempty"`
}

type ImageURLBlock struct {
    URL    string `json:"url"`
    Detail string `json:"detail,omitempty"`
}
```

When building messages, if the model.Message has images (checking a `Images` field or parsing content), build content as `[]ContentBlock` instead of plain string.

Simpler approach: add an `Images []string` field to the API message input, and in `ModelMessagesToLLM` helper, convert to vision format:

```go
// In messages.go handler, when building llm messages:
if len(req.Images) > 0 {
    blocks := []llm.ContentBlock{{Type: "text", Text: req.Content}}
    for _, imgURL := range req.Images {
        absURL := "https://" + r.Host + imgURL
        blocks = append(blocks, llm.ContentBlock{
            Type: "image_url",
            ImageURL: &llm.ImageURLBlock{URL: absURL, Detail: "auto"},
        })
    }
    llmMsg.Content = blocks
} else {
    llmMsg.Content = req.Content
}
```

- [ ] **Step 4: Update message handler to accept images**

In `internal/api/messages.go`, `handleSendMessage`, update the input struct:

```go
var req struct {
    Role    string   `json:"role"`
    Content string   `json:"content"`
    Images  []string `json:"images"`
}
```

When building the user LLM message, if `req.Images` is non-empty, build vision format content blocks with the image URLs (prepended with `https://` + Host header).

- [ ] **Step 5: Update API client in frontend**

In `web/src/lib/api.ts`, add:

```typescript
export const uploadImage = async (file: File): Promise<{id: string, url: string}> => {
  const form = new FormData();
  form.append('image', file);
  const res = await fetch(`${BASE}/upload`, { method: 'POST', body: form });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || 'Upload failed');
  }
  return res.json();
};
```

Update `sendMessage` to accept optional `images` array.

- [ ] **Step 6: Add image picker to chat page**

In `web/src/routes/chat/+page.svelte`:
- Add state: `let pendingImages = $state<{id:string, url:string}[]>([])`
- Add upload button (📎 or `[+]`) next to the textarea
- On file select: upload via `uploadImage()`, add to `pendingImages`
- Show pending images as small thumbnails above the input
- When sending message, pass image URLs to `sendMessage` and also include them in the optimistic user message content
- Clear pending images after send
- In message bubbles, render images inline (if content contains image URLs or if message has an images field)

- [ ] **Step 7: Build, test, deploy**

Run: `cd web && npm run build && cd .. && go build -o bin/server ./cmd/server`

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "feat: image uploads with vision model support"
```

---

### Task 5: Integration Build & Release

**Files:**
- None new

- [ ] **Step 1: Full build**

```bash
make build
```

- [ ] **Step 2: Run tests**

```bash
go test ./...
```

- [ ] **Step 3: Deploy**

```bash
git push origin master && ssh server "/opt/apps/redeploy-agentic-framework.sh"
```

- [ ] **Step 4: Create release**

```bash
gh release create v0.1.1 --title "agentic-framework v0.1.1" --generate-notes
```

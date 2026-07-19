# Agentic Framework v0.1.1 — Design Spec

## Overview

Five improvements to the agentic-framework: public docs, correct defaults, shareable chat URLs, image uploads, and docs link in navigation.

## 1. Public API Docs (`/api/docs`)

Exempt `/api/docs` from authentication middleware. Add `r.URL.Path == "/api/docs"` to the exempt paths in `auth.go`.

## 2. Default Model

Change the frontend default model from `"gpt-4o"` to `"deepseek-v4-pro"`. The datalist from settings already provides all OpenCode Go models.

## 3. Chat URL with Conversation ID

Use query parameter `?conv=<id>` on the `/chat` route:
- On page load, read `conv` from `URLSearchParams`, load that conversation if it exists
- When the user selects a conversation, update the URL via `history.pushState` (no page reload)
- On refresh, the same conversation is loaded from the URL parameter
- No `conv` param → empty state (select a conversation prompt)

This keeps the SPA routing simple — no new SvelteKit routes needed.

## 4. Image Uploads

### API
- **Upload:** `POST /api/upload` — multipart form with `image` field (PNG, JPEG, GIF, WebP up to 10MB). Returns `{id, url}`.
- **Serve:** `GET /uploads/<filename>` — serves uploaded images from local filesystem
- **Message format:** user messages can include `images: [{url: "/uploads/file.png", detail: "auto"}]`

### Storage
- Path: `/app/data/uploads/<uuid>.<ext>`
- Exposed via Go file server at `/uploads/`
- Container volume `agentic-framework-data` already mounted at `/app/data`

### LLM Vision Support
When a message has images, convert `content` from plain string to OpenAI vision format:
```json
[
  {"type": "text", "text": "user message"},
  {"type": "image_url", "image_url": {"url": "https://host/uploads/file.png", "detail": "auto"}}
]
```
The image URL must be absolute (with host). Use the request's `Host` header to construct the full URL.

Images in assistant responses are rendered inline if the LLM returns image content (multimodal models).

### Frontend
- Image upload button (paperclip icon) next to the chat textarea
- Click → file picker (accept images)
- Upload the file first, get URL, then include in message payload
- Show uploaded image previews above the input before sending
- Show images in message bubbles (both user and assistant)

## 5. Docs Link in Sidebar

Add `[+] Docs` nav item in the sidebar layout, linking to `/api/docs` with `target="_blank"`.

---

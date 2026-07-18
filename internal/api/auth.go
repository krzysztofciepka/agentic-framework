package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

const authCookieName = "agentic_auth"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := os.Getenv("APP_PASSWORD")
		if password == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/api/login" || isAuthenticated(r, password) {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing auth")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(loginHTML))
	})
}

func isAuthenticated(r *http.Request, password string) bool {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ") == password
	}

	cookie, err := r.Cookie(authCookieName)
	if err == nil && cookie.Value == password {
		return true
	}

	return false
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	password := os.Getenv("APP_PASSWORD")
	if password == "" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(loginHTML))
		return
	}

	var input struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if input.Password != password {
		writeError(w, http.StatusUnauthorized, "INVALID_PASSWORD", "wrong password")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    password,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 30,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Agentic Framework</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: 'Berkeley Mono', 'IBM Plex Mono', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', monospace;
  background: #fdfcfc; color: #201d1d;
  display: flex; align-items: center; justify-content: center;
  height: 100vh; font-size: 16px; line-height: 1.5;
}
.card {
  border: 1px solid rgba(15,0,0,0.12); padding: 24px; width: 360px;
  display: flex; flex-direction: column; gap: 16px;
}
h1 { font-size: 16px; font-weight: 700; }
label { font-size: 14px; color: #646262; line-height: 2; }
input {
  background: #f8f7f7; border: 1px solid rgba(15,0,0,0.12);
  border-radius: 4px; padding: 8px 12px; font-family: inherit;
  font-size: 16px; color: #201d1d; outline: none; width: 100%;
}
input:focus { background: #fdfcfc; border-color: #201d1d; }
button {
  cursor: pointer; border: none; border-radius: 4px;
  padding: 4px 20px; height: 36px; background: #201d1d;
  color: #fdfcfc; font-family: inherit; font-size: 16px; font-weight: 500; line-height: 2;
}
button:active { background: #0f0000; }
.error { color: #ff3b30; font-size: 14px; display: none; }
</style>
</head>
<body>
<div class="card">
  <h1>[agentic]</h1>
  <form id="login">
    <label for="pw">Password</label>
    <input id="pw" type="password" placeholder="Enter password" autofocus />
    <div class="error" id="err">Wrong password</div>
    <button type="submit" style="margin-top:12px;width:100%">[x] Login</button>
  </form>
</div>
<script>
document.getElementById('login').addEventListener('submit', async (e) => {
  e.preventDefault();
  const pw = document.getElementById('pw').value;
  const res = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password: pw })
  });
  if (res.ok) location.reload();
  else document.getElementById('err').style.display = 'block';
});
</script>
</body>
</html>`

package handler

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles login and logout requests.
type AuthHandler struct {
	login    string
	password string // bcrypt hash
	sessions SessionManager
}

type SessionManager interface {
	CreateCookie(login string) (*http.Cookie, error)
	DeleteCookie() *http.Cookie
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(login, password string, sessions SessionManager) *AuthHandler {
	return &AuthHandler{
		login:    login,
		password: password,
		sessions: sessions,
	}
}

// LoginPage renders the login form (GET /login).
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Simple inline login form — will be replaced with templ later
	w.Write([]byte(loginPageHTML))
}

// Login handles login form submission (POST /login).
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	if login != h.login {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(h.password), []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	cookie, err := h.sessions.CreateCookie(login)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout handles logout (POST /logout).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.sessions.DeleteCookie())
	http.Redirect(w, r, "/login", http.StatusFound)
}

const loginPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login — a3b-webui</title>
    <script src="https://unpkg.com/htmx.org@2"></script>
    <style>
        body { font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #f5f5f5; }
        .card { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); width: 100%; max-width: 360px; }
        h1 { margin-top: 0; font-size: 1.5rem; }
        label { display: block; margin-bottom: 0.25rem; font-weight: 500; }
        input { width: 100%; padding: 0.5rem; margin-bottom: 1rem; border: 1px solid #ccc; border-radius: 4px; box-sizing: border-box; }
        button { width: 100%; padding: 0.5rem; background: #2563eb; color: white; border: none; border-radius: 4px; font-size: 1rem; cursor: pointer; }
        button:hover { background: #1d4ed8; }
        .error { color: #dc2626; margin-bottom: 1rem; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Login</h1>
        <form hx-post="/login" hx-target="body" hx-push-url="true">
            <label for="login">Login</label>
            <input type="text" id="login" name="login" required autocomplete="username">
            <label for="password">Password</label>
            <input type="password" id="password" name="password" required autocomplete="current-password">
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`

package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/a3bot6/a3b-webui/templates"
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
	component := templates.LoginPage("")
	component.Render(r.Context(), w)
}

// Login handles login form submission (POST /login).
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		component := templates.LoginCard("Bad request")
		component.Render(r.Context(), w)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	if login != h.login {
		component := templates.LoginCard("Invalid credentials")
		component.Render(r.Context(), w)
		return
	}

	if !h.checkPassword(password) {
		component := templates.LoginCard("Invalid credentials")
		component.Render(r.Context(), w)
		return
	}

	cookie, err := h.sessions.CreateCookie(login)
	if err != nil {
		component := templates.LoginCard("Internal error")
		component.Render(r.Context(), w)
		return
	}

	http.SetCookie(w, cookie)
	// HTMX: use HX-Redirect header for AJAX navigation
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

// checkPassword supports both bcrypt hash and plaintext password.
func (h *AuthHandler) checkPassword(password string) bool {
	if strings.HasPrefix(h.password, "$2") {
		// bcrypt hash
		return bcrypt.CompareHashAndPassword([]byte(h.password), []byte(password)) == nil
	}
	// plaintext comparison
	return subtle.ConstantTimeCompare([]byte(password), []byte(h.password)) == 1
}

// Logout handles logout (POST /logout).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.sessions.DeleteCookie())
	http.Redirect(w, r, "/login", http.StatusFound)
}

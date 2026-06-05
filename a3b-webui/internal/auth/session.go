package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SessionManager handles HMAC-signed session cookies.
type SessionManager struct {
	secret []byte
}

// NewSessionManager creates a new session manager with the given secret.
func NewSessionManager(secret string) *SessionManager {
	return &SessionManager{secret: []byte(secret)}
}

const (
	cookieName  = "session"
	cookieTTL   = 24 * time.Hour
	saltAndPepper = "a3b-webui-session"
)

// CreateCookie creates a signed session cookie for the given login.
func (sm *SessionManager) CreateCookie(login string) (*http.Cookie, error) {
	expires := time.Now().Add(cookieTTL)
	signed, err := sm.sign(login, expires)
	if err != nil {
		return nil, fmt.Errorf("sign session: %w", err)
	}
	return &http.Cookie{
		Name:     cookieName,
		Value:    signed,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// ValidateCookie validates the session cookie from the request and returns the login.
func (sm *SessionManager) ValidateCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("no session cookie")
	}
	login, err := sm.verify(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("invalid session: %w", err)
	}
	return login, nil
}

// DeleteCookie creates a cookie that expires immediately (for logout).
func (sm *SessionManager) DeleteCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// sign creates an HMAC-signed token: hex(expires_unix).login.hex(signature)
func (sm *SessionManager) sign(login string, expires time.Time) (string, error) {
	expStr := fmt.Sprintf("%d", expires.Unix())
	msg := fmt.Sprintf("%s.%s.%s", expStr, login, saltAndPepper)
	mac := hmac.New(sha256.New, sm.secret)
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s.%s", expStr, login, sig), nil
}

// verify checks the HMAC signature and expiry, returns login on success.
func (sm *SessionManager) verify(token string) (string, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed token")
	}
	expStr, login, sig := parts[0], parts[1], parts[2]

	// Check expiry
	expUnix, err := parseExpiry(expStr)
	if err != nil {
		return "", fmt.Errorf("invalid expiry: %w", err)
	}
	if time.Now().After(expUnix) {
		return "", fmt.Errorf("session expired")
	}

	// Verify signature
	msg := fmt.Sprintf("%s.%s.%s", expStr, login, saltAndPepper)
	mac := hmac.New(sha256.New, sm.secret)
	mac.Write([]byte(msg))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid signature")
	}

	return login, nil
}

func parseExpiry(s string) (time.Time, error) {
	var sec int64
	if _, err := fmt.Sscanf(s, "%d", &sec); err != nil {
		return time.Time{}, fmt.Errorf("cannot parse expiry: %s", s)
	}
	return time.Unix(sec, 0), nil
}

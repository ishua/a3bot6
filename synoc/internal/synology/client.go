package synology

import (
	"net/http"
	"time"
)

// Client для работы с Synology Download Station API
type Client struct {
	baseURL string
	sid     string
	http    *http.Client
}

// NewClient создает новый клиент для Synology API
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsAuthenticated проверяет, авторизован ли клиент
func (c *Client) IsAuthenticated() bool {
	return c.sid != ""
}

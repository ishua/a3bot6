package synology

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Login выполняет авторизацию в Synology
func (c *Client) Login(username, password string) error {
	apiURL := fmt.Sprintf("%s/webapi/auth.cgi", c.baseURL)
	params := url.Values{
		"api":     {"SYNO.API.Auth"},
		"version": {"7"},
		"method":  {"login"},
		"account": {username},
		"passwd":  {password},
		"session": {"DownloadStation"},
		"format":  {"sid"},
	}

	resp, err := c.http.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	var result synoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("login decode failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("login failed, error code: %d", result.Error.Code)
	}

	var data struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return fmt.Errorf("login parse sid failed: %w", err)
	}

	c.sid = data.SID
	return nil
}

// Logout выполняет выход из системы
func (c *Client) Logout() error {
	if !c.IsAuthenticated() {
		return nil
	}

	apiURL := fmt.Sprintf("%s/webapi/auth.cgi", c.baseURL)
	params := url.Values{
		"api":     {"SYNO.API.Auth"},
		"version": {"7"},
		"method":  {"logout"},
		"session": {"DownloadStation"},
		"_sid":    {c.sid},
	}

	resp, err := c.http.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer resp.Body.Close()

	var result synoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("logout decode failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("logout failed")
	}

	c.sid = ""
	return nil
}

// synoResponse базовая структура ответа от Synology API
type synoResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *synoError      `json:"error,omitempty"`
}

type synoError struct {
	Code int `json:"code"`
}

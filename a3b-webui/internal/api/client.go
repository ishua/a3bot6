package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is an HTTP client for xray-manual-svc.
type Client struct {
	baseURL   string
	secret    string
	httpClient *http.Client
}

// response is the generic API response wrapper.
type response struct {
	Data  json.RawMessage `json:"data"`
	Error *string         `json:"error"`
}

// New creates a new API client.
func New(baseURL, secret string) *Client {
	return &Client{
		baseURL:    baseURL,
		secret:     secret,
		httpClient: &http.Client{},
	}
}

// doGet performs a GET request and returns the raw data.
func (c *Client) doGet(path string) (json.RawMessage, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.do(req)
}

// doPost performs a POST request with optional body and returns the raw data.
func (c *Client) doPost(path string, body any) (json.RawMessage, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// do executes an HTTP request, validates the response, and returns data.
func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	req.Header.Set("secret", c.secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if r.Error != nil && *r.Error != "" {
		return nil, fmt.Errorf("api error: %s", *r.Error)
	}

	return r.Data, nil
}

// StatusData represents the VPN status response.
type StatusData struct {
	Override        string `json:"override"`
	PrincipleTarget string `json:"principle_target"`
}

// GetList returns the list of available VPN tags.
func (c *Client) GetList() ([]string, error) {
	data, err := c.doGet("/list")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}
	return resp.Tags, nil
}

// GetStatus returns the current VPN status.
func (c *Client) GetStatus() (*StatusData, error) {
	data, err := c.doGet("/status")
	if err != nil {
		return nil, err
	}
	var s StatusData
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}
	return &s, nil
}

// Use switches VPN to the given tag.
func (c *Client) Use(tag string) error {
	_, err := c.doPost("/use", map[string]string{"tag": tag})
	return err
}

// Auto enables automatic VPN switching.
func (c *Client) Auto() error {
	_, err := c.doPost("/auto", nil)
	return err
}

// PingData represents the ping result.
type PingData struct {
	IP        string `json:"ip"`
	LatencyMs int64  `json:"latency_ms"`
}

// Ping pings the current VPN endpoint.
func (c *Client) Ping() (*PingData, error) {
	data, err := c.doGet("/ping")
	if err != nil {
		return nil, err
	}
	var p PingData
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("decode ping: %w", err)
	}
	return &p, nil
}

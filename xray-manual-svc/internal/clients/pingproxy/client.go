package pingproxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const pingURL = "https://ifconfig.me/ip"
const timeOutSec = 2

type PingResult struct {
	IP      string
	Latency time.Duration
}

type Client struct {
	http *http.Client
}

func New(proxyAddr string) (*Client, error) {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address: %w", err)
	}

	return &Client{
		http: &http.Client{
			Timeout: timeOutSec * time.Second,
			Transport: &http.Transport{
				Proxy:             http.ProxyURL(proxyURL),
				DisableKeepAlives: true,
			},
		},
	}, nil
}

func (c *Client) Ping(ctx context.Context) (*PingResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pingURL, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	latency := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &PingResult{
		IP:      strings.TrimSpace(string(body)),
		Latency: latency,
	}, nil
}

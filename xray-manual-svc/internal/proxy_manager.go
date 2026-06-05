package internal

import (
	"context"
	"xray-manual-svc/internal/clients/pingproxy"
	"xray-manual-svc/internal/clients/xray"
    "fmt"
)

type XrayClient interface {
	GetStatus(ctx context.Context) (*xray.BalancerStatus, error)
	SetTarget(ctx context.Context, tag string) error
	ResetTarget(ctx context.Context) error
}

type PingClient interface {
	Ping(ctx context.Context) (*pingproxy.PingResult, error)
}

type ProxyManager struct {
	tags []string
	xray XrayClient
	ping PingClient
}

func NewProxyManager(tags []string, xray XrayClient, ping PingClient) *ProxyManager {
	return &ProxyManager{
		tags: tags,
		xray: xray,
		ping: ping,
	}
}

func (m *ProxyManager) List() []string {
	return m.tags
}

func (m *ProxyManager) Status() (*xray.BalancerStatus, error) {
	return m.xray.GetStatus(context.Background())
}

func (m *ProxyManager) Use(tag string) error {
	if !m.validateTag(tag) {
		return fmt.Errorf("unknown tag: %s", tag)
	}
	return m.xray.SetTarget(context.Background(), tag)
}

func (m *ProxyManager) validateTag(tag string) bool {
	for _, t := range m.tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (m *ProxyManager) Auto() error {
	return m.xray.ResetTarget(context.Background())
}

func (m *ProxyManager) Ping() (*pingproxy.PingResult, error) {
	return m.ping.Ping(context.Background())
}

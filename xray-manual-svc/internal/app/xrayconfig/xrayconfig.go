package xrayconfig

import (
    "encoding/json"
    "os"
    "strings"
)

var serviceOutbounds = map[string]struct{}{
    "api":    {},
    "direct": {},
    "block":  {},
}

type Config struct {
    Routing   Routing    `json:"routing"`
    Outbounds []Outbound `json:"outbounds"`
}

type Routing struct {
    Balancers []Balancer `json:"balancers"`
}

type Balancer struct {
    Tag      string   `json:"tag"`
    Selector []string `json:"selector"`
}

type Outbound struct {
    Tag string `json:"tag"`
}

func Load(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var cfg Config
    if err := json.NewDecoder(f).Decode(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func (c *Config) BalancerTags(balancerTag string) []string {
    var selectors []string
    for _, b := range c.Routing.Balancers {
        if b.Tag == balancerTag {
            selectors = b.Selector
            break
        }
    }

    var tags []string
    for _, o := range c.Outbounds {
        if _, svc := serviceOutbounds[o.Tag]; svc {
            continue
        }
        for _, prefix := range selectors {
            if strings.HasPrefix(o.Tag, prefix) {
                tags = append(tags, o.Tag)
                break
            }
        }
    }

    return tags
}

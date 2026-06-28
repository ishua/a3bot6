package xrayconfig

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "slices"
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

type partialConfig struct {
    Outbounds []Outbound `json:"outbounds"`
}

func Load(dirPath string) (*Config, error) {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return nil, fmt.Errorf("read dir %s: %w", dirPath, err)
    }

    var jsonFiles []string
    for _, e := range entries {
        if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
            continue
        }
        jsonFiles = append(jsonFiles, e.Name())
    }

    slices.Sort(jsonFiles)

    if len(jsonFiles) == 0 {
        return nil, fmt.Errorf("no json files in xray config dir: %s", dirPath)
    }

    var cfg Config
    first := true

    for _, name := range jsonFiles {
        path := filepath.Join(dirPath, name)
        f, err := os.Open(path)
        if err != nil {
            return nil, fmt.Errorf("open %s: %w", path, err)
        }

        if first {
            if err := json.NewDecoder(f).Decode(&cfg); err != nil {
                f.Close()
                return nil, fmt.Errorf("decode %s: %w", path, err)
            }
            first = false
        } else {
            var pc partialConfig
            if err := json.NewDecoder(f).Decode(&pc); err != nil {
                f.Close()
                return nil, fmt.Errorf("decode %s: %w", path, err)
            }
            cfg.Outbounds = append(cfg.Outbounds, pc.Outbounds...)
        }
        f.Close()
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

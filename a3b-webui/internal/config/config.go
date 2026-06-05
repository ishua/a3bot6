package config

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	API    APIConfig    `yaml:"api"`
	Auth   AuthConfig   `yaml:"auth"`
}

type ServerConfig struct {
	Addr string `yaml:"addr" default:":8090" usage:"HTTP server address"`
}

type APIConfig struct {
	URL    string `yaml:"url" required:"true" usage:"xray-manual-svc base URL"`
	Secret string `yaml:"secret" required:"true" usage:"API secret header"`
}

type AuthConfig struct {
	Login         string `yaml:"login" required:"true" usage:"Admin login"`
	Password      string `yaml:"password" required:"true" usage:"Admin password (bcrypt hash)"`
	SessionSecret string `yaml:"session_secret" required:"true" usage:"HMAC secret for session cookie"`
}

func Load() (*Config, error) {
	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"conf/a3b_webui_config.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})
	if err := loader.Load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &cfg, nil
}

package config

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

const configPath = "conf/xray_manual_svc_config.yaml"

func Load() (*Config, error) {
	var cfg Config

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{configPath},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	if err := loader.Load(); err != nil {
		return nil, err
	}

	if cfg.Xray.ConfigDir == "" {
		return nil, fmt.Errorf("cfg.Xray.ConfigDir is nil")
	}

	return &cfg, nil
}

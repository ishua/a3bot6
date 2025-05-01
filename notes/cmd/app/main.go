package main

import (
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/mcore/pkg/logger"
)

type MyConfig struct {
	RepoPath        string `default:"data/fsnotes" usage:" path to repository fsnotes"`
	RepoUrl         string `required:"true"`
	RepoAccessToken string `env:"REPOACCESSTOKEN" required:"true"`
	MCoreAddr       string `default:"http://127.0.0.1:8080" usage:"host and port for mcore"`
	MCoreSecret     string `required:"true" usage:"secret key for api"`
	Debug           bool   `default:"false" usage:"turn on debug mode"`
}

var (
	cfg MyConfig
)

func main() {

	//config init
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"conf/notes_config.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})
	if err := loader.Load(); err != nil {
		logger.Fatal(err.Error())
	}

	if cfg.Debug {
		logger.SetLogLevel(logger.DEBUG)
		logger.Debugf("debug logging enabled")
	}

	fmt.Print("done")

}

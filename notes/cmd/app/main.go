package main

import (
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/mcore/pkg/logger"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"github.com/ishua/a3bot6/notes/internal/clients/gitapi"
	"github.com/ishua/a3bot6/notes/internal/domain"
)

type MyConfig struct {
	GitPath        string `default:"data/fsnotes" usage:" path to repository fsnotes"`
	GitUrl         string `required:"true" usage:"url for repository"`
	GitAccessToken string `env:"REPOACCESSTOKEN" required:"true"`
	GitEmail       string `required:"true" usage:"email for repo commits"`
	MCoreAddr      string `default:"http://127.0.0.1:8080" usage:"host and port for mcore"`
	MCoreSecret    string `required:"true" usage:"secret key for api"`
	Debug          bool   `default:"false" usage:"turn on debug mode"`
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
	fmt.Println(cfg.GitAccessToken)

	gc, err := gitapi.NewClient(cfg.GitPath, cfg.GitUrl, cfg.GitAccessToken, "bot notes", cfg.GitEmail)
	if err != nil {
		logger.Fatal(err.Error())
	}
	model := domain.NewModel(gc)
	s, err := model.Execute(schema.TaskNoteReadInbox, "")
	if err != nil {
		logger.Fatal(err.Error())
	}

	fmt.Println(s)
	fmt.Println("done")

}

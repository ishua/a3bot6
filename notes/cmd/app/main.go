package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/mcore/pkg/logger"
	"github.com/ishua/a3bot6/mcore/pkg/mcoreclient"
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
	DiaryPath      string `default:"Diary/5BX.markdown" usage:"file for diary"`
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
	model := domain.NewModel(gc, cfg.DiaryPath)

	mcore := mcoreclient.NewClient(cfg.MCoreAddr, cfg.MCoreSecret)
	ctx, cancel := context.WithCancel(context.Background())

	mcore.ListeningTasks(ctx, schema.TaskTypeNote, model, time.Duration(1*time.Second))
	log.Println("listen mcore")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// waiting signal for stop
	sig := <-sigChan
	log.Printf("Received signal: %s. Stopping...\n", sig)
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Program has stopped.")

}

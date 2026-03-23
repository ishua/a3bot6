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
	"github.com/ishua/a3bot6/synoc/internal/synology"
	"github.com/ishua/a3bot6/synoc/internal/worker"
)

type Config struct {
	Synology struct {
		URL      string            `yaml:"url"`
		Username string            `yaml:"username"`
		Password string            `yaml:"password"`
		Paths    map[string]string `yaml:"paths"`
	} `yaml:"synology"`
	MCore struct {
		Addr   string `yaml:"addr"`
		Secret string `yaml:"secret"`
	} `yaml:"mcore"`
	Debug bool `yaml:"debug"`
}

func main() {
	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"conf/config.yaml"},
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

	log.Println("Starting synoc service...")

	synoClient := synology.NewClient(cfg.Synology.URL)

	log.Println("Login to Synology...")
	err := synoClient.Login(cfg.Synology.Username, cfg.Synology.Password)
	if err != nil {
		logger.Fatal(fmt.Sprintf("login failed: %v", err))
	}
	log.Println("Login successful")

	w := worker.New(synoClient, cfg.Synology.Paths)

	mcore := mcoreclient.NewClient(cfg.MCore.Addr, cfg.MCore.Secret)
	ctx, cancel := context.WithCancel(context.Background())

	mcore.ListeningTasks(ctx, schema.TaskTypeSyno, w, time.Duration(1*time.Second))
	log.Println("listen mcore")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Received signal: %s. Stopping...\n", sig)
	cancel()

	log.Println("Logout from Synology...")
	synoClient.Logout()

	time.Sleep(1 * time.Second)
	log.Println("Program has stopped.")
}

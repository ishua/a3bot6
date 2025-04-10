package main

import (
	"github.com/ishua/a3bot6/mcore/internal/dialogmng"
	"github.com/ishua/a3bot6/mcore/internal/rest"
	"github.com/ishua/a3bot6/mcore/internal/routing"
	"github.com/ishua/a3bot6/mcore/internal/taskmng"
	"github.com/ishua/a3bot6/mcore/pkg/logger"
	_ "github.com/mattn/go-sqlite3"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/mcore/internal/storage/msqlclient"
)

type MyConfig struct {
	HttpPort       string   `default:"8080" usage:"port where start http rest"`
	Debug          bool     `default:"false" usage:"turn on debug mode"`
	SqliteFileName string   `default:"sql.db" usage:"path to sqllite db"`
	Secrets        []string `usage:"secrets for api"`
	Users          []string `usage:"users bot allowed"`
}

var (
	cfg MyConfig
)

func main() {

	//config init
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"conf/mcore_config.yaml"},
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
	if len(cfg.Secrets) == 0 {
		logger.Fatal("no secrets configured")
	}

	if len(cfg.Users) == 0 {
		logger.Fatal("no users configured")
	}

	//db init
	db := msqlclient.NewSqlClient(cfg.SqliteFileName)
	defer db.DbClose()

	taskMng := taskmng.NewTaskMng(db)
	dialogMng := dialogmng.NewDialogMng(db)

	router := routing.NewRouter(cfg.Users, dialogMng, taskMng)

	server := rest.NewApi("", taskMng, router, cfg.Debug, cfg.Secrets, cfg.HttpPort)
	err := server.Run()
	if err != nil {
		logger.Info(err.Error())
	}

}

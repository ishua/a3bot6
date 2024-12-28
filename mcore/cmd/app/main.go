package main

import (
	"github.com/ishua/a3bot6/mcore/internal/rest"
	"github.com/ishua/a3bot6/mcore/internal/routing"
	"github.com/ishua/a3bot6/mcore/internal/taskmng"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/mcore/internal/storage/msqlclient"
)

type MyConfig struct {
	HttpPort       string `default:"8080" usage:"port where start http rest"`
	Debug          bool   `default:"false" usage:"turn on debug mode"`
	SqliteFileName string `default:"sql.db" usage:"path to sqllite db"`
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
		log.Fatal(err)
	}

	//db init
	db := msqlclient.NewSqlClient(cfg.SqliteFileName)
	defer db.DbClose()

	router := routing.NewRouter([]string{"AlekseyIm"}, db)
	taskMng := taskmng.NewClient(db)

	server := rest.NewApi("", taskMng, router, cfg.Debug, []string{"test"}, cfg.HttpPort)
	err := server.Run()
	if err != nil {
		log.Println(err)
	}

	//m := schema.Message{
	//	UserName:         "AlekseyIm",
	//	ChatId:           10,
	//	MessageId:        11,
	//	ReplyToMessageID: 9,
	//	Text:             "y https://youtube.com/sdjfkl",
	//}
	//res := router.Build(m)
	//fmt.Println(res)

	//taskMng := taskmng.NewClient(db)
	//task, err := taskMng.GetTask(schema.TaskTypeYtdl)
	//if err != nil {
	//	log.Fatalf(err.Error())
	//}
	//fmt.Println(task)
	//
	//err = taskMng.ReportTask(task.Id, schema.TaskStatusDone, "ok")
	//if err != nil {
	//	log.Fatalf(err.Error())
	//}
	//fmt.Println("donnee")

}

package msqlclient

import (
	"database/sql"
	"log"
	"os"
	"path"
)

func initDBIfNeeded(dirPath, fileName string) {
	dbPath := path.Join(dirPath, fileName)
	_, err := os.Stat(dbPath)
	if err == nil {
		return
	}

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatalf("cant create path %s", err.Error())
	}
	_, err = os.Create(dbPath)
	if err != nil {
		log.Fatalf("cant create file %s", err.Error())
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("no open db %s", err.Error())
	}
	defer db.Close()

	_, err = db.Exec(creatTask)
	if err != nil {
		log.Fatalf("create table creatTask %s", err.Error())
	}

	_, err = db.Exec(createDialog)
	if err != nil {
		log.Fatalf("create table createDialog %s", err.Error())
	}
}

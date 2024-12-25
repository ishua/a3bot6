package msqlclient

import (
	"database/sql"
	"log"
	"path"
)

type SqliteClient struct {
	db *sql.DB
}

const (
	dataPath = "data"
)

func NewSqlClient(dbFileName string) *SqliteClient {
	initDBIfNeeded(dataPath, dbFileName)
	dbPath := path.Join(dataPath, dbFileName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("no open db %s", err.Error())
	}

	var version string
	err = db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("sqlite version: %s", version)

	return &SqliteClient{
		db: db,
	}
}

func (c *SqliteClient) DbClose() error {
	return c.db.Close()
}

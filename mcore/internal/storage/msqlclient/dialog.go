package msqlclient

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/internal/schema"
)

const createDialog string = `
CREATE TABLE IF NOT EXISTS dialog (
    id INTEGER NOT NULL PRIMARY KEY,
	key TEXT NOT NULL,
	dialogstatus INTEGER NOT NULL,
	data blob null default (x'') 
  );
`

func (c *SqliteClient) AddDialog(d schema.Dialog) (int64, error) {
	data, err := d.GetMessagesAsByte()
	if err != nil {
		return 0, fmt.Errorf("addDialog can't parse messages: %w", err)
	}

	sqlQuery := `INSERT INTO dialog( key, dialogstatus, data) VALUES( ?, ?, ?);`
	res, err := c.db.Exec(sqlQuery, d.Key, d.DialogStatus, data)
	if err != nil {
		return 0, fmt.Errorf("insert addDialog: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert addDialog can't return id: %w", err)
	}

	return id, nil
}

func (c *SqliteClient) GetDialogById(id int64) (schema.Dialog, error) {
	sqlQuery := "SELECT id, key, dialogstatus, data FROM dialog WHERE id = ?"

	row := c.db.QueryRow(sqlQuery, id)
	d := &schema.Dialog{}
	var data []byte
	err := row.Scan(&d.Id, &d.Key, &d.DialogStatus, &data)
	if err != nil {
		return *d, fmt.Errorf("getDialogById scan %w", err)
	}
	err = d.SetMessagesFromByte(data)
	if err != nil {
		return *d, fmt.Errorf("getDialogById unmarshal %w", err)
	}
	return *d, err
}

func (c *SqliteClient) UpdateDialog(d schema.Dialog) error {
	if d.Id == 0 {
		return fmt.Errorf("updateDialog dialog.id is 0 nothink to update")
	}

	sqlQuery := "UPDATE dialog SET dialogstatus = ?, data = ? WHERE id = ?"
	msgByte, err := d.GetMessagesAsByte()
	if err != nil {
		return fmt.Errorf("updateDialog cat't marshal messages %w", err)
	}
	_, err = c.db.Exec(sqlQuery, d.DialogStatus, msgByte, d.Id)
	if err != nil {
		return fmt.Errorf("updateDialog : %w", err)
	}
	return nil
}

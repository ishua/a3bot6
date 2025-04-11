package msqlclient

func (c *SqliteClient) DeleteAllTasks() error {
	sqlQuery := "delete from task;"
	_, err := c.db.Exec(sqlQuery)
	return err
}

func (c *SqliteClient) DeleteAllDialogs() error {
	sqlQuery := "delete from dialog;"
	_, err := c.db.Exec(sqlQuery)
	return err
}

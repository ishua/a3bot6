package msqlclient

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

const creatTask = `
CREATE TABLE IF NOT EXISTS task (
	id integer NOT NULL PRIMARY KEY,
 	dialog INTEGER NOT NULL,
  	status INTEGER NOT NULL,
  	type INTEGER NOT NULL,
  	data blob null default (x'') 
  );
`

func (c *SqliteClient) AddTask(task schema.Task) (int64, error) {
	if task.Type == schema.TaskTypeUndefined {
		return 0, errors.New("invalid task type")
	}

	sqlQuery := `
INSERT INTO task( dialog, status, type, data)
	VALUES( ?, ?, ?, ?);
	`

	data, err := task.TaskData.Marshal()
	if err != nil {
		return 0, fmt.Errorf("addtask task data marshal: %w", err)
	}
	res, err := c.db.Exec(sqlQuery, task.DialogId, task.Status, task.Type, data)
	if err != nil {
		return 0, fmt.Errorf("insert addTask: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert addTask can't return id: %w", err)
	}

	return id, nil
}

func (c *SqliteClient) UpdateTaskStatus(task schema.Task) error {
	if task.Id == 0 {
		return fmt.Errorf("smt is wrong try to updata task without id")
	}

	sqlQuery := "UPDATE task SET status = ? WHERE id = ?"

	_, err := c.db.Exec(sqlQuery, task.Status, task.Id)
	if err != nil {
		return fmt.Errorf("updateTaskStatus : %w", err)
	}
	return nil
}

func (c *SqliteClient) getTaskFromRow(row *sql.Row) (schema.Task, error) {
	t := &schema.Task{}
	var data []byte

	err := row.Scan(&t.Id, &t.DialogId, &t.Status, &t.Type, &data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return schema.Task{}, nil
		}
		return schema.Task{}, fmt.Errorf("getTaskFromRow scan %w", err)
	}
	err = t.TaskData.Unmarshal(data)
	if err != nil {
		return schema.Task{}, fmt.Errorf("getTaskFromRow unmarshal %w", err)
	}
	return *t, err
}

func (c *SqliteClient) GetTaskById(id int64) (schema.Task, error) {
	sqlQuery := `
select id, dialog, status, type, data from task where id = ?
`
	return c.getTaskFromRow(c.db.QueryRow(sqlQuery, id))
}

func (c *SqliteClient) GetFirstTaskByType(t schema.TaskType) (schema.Task, error) {
	if t == schema.TaskTypeUndefined {
		return schema.Task{}, fmt.Errorf("getFirstTaskByType: wrong task type")
	}
	sqlQuery := `
SELECT id, dialog, status, type, data FROM task WHERE type = ? and status = ? ORDER BY ID LIMIT 1
`
	return c.getTaskFromRow(c.db.QueryRow(sqlQuery, t, schema.TaskStatusCreate))
}

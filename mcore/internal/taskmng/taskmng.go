package taskmng

import (
	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

type Mng struct {
	repo repo
}

func NewTaskMng(repo repo) *Mng {
	return &Mng{
		repo: repo,
	}
}

type repo interface {
	AddTask(task schema.Task) (int64, error)
	GetFirstTaskByType(t schema.TaskType) (schema.Task, error)
	GetTaskById(id int64) (schema.Task, error)
	GetDialogById(id int64) (schema.Dialog, error)
	UpdateTaskStatus(task schema.Task) error
	UpdateDialog(d schema.Dialog) error
}

func (m *Mng) GetTask(taskType schema.TaskType) (schema.Task, error) {
	return m.repo.GetFirstTaskByType(taskType)
}

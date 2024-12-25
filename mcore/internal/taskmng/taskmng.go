package taskmng

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/internal/schema"
	"log"
)

type Client struct {
	repo repo
}

func NewClient(repo repo) *Client {
	return &Client{
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

func (t *Client) ReportTask(taskId int64, status schema.TaskStatus, textMsg string) error {
	task, err := t.repo.GetTaskById(taskId)
	if err != nil {
		return fmt.Errorf("reportTask %w", err)
	}
	task.Status = status

	dialog, err := t.repo.GetDialogById(task.DialogId)
	if err != nil {
		return fmt.Errorf("reportTask %w", err)
	}
	if status == schema.TaskStatusError {
		dialog.DialogStatus = schema.DialogStatusError
	}
	if status == schema.TaskStatusDone {
		dialog.DialogStatus = schema.DialogStatusClose
	}

	newTask := schema.Task{
		DialogId: dialog.Id,
		Status:   schema.TaskStatusCreate,
		Type:     schema.TaskTypeMsg,
		TaskData: schema.TaskData{
			Msg: schema.TaskMsg{
				ChatId:         dialog.Messages[0].ChatId,
				ReplyMessageId: dialog.Messages[0].MessageId,
				Text:           textMsg,
			},
		},
	}

	_, err = t.repo.AddTask(newTask)
	if err != nil {
		return fmt.Errorf("reportTask %w", err)
	}

	err = t.repo.UpdateTaskStatus(task)
	if err != nil {
		log.Printf("reportTask UpdateTaskStatus %w", err)
	}
	err = t.repo.UpdateDialog(dialog)
	if err != nil {
		log.Printf("reportTask UpdateDialog %w", err)
	}
	return nil
}

func (t *Client) GetTask(taskType schema.TaskType) (schema.Task, error) {
	task, err := t.repo.GetFirstTaskByType(taskType)
	if err != nil {
		return task, fmt.Errorf("getTask %w", err)
	}
	if task.Id == 0 {
		return task, err
	}
	task.Status = schema.TaskStatusSended
	err = t.repo.UpdateTaskStatus(task)
	if err != nil {
		return task, fmt.Errorf("getTask %w", err)
	}
	return task, nil
}

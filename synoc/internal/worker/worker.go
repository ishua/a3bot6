package worker

import (
	"fmt"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"github.com/ishua/a3bot6/synoc/internal/synology"
)

type SynoClient interface {
	CreateTask(uri, destination, filename string) (string, error)
	ListTasks() ([]synology.Task, error)
	DeleteTask(taskID string) error
}

type Worker struct {
	client SynoClient
	paths  map[string]string
}

func New(client SynoClient, paths map[string]string) *Worker {
	return &Worker{
		client: client,
		paths:  paths,
	}
}

func (w *Worker) DoTask(task schema.Task) schema.ReportTaskReq {
	syno := task.TaskData.Syno

	switch syno.Command {
	case schema.SynoTaskCmdAdd:
		return w.handleAdd(task, syno)
	case schema.SynoTaskCmdList:
		return w.handleList(task)
	case schema.SynoTaskCmdDelete:
		return w.handleDelete(task, syno)
	default:
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("unknown command: %s", syno.Command),
		}
	}
}

func (w *Worker) handleAdd(task schema.Task, syno schema.TaskSyno) schema.ReportTaskReq {
	destination := ""
	if syno.Category != "" {
		if path, ok := w.paths[string(syno.Category)]; ok {
			destination = path
		}
	}

	taskID, err := w.client.CreateTask(syno.TorrentUrl, destination, "")
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("add task failed: %v", err),
		}
	}

	return schema.ReportTaskReq{
		TaskId:  task.Id,
		Status:  schema.TaskStatusDone,
		TextMsg: fmt.Sprintf("task created: %s", taskID),
	}
}

func (w *Worker) handleList(task schema.Task) schema.ReportTaskReq {
	tasks, err := w.client.ListTasks()
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("list tasks failed: %v", err),
		}
	}

	msg := fmt.Sprintf("found %d tasks", len(tasks))
	for _, t := range tasks {
		msg += fmt.Sprintf("\n%s: %s (%s)", t.ID, t.Title, t.Status)
	}

	return schema.ReportTaskReq{
		TaskId:  task.Id,
		Status:  schema.TaskStatusDone,
		TextMsg: msg,
	}
}

func (w *Worker) handleDelete(task schema.Task, syno schema.TaskSyno) schema.ReportTaskReq {
	if syno.TaskId == "" {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: "taskId is required for delete",
		}
	}

	err := w.client.DeleteTask(syno.TaskId)
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("delete failed: %v", err),
		}
	}

	return schema.ReportTaskReq{
		TaskId:  task.Id,
		Status:  schema.TaskStatusDone,
		TextMsg: fmt.Sprintf("task %s deleted", syno.TaskId),
	}
}

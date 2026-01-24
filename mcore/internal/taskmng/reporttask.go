package taskmng

import (
	"fmt"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

func (m *Mng) ReportTask(taskId int64, status schema.TaskStatus, msg string) error {
	task, err := m.repo.GetTaskById(taskId)
	if err != nil {
		return fmt.Errorf("reportTask getTask err: %w", err)
	}
	dialog, err := m.repo.GetDialogById(task.DialogId)
	if err != nil {
		return fmt.Errorf("reportTask getDialog err: %w", err)
	}

	task.Status = status
	err = m.repo.UpdateTaskStatus(task)
	if err != nil {
		return fmt.Errorf("reportTask updateTaskStatus err: %w", err)
	}

	if status == schema.TaskStatusSended {
		return nil
	}

	if status == schema.TaskStatusError {
		dialog.DialogStatus = schema.DialogStatusError
	}

	if status == schema.TaskStatusDone {
		dialog.DialogStatus = schema.DialogStatusClose
	}
	err = m.repo.UpdateDialog(dialog)
	if err != nil {
		return fmt.Errorf("reportTask updateDialog err: %w", err)
	}

	if task.Type == schema.TaskTypeMsg {
		return nil
	}

	if msg == "" {
		return nil
	}

	replyTask := schema.Task{
		DialogId: dialog.Id,
		Type:     schema.TaskTypeMsg,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Msg: schema.TaskMsg{
				ChatId:         dialog.Messages[0].ChatId,
				ReplyMessageId: dialog.Messages[0].MessageId,
				Text:           msg,
			},
		},
	}
	_, err = m.repo.AddTask(replyTask)
	if err != nil {
		return fmt.Errorf("reportTask addTask err: %w", err)
	}
	return nil
}

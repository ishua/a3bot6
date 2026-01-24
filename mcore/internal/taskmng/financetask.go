package taskmng

import (
	"fmt"
	"strings"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

func (m *Mng) createFinanceTask(dialogId int64, userText string) (string, error) {
	words := strings.Split(userText, " ")
	if len(words) < 2 {
		return "", fmt.Errorf("for finance need command")
	}

	task := schema.Task{
		DialogId: dialogId,
		Type:     schema.TaskTypeFinance,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Fin: schema.FinanceTask{},
		},
	}

	switch words[1] {
	case "run", "r":
		task.TaskData.Fin.Command = "run"
	case "load", "l":
		task.TaskData.Fin.Command = "load"
	case "transactions", "t":
		task.TaskData.Fin.Command = "transactions"
	default:
		return "", fmt.Errorf("unknown command")
	}

	_, err := m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}

	return "task note created", nil
}

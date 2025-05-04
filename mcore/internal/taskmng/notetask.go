package taskmng

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"strings"
)

func (m *Mng) createNoteTask(dialogId int64, text string) (string, error) {
	words := strings.Split(text, " ")
	if len(words) < 2 {
		return "", fmt.Errorf("for note need command")
	}
	var err error
	if words[0] != "/note" {
		words, err = parseSimplifications(words)
		if err != nil {
			return "", err
		}
	}

	task := schema.Task{
		DialogId: dialogId,
		Type:     schema.TaskTypeNote,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Tn: schema.TaskNote{},
		},
	}
	switch words[1] {
	case "diary":
		{
			if !(words[2] == "entry" || words[2] == "5bx") {
				return "", fmt.Errorf("for diary need label")
			}
			task.TaskData.Tn = schema.TaskNote{
				Command: schema.TaskNoteCmdAddDiary,
				AddText: strings.Join(words[2:], " "),
			}
		}
	case "inbox":
		{
			switch words[2] {
			case "add":
				{
					task.TaskData.Tn = schema.TaskNote{
						Command: schema.TaskNoteCmdAddInbox,
						AddText: strings.Join(words[3:], " "),
					}
				}
			case "read":
				{
					task.TaskData.Tn = schema.TaskNote{
						Command: schema.TaskNoteReadInbox,
					}
				}
			default:
				return "", fmt.Errorf("for inbox need label")
			}
		}
	case "pull":
		{
			task.TaskData.Tn = schema.TaskNote{
				Command: schema.TaskNoteCmdPull,
				AddText: "",
			}
		}
	case "help":
		{
			return tnHelpText, nil
		}
	default:
		return "", fmt.Errorf("unknown command")
	}

	_, err = m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}

	return "task note created", nil
}

func parseSimplifications(words []string) ([]string, error) {
	switch words[0] {
	case "nd", "Nd":
		return append([]string{"/note", "diary", "entry"}, words[1:]...), nil
	case "n5", "N5":
		return append([]string{"/note", "diary", "5bx"}, words[1:]...), nil
	case "ni", "Ni":
		return append([]string{"/note", "inbox", "add"}, words[1:]...), nil
	case "nir", "Nir":
		return append([]string{"/note", "inbox", "read"}, words[1:]...), nil
	}
	return nil, fmt.Errorf("synonyms command not found")
}

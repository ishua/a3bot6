package taskmng

import (
	"fmt"
	"strings"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

const (
	ntAlias = "nt"
	NtAlias = "Nt"
)

func (m *Mng) createNoteTask(dialogId int64, text string) (string, error) {
	words := strings.Split(text, " ")
	if len(words) < 2 {
		// Bare aliases still need a command, except nt/Nt which parseSimplifications
		// maps to "/note task" (empty AddText -> default).
		if words[0] != ntAlias && words[0] != NtAlias {
			return "", fmt.Errorf("for note need command")
		}
	}
	var err error
	if words[0] != "/note" {
		words, err = parseSimplifications(words)
		if err != nil {
			return "", err
		}
	}

	if words[1] == "help" {
		return tnHelpText, nil
	}

	taskNote, err := parseNoteTaskArgs(words)
	if err != nil {
		return "", fmt.Errorf("parse note task args: %w", err)
	}

	task := schema.Task{
		DialogId: dialogId,
		Type:     schema.TaskTypeNote,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Tn: taskNote,
		},
	}

	_, err = m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}

	return "task note created", nil
}

func parseNoteTaskArgs(words []string) (schema.TaskNote, error) {
	if len(words) < 2 {
		return schema.TaskNote{}, fmt.Errorf("for note need command")
	}

	cmdByKey := map[string]schema.TaskNoteCmd{
		"5bx":    schema.TaskNoteCmdAdd5bx,
		"entry":  schema.TaskNoteCmdAddEntry,
		"weight": schema.TaskNoteCmdAddWeight,
		"bp":     schema.TaskNoteCmdAddBP,
	}
	if cmd, ok := cmdByKey[words[1]]; ok {
		return schema.TaskNote{
			Command: cmd,
			AddText: strings.Join(words[2:], " "),
		}, nil
	}

	switch words[1] {
	case "inbox":
		if len(words) < 3 {
			return schema.TaskNote{}, fmt.Errorf("for inbox need label")
		}
		switch words[2] {
		case "add":
			return schema.TaskNote{
				Command: schema.TaskNoteCmdAddInbox,
				AddText: strings.Join(words[3:], " "),
			}, nil
		case "read":
			return schema.TaskNote{
				Command: schema.TaskNoteReadInbox,
			}, nil
		default:
			return schema.TaskNote{}, fmt.Errorf("for inbox need label")
		}
	case "pull":
		return schema.TaskNote{
			Command: schema.TaskNoteCmdPull,
			AddText: "",
		}, nil
	case "task", "tasks":
		var tokens []string
		for _, w := range words[2:] {
			if w != "" {
				tokens = append(tokens, w)
			}
		}
		addText := strings.Join(tokens, " ")
		return schema.TaskNote{
			Command: schema.TaskNoteCmdTaskList,
			AddText: addText,
		}, nil
	default:
		return schema.TaskNote{}, fmt.Errorf("unknown command %q", words[1])
	}
}

func parseSimplifications(words []string) ([]string, error) {
	switch words[0] {
	case "nd", "Nd":
		return append([]string{"/note", "entry"}, words[1:]...), nil
	case "n5", "N5":
		return append([]string{"/note", "5bx"}, words[1:]...), nil
	case "ni", "Ni":
		return append([]string{"/note", "inbox", "add"}, words[1:]...), nil
	case "nir", "Nir":
		return append([]string{"/note", "inbox", "read"}, words[1:]...), nil
	case "nw", "Nw":
		return append([]string{"/note", "weight"}, words[1:]...), nil
	case "nbp", "Nbp":
		return append([]string{"/note", "bp"}, words[1:]...), nil
	case ntAlias, NtAlias:
		return append([]string{"/note", "task"}, words[1:]...), nil
	}
	return nil, fmt.Errorf("synonyms command not found")
}

package taskmng

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

func (m *Mng) createTrTask(dialogId int64, text string, torrentUrl string) (string, error) {
	w := strings.Split(text, " ")
	if len(w) < 2 {
		return "", fmt.Errorf("for tr need command")
	}
	var folderPath string
	var torrentId int
	var err error
	var command string

	if w[1] == "add" {
		if len(w) < 3 {
			return "", fmt.Errorf("for tr add need label")
		}
		folderPath, err = chooseFolderPath(w[2])
		if err != nil {
			return "", err
		}

		if len(torrentUrl) == 0 {
			return "", fmt.Errorf("for tr add need torrent url")
		}
		command = w[1]
	}

	if w[1] == "del" {
		if len(w) < 3 {
			return "", fmt.Errorf("for tr del need id")
		}
		torrentId, err = strconv.Atoi(w[2])
		if err != nil {
			return "", fmt.Errorf("for tr del id is not an int")
		}
		command = w[1]
	}

	if w[1] == "list" {
		if len(w) != 2 {
			return "", fmt.Errorf("list command does not have arguments")
		}

		command = w[1]
	}

	if w[1] == "help" {
		return trHelpText, nil
	}
	if command == "" {
		return "", fmt.Errorf("command not found")
	}

	task := schema.Task{
		DialogId: dialogId,
		Type:     schema.TaskTypeTorrent,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Tr: schema.TaskTr{
				FolderPath: folderPath,
				TorrentUrl: torrentUrl,
				TorrentId:  torrentId,
				Command:    command,
			},
		},
	}

	_, err = m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}

	return "task tr created", nil
}

func chooseFolderPath(label string) (string, error) {
	switch label {
	case "m", "movie":
		return "movie", nil
	case "s", "shows":
		return "shows", nil
	case "c", "cartoon":
		return "cartoon", nil
	case "a", "audiobook":
		return "audiobook", nil
	case "ap", "audiobook_p":
		return "audiobook_p", nil
	case "cs", "cartoon_s":
		return "cartoon_s", nil
	}

	return "", fmt.Errorf("wrong label")
}

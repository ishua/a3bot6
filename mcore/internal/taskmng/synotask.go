package taskmng

import (
	"fmt"
	"strings"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

const synoHelpText = `This is help for /ds (Download Station) commands:
- /ds add <category> - add torrent/magnet link. 
  Categories: movie/m, cartoon/c, shows/s, audiobook/a, other/o, shows_cartoons/cs
  Example: /ds add movie https://example.com/file.torrent
  Shortcuts: dsm, dsc, dss, dsa, dso, dscs
- /ds list - show active downloads
- /ds del <id> - delete download task by id
- /ds help - show this help
`

func (m *Mng) createSynoTask(dialogId int64, text string, fileUrl string) (string, error) {
	words := strings.Split(text, " ")
	if len(words) < 2 {
		return "", fmt.Errorf("for ds need command")
	}

	var err error
	if words[0] != "/ds" {
		words, err = parseSynoSimplifications(words)
		if err != nil {
			return "", err
		}
	}

	task := schema.Task{
		DialogId: dialogId,
		Type:     schema.TaskTypeSyno,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{},
		},
	}

	switch words[1] {
	case "add":
		{
			if len(words) < 3 {
				return "", fmt.Errorf("for ds add need category")
			}

			category, err := parseSynoCategory(words[2])
			if err != nil {
				return "", err
			}

			torrentUrl := fileUrl
			if len(words) > 3 {
				torrentUrl = words[3]
			}

			if len(torrentUrl) == 0 {
				return "", fmt.Errorf("for ds add need torrent url or file attachment")
			}

			task.TaskData.Syno = schema.TaskSyno{
				Command:    schema.SynoTaskCmdAdd,
				Category:   category,
				TorrentUrl: torrentUrl,
			}
		}
	case "list":
		{
			if len(words) != 2 {
				return "", fmt.Errorf("list command does not have arguments")
			}
			task.TaskData.Syno = schema.TaskSyno{
				Command: schema.SynoTaskCmdList,
			}
		}
	case "del", "delete":
		{
			if len(words) < 3 {
				return "", fmt.Errorf("for ds del need task id")
			}
			task.TaskData.Syno = schema.TaskSyno{
				Command: schema.SynoTaskCmdDelete,
				TaskId:  words[2],
			}
		}
	case "help":
		{
			return synoHelpText, nil
		}
	default:
		return "", fmt.Errorf("unknown command")
	}

	_, err = m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}

	return "task syno created", nil
}

func parseSynoCategory(label string) (schema.SynoCategory, error) {
	switch label {
	case "m", "movie":
		return schema.SynoCategoryMovie, nil
	case "c", "cartoon":
		return schema.SynoCategoryCartoon, nil
	case "s", "shows":
		return schema.SynoCategoryShows, nil
	case "a", "audiobook":
		return schema.SynoCategoryAudiobook, nil
	case "o", "other":
		return schema.SynoCategoryOther, nil
	case "cs", "shows_cartoons":
		return schema.SynoCategoryShowsCartoons, nil
	}
	return "", fmt.Errorf("unknown category: %s", label)
}

func parseSynoSimplifications(words []string) ([]string, error) {
	switch words[0] {
	case "dsm", "Dsm":
		return append([]string{"/ds", "add", "movie"}, words[1:]...), nil
	case "dsc", "Dsc":
		return append([]string{"/ds", "add", "cartoon"}, words[1:]...), nil
	case "dss", "Dss":
		return append([]string{"/ds", "add", "shows"}, words[1:]...), nil
	case "dsa", "Dsa":
		return append([]string{"/ds", "add", "audiobook"}, words[1:]...), nil
	case "dso", "Dso":
		return append([]string{"/ds", "add", "other"}, words[1:]...), nil
	case "dscs", "Dscs":
		return append([]string{"/ds", "add", "shows_cartoons"}, words[1:]...), nil
	case "dsl", "Dsl":
		return []string{"/ds", "list"}, nil
	}
	return nil, fmt.Errorf("synonyms command not found")
}

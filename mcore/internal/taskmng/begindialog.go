package taskmng

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"net/url"
	"strconv"
	"strings"
)

const trHelpText = ` This is help for /torrent command
start words: /torrent, t, T
next words:
- "add" + category: "movie/m","shows/s", "cartoon/c", "cartoon_s/cs", "audiobook/a", "audiobook_p/ap" - add torrents in category
- "list" - list torrens in actions
- "del #" - where # it is id torrent
`

const helpCommon = ` My commands:
- /help
- /ping
- /y2d
- /torrent
`

func (m *Mng) ProcessDialogBegin(dialogId int64) (string, error) {
	dialog, err := m.repo.GetDialogById(dialogId)
	if err != nil {
		return "", fmt.Errorf("taskMng get dialog by id: %w", err)
	}
	if dialog.DialogStatus != schema.DialogStatusBegin {
		return "", fmt.Errorf("wrong dialog status")
	}

	if len(dialog.Messages) > 1 {
		return "", fmt.Errorf("wrong number of messages")
	}

	userText := dialog.Messages[0].Text
	if len(userText) == 0 {
		userText = dialog.Messages[0].Caption
		if len(userText) == 0 {
			return "", fmt.Errorf("dialog text and captions is empty")
		}
	}

	return m.createReply(dialogId, dialog.Messages[0].UserName, userText, dialog.Messages[0].FileUrl)
}

func (m *Mng) createReply(dialogId int64, userName string, userText string, fileUrl string) (string, error) {
	s := strings.Split(userText, " ")
	switch s[0] {
	case "/help", "h", "H":
		return helpCommon, nil
	case "/ping", "ping", "Ping":
		return "Pong", nil
	case "/y2d", "y", "Y":
		return m.createYtdlTask(dialogId, userName, userText)
	case "/torrent", "torrent", "t", "T":
		return m.createTrTask(dialogId, userText, fileUrl)
	}

	return "", fmt.Errorf("command not found")
}

func (m *Mng) createYtdlTask(dialogId int64, userName string, text string) (string, error) {
	w := strings.Split(text, " ")
	if len(w) < 2 {
		return "", fmt.Errorf("for y2d need a link")
	}
	u, err := url.Parse(w[1])
	if err != nil {
		return "", fmt.Errorf("can't parse url %w", err)
	}

	if u.Host != "youtube.com" && u.Host != "www.youtube.com" && u.Host != "youtu.be" {
		return "", fmt.Errorf("host: %s not yuotube", u.Host)
	}

	task := schema.Task{
		DialogId: dialogId,
		Type:     schema.TaskTypeYtdl,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Ytdl: schema.TaskYtdl{
				Link:     w[1],
				UserName: userName,
			},
		},
	}

	_, err = m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}
	return "task ytd created", nil
}

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

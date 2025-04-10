package taskmng

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"net/url"
	"strings"
)

func (m Mng) ProcessDialogBegin(dialogId int64) (string, error) {
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

	return m.createReply(dialogId, userText)
}

func (m Mng) createReply(dialogId int64, userText string) (string, error) {
	s := strings.Split(userText, " ")
	switch s[0] {
	case "/help", "h", "H":
		return helpDialog(), nil
	case "/ping", "ping", "Ping":
		return "Pong", nil
	case "/y2d", "y", "Y":
		return m.createYtdlTask(dialogId, userText)
	}

	return "", fmt.Errorf("command not found")
}

func helpDialog() string {
	return "help text"
}

func (m Mng) createYtdlTask(dialogId int64, text string) (string, error) {
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
				Link: w[1],
			},
		},
	}

	_, err = m.repo.AddTask(task)
	if err != nil {
		return "", fmt.Errorf("taskMng add task: %w", err)
	}
	return "task created", nil
}

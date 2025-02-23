package routing

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"net/url"
	"slices"
	"strings"
)

type Router struct {
	users []string
	repo  repo
}

func NewRouter(users []string, repo repo) *Router {
	return &Router{users: users, repo: repo}
}

type repo interface {
	AddDialog(dialog schema.Dialog) (int64, error)
	AddTask(task schema.Task) (int64, error)
}

func (r *Router) Build(m schema.Message) schema.TaskMsg {
	m.Type = schema.MessageTypeUser
	reply := schema.TaskMsg{
		ChatId:         m.ChatId,
		ReplyMessageId: m.MessageId,
	}
	if !slices.Contains(r.users, m.UserName) {
		reply.Text = fmt.Sprintf("I don't answer to user %s", m.UserName)
		return reply
	}

	task, err := getTask(m)
	if err != nil {
		reply.Text = err.Error()
		return reply
	}

	dialogId, err := r.repo.AddDialog(schema.Dialog{
		Key:          schema.GenerateKey(m),
		DialogStatus: schema.DialogStatusBegin,
		Messages:     []schema.Message{m},
	})
	if err != nil {
		reply.Text = err.Error()
		return reply
	}

	task.DialogId = dialogId
	_, err = r.repo.AddTask(task)
	if err != nil {
		reply.Text = err.Error()
		return reply
	}
	reply.Text = "OK"
	return reply
}

func getTask(m schema.Message) (schema.Task, error) {
	var task schema.Task
	userText := m.Text
	if len(userText) == 0 {
		userText = m.Caption
		if len(userText) == 0 {
			return task, fmt.Errorf("text is empty")
		}
	}

	s := strings.Split(userText, " ")
	switch s[0] {
	case "/y2d", "y", "Y":
		return buildYoutubeTask(s)
	case "/help", "h", "H":
		return buildHelpTask(m.ChatId)
	case "/ping", "ping", "Ping":
		return buildPongTask(m.ChatId, m.ReplyToMessageID)
	}

	return task, fmt.Errorf("command not found")

}

func buildYoutubeTask(w []string) (schema.Task, error) {
	if len(w) < 2 {
		return schema.Task{}, fmt.Errorf("for y2d need a link")
	}

	u, err := url.Parse(w[1])
	if err != nil {
		return schema.Task{}, fmt.Errorf("can't parse url %w", err)
	}

	if u.Host != "youtube.com" && u.Host != "www.youtube.com" && u.Host != "youtu.be" {
		return schema.Task{}, fmt.Errorf("host: %s not yuotube", u.Host)
	}

	task := schema.Task{
		Type:   schema.TaskTypeYtdl,
		Status: schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Ytdl: schema.TaskYtdl{
				Link: w[1],
			},
		},
	}

	return task, nil
}

func buildHelpTask(chatId int64) (schema.Task, error) {
	text := "You can use next command: \n" +
		"- /y2d \n"
	task := schema.Task{
		Type:   schema.TaskTypeMsg,
		Status: schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Msg: schema.TaskMsg{
				ChatId: chatId,
				Text:   text,
			},
		},
	}
	return task, nil
}

func buildPongTask(chatId int64, replyMessageId int) (schema.Task, error) {
	text := "pong"
	task := schema.Task{
		Type:   schema.TaskTypeMsg,
		Status: schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Msg: schema.TaskMsg{
				ChatId:         chatId,
				Text:           text,
				ReplyMessageId: replyMessageId,
			},
		},
	}
	return task, nil
}

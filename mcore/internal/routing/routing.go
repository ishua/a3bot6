package routing

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"slices"
)

type Router struct {
	allowedUsers []string
	dialogMng    dialogMng
	taskMng      taskMng
}

type dialogMng interface {
	Create(m schema.Message) (int64, error)
}

type taskMng interface {
	ProcessDialogBegin(dialogId int64) (string, error)
}

func NewRouter(users []string, dialogMng dialogMng, taskMng taskMng) *Router {
	return &Router{allowedUsers: users, dialogMng: dialogMng, taskMng: taskMng}
}

func (r *Router) ProcessMsg(m schema.Message) schema.TaskMsg {
	m.Type = schema.MessageTypeUser
	reply := schema.TaskMsg{
		ChatId:         m.ChatId,
		ReplyMessageId: m.MessageId,
	}
	if !slices.Contains(r.allowedUsers, m.UserName) {
		reply.Text = fmt.Sprintf("I don't answer to user %s", m.UserName)
		return reply
	}

	dialogId, err := r.dialogMng.Create(m)
	if err != nil {
		reply.Text = err.Error()
		return reply
	}

	reply.Text, err = r.taskMng.ProcessDialogBegin(dialogId)
	if err != nil {
		reply.Text = err.Error()
		return reply
	}

	return reply
}

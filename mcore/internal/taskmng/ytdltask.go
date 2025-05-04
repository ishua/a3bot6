package taskmng

import (
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"net/url"
	"strings"
)

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

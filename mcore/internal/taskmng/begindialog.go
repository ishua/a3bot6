package taskmng

import (
	"fmt"
	"strings"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

const trHelpText = ` This is help for /torrent command
start words: /torrent, t, T
next words:
- "add" + category: "movie/m","shows/s", "cartoon/c", "cartoon_s/cs", "audiobook/a", "audiobook_p/ap" - add torrents in category
- "list" - list torrens in actions
- "del #" - where # it is id torrent
`

const tnHelpText = `This is help for /note commands:
- /note diary entry - add message to diary. Synonyms: nd/Nd
- /note diary 5bx - add 5bx line. Synonyms: n5/N5
- /note inbox add - add text to inbox. Synonyms: ni/Ni
- /note inbox read - read inbox. Synonyms: nir/Nir
- /note pull - just update repo
`

const helpCommon = ` My commands:
- /help
- /ping
- /y2d
- /torrent
- /finance
- /ds
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
	case "/note", "n", "nd", "Nd", "n5", "N5", "ni", "Ni", "nir", "Nir":
		return m.createNoteTask(dialogId, userText)
	case "health", "/health":
		return m.createHealth(dialogId)
	case "/finance", "f", "F":
		return m.createFinanceTask(dialogId, userText)
	case "/free":
		return m.createFreeTask(dialogId)
	case "/ds", "ds", "dsm", "Dsm", "dsc", "Dsc", "dss", "Dss", "dsa", "Dsa", "dso", "Dso", "dscs", "Dscs", "dsl", "Dsl":
		return m.createSynoTask(dialogId, userText, fileUrl)
	}

	return "", fmt.Errorf("command not found")
}

func (m *Mng) createHealth(dialogId int64) (string, error) {
	taskTemolata := schema.Task{
		DialogId: dialogId,
		Status:   schema.TaskStatusCreate,
		TaskData: schema.TaskData{
			Health: "health",
		},
	}

	taskTypes := []schema.TaskType{
		schema.TaskTypeNote,
		schema.TaskTypeYtdl,
		// schema.TaskTypeTorrent,
	}

	for _, taskType := range taskTypes {
		taskTemolata.Type = taskType
		_, err := m.repo.AddTask(taskTemolata)
		if err != nil {
			return "", fmt.Errorf("health type=%d: %w", taskTypes, err)
		}
	}
	return "tasks health created", nil

}

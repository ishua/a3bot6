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

	return "task tr created", nil
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

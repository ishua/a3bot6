package schema

import "encoding/json"

type TaskType int

const (
	TaskTypeUndefined = iota
	TaskTypeMsg
	TaskTypeYtdl //use in python project
	TaskTypeRest
	TaskTypeNote
	TaskTypeTorrent
)

type TaskStatus int

const (
	TaskStatusUndefined = iota
	TaskStatusCreate
	TaskStatusError
	TaskStatusSended
	TaskStatusDone
)

type Task struct {
	Id       int64      `json:"id"`
	DialogId int64      `json:"dialogId"`
	Status   TaskStatus `json:"status"`
	Type     TaskType   `json:"type"`
	TaskData TaskData   `json:"taskData"`
}

type TaskData struct {
	Ytdl TaskYtdl `json:"ytdl"`
	Msg  TaskMsg  `json:"msg"`
}

type TaskMsg struct {
	Text           string `json:"text"`
	ChatId         int64  `json:"chatId"`
	ReplyMessageId int    `json:"replyMessageId"`
}

type TaskYtdl struct {
	Link string `json:"link"`
}

func (t *TaskData) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TaskData) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	return nil
}

func (t *Task) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Task) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	return nil
}

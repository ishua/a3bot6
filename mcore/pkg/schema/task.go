package schema

import (
	"encoding/json"
)

type TaskType int

const (
	TaskTypeUndefined = iota
	TaskTypeMsg       //task for Bot, can't create another task
	TaskTypeYtdl      //use in python project
	TaskTypeRest
	TaskTypeNote
	TaskTypeTorrent
)

type TaskStatus int

const (
	TaskStatusUndefined = iota
	TaskStatusCreate    //worker need to do this task
	TaskStatusError     //worker can't complete the task
	TaskStatusSended    //worker recived the task for work
	TaskStatusDone      // worker completed the task
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
	Tr   TaskTr   `json:"tr"`
}

type TaskMsg struct {
	Text           string `json:"text"`
	ChatId         int64  `json:"chatId"`
	ReplyMessageId int    `json:"replyMessageId"`
}

type TaskYtdl struct {
	Link     string `json:"link"`
	UserName string `json:"userName"`
}

type TaskTr struct {
	Command    string `json:"command"`
	TorrentUrl string `json:"torrentUrl"`
	FolderPath string `json:"folderPath"`
	TorrentId  int    `json:"torrentId"`
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

package schema

import (
	"encoding/json"
	"fmt"
)

type MessageType int

const (
	MessageTypeUndefined = iota
	MessageTypeUser
	MessageTypeBot
)

type DialogStatus int

const (
	DialogStatusUndefined = iota
	DialogStatusError
	DialogStatusBegin
	DialogStatusClose = 100
)

type Dialog struct {
	Id           int64        `json:"id"`
	Key          string       `json:"key"`
	DialogStatus DialogStatus `json:"dialogStatus"`
	Messages     []Message    `json:"messages"`
}

type Message struct {
	UserName         string      `json:"userName"`
	MessageId        int         `json:"messageId"`
	ReplyToMessageID int         `json:"replyToMessageID"`
	ChatId           int64       `json:"chatId"`
	Text             string      `json:"text"`
	Caption          string      `json:"caption"`
	FileUrl          string      `json:"fileUrl"`
	Type             MessageType `json:"type"`
}

func (d *Dialog) GetMessagesAsByte() ([]byte, error) {
	return json.Marshal(d.Messages)
}

func (d *Dialog) SetMessagesFromByte(b []byte) error {
	err := json.Unmarshal(b, &d.Messages)
	if err != nil {
		return err
	}
	return nil
}

func GenerateKey(m Message) string {
	return fmt.Sprintf("%d-%s", m.ChatId, m.UserName)
}

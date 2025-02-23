package routing

import (
	"errors"
	"github.com/ishua/a3bot6/mcore/pkg/schema"

	"testing"
)

// Моковая реализация repo
type mockRepo struct {
	AddDialogFunc func(dialog schema.Dialog) (int64, error)
	AddTaskFunc   func(task schema.Task) (int64, error)
}

func (m *mockRepo) AddDialog(dialog schema.Dialog) (int64, error) {
	return m.AddDialogFunc(dialog)
}

func (m *mockRepo) AddTask(task schema.Task) (int64, error) {
	return m.AddTaskFunc(task)
}

// Тест для неавторизованного пользователя
func TestRouter_Build_UnauthorizedUser(t *testing.T) {
	mock := &mockRepo{}
	router := NewRouter([]string{"authorizedUser"}, mock)

	msg := schema.Message{
		UserName:  "unauthorizedUser",
		ChatId:    1,
		MessageId: 1,
		Text:      "/y2d https://youtube.com/video",
	}

	reply := router.Build(msg)

	expected := "I don't answer to user unauthorizedUser"
	if reply.Text != expected {
		t.Errorf("expected '%s', got '%s'", expected, reply.Text)
	}
}

// Тест для некорректной команды
func TestRouter_Build_InvalidCommand(t *testing.T) {
	mock := &mockRepo{}
	router := NewRouter([]string{"authorizedUser"}, mock)

	msg := schema.Message{
		UserName:  "authorizedUser",
		ChatId:    1,
		MessageId: 1,
		Text:      "/invalidCommand",
	}

	reply := router.Build(msg)

	expected := "command not found"
	if reply.Text != expected {
		t.Errorf("expected '%s', got '%s'", expected, reply.Text)
	}
}

// Тест для успешного выполнения команды
func TestRouter_Build_Success(t *testing.T) {
	mock := &mockRepo{
		AddDialogFunc: func(dialog schema.Dialog) (int64, error) {
			return 123, nil
		},
		AddTaskFunc: func(task schema.Task) (int64, error) {
			return 456, nil
		},
	}
	router := NewRouter([]string{"authorizedUser"}, mock)

	msg := schema.Message{
		UserName:  "authorizedUser",
		ChatId:    1,
		MessageId: 1,
		Text:      "/y2d https://youtube.com/watch?v=dQw4w9WgXcQ",
	}

	reply := router.Build(msg)

	if reply.Text != "OK" {
		t.Errorf("expected 'OK', got '%s'", reply.Text)
	}
	if reply.ChatId != 123 {
		t.Errorf("expected %d, got %d", 123, reply.ChatId)
	}
}

// Тест ошибки при добавлении диалога
func TestRouter_Build_AddDialogError(t *testing.T) {
	mock := &mockRepo{
		AddDialogFunc: func(dialog schema.Dialog) (int64, error) {
			return 0, errors.New("failed to add dialog")
		},
	}
	router := NewRouter([]string{"authorizedUser"}, mock)

	msg := schema.Message{
		UserName:  "authorizedUser",
		ChatId:    1,
		MessageId: 1,
		Text:      "/y2d https://youtube.com/watch?v=dQw4w9WgXcQ",
	}

	reply := router.Build(msg)

	expected := "failed to add dialog"
	if reply.Text != expected {
		t.Errorf("expected '%s', got '%s'", expected, reply.Text)
	}
}

// Тест ошибки при добавлении задачи
func TestRouter_Build_AddTaskError(t *testing.T) {
	mock := &mockRepo{
		AddDialogFunc: func(dialog schema.Dialog) (int64, error) {
			return 123, nil
		},
		AddTaskFunc: func(task schema.Task) (int64, error) {
			return 0, errors.New("failed to add task")
		},
	}
	router := NewRouter([]string{"authorizedUser"}, mock)

	msg := schema.Message{
		UserName:  "authorizedUser",
		ChatId:    1,
		MessageId: 1,
		Text:      "/y2d https://youtube.com/watch?v=dQw4w9WgXcQ",
	}

	reply := router.Build(msg)

	expected := "failed to add task"
	if reply.Text != expected {
		t.Errorf("expected '%s', got '%s'", expected, reply.Text)
	}
}

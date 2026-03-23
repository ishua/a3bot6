package worker

import (
	"errors"
	"testing"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"github.com/ishua/a3bot6/synoc/internal/synology"
)

type mockSynoClient struct {
	tasks     []mockTask
	nextID    int
	createErr error
	listErr   error
	deleteErr error
}

type mockTask struct {
	ID    string
	Title string
}

func (m *mockSynoClient) CreateTask(uri, destination, filename string) (string, error) {
	if m.createErr != nil {
		return "", m.createErr
	}
	m.nextID++
	id := "dbid_1"
	if m.nextID > 1 {
		id = "dbid_" + string(rune('0'+m.nextID))
	}
	m.tasks = append(m.tasks, mockTask{ID: id, Title: uri})
	return id, nil
}

func (m *mockSynoClient) ListTasks() ([]synology.Task, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]synology.Task, len(m.tasks))
	for i, t := range m.tasks {
		result[i] = synology.Task{ID: t.ID, Title: t.Title, Status: "waiting"}
	}
	return result, nil
}

func (m *mockSynoClient) DeleteTask(taskID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, t := range m.tasks {
		if t.ID == taskID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return errors.New("task not found")
}

func TestHandleAdd(t *testing.T) {
	client := &mockSynoClient{}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command:    schema.SynoTaskCmdAdd,
				Category:   schema.SynoCategoryOther,
				TorrentUrl: "https://example.com/file.torrent",
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusDone {
		t.Errorf("expected status %v, got %v", schema.TaskStatusDone, report.Status)
	}

	if len(client.tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(client.tasks))
	}
}

func TestHandleAddWithError(t *testing.T) {
	client := &mockSynoClient{
		createErr: errors.New("create failed"),
	}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command:    schema.SynoTaskCmdAdd,
				Category:   schema.SynoCategoryOther,
				TorrentUrl: "https://example.com/file.torrent",
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusError {
		t.Errorf("expected status %v, got %v", schema.TaskStatusError, report.Status)
	}
}

func TestHandleList(t *testing.T) {
	client := &mockSynoClient{
		tasks: []mockTask{
			{ID: "dbid_1", Title: "file1.torrent"},
			{ID: "dbid_2", Title: "file2.torrent"},
		},
	}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdList,
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusDone {
		t.Errorf("expected status %v, got %v", schema.TaskStatusDone, report.Status)
	}
}

func TestHandleListWithError(t *testing.T) {
	client := &mockSynoClient{
		listErr: errors.New("list failed"),
	}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdList,
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusError {
		t.Errorf("expected status %v, got %v", schema.TaskStatusError, report.Status)
	}
}

func TestHandleDelete(t *testing.T) {
	client := &mockSynoClient{
		tasks: []mockTask{
			{ID: "dbid_1", Title: "file1.torrent"},
		},
	}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdDelete,
				TaskId:  "dbid_1",
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusDone {
		t.Errorf("expected status %v, got %v", schema.TaskStatusDone, report.Status)
	}

	if len(client.tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(client.tasks))
	}
}

func TestHandleDeleteWithoutTaskId(t *testing.T) {
	client := &mockSynoClient{}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdDelete,
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusError {
		t.Errorf("expected status %v, got %v", schema.TaskStatusError, report.Status)
	}
}

func TestHandleUnknownCommand(t *testing.T) {
	client := &mockSynoClient{}
	paths := map[string]string{"other": "downloads/other"}
	w := New(client, paths)

	task := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: "unknown",
			},
		},
	}

	report := w.DoTask(task)

	if report.Status != schema.TaskStatusError {
		t.Errorf("expected status %v, got %v", schema.TaskStatusError, report.Status)
	}
}

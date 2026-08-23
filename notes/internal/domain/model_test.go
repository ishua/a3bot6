package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

// fakeGit is a minimal gitClient stub used to point readFile at a temp dir.
type fakeGit struct {
	repoPath string
}

func (f fakeGit) Pull() error                       { return nil }
func (f fakeGit) GetRepoPath() string               { return f.repoPath }
func (f fakeGit) CommitAndPush(path []string) error { return nil }

// newTaskListModel builds a Model whose taskFilePath resolves inside a temp
// dir. When fileContent is empty the task file is not created, which exercises
// the missing-file error path.
func newTaskListModel(t *testing.T, fileContent, defaultBehavior string) *Model {
	t.Helper()
	dir := t.TempDir()
	if fileContent != "" {
		if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(fileContent), 0o600); err != nil {
			t.Fatalf("write tasks.md: %v", err)
		}
	}
	return &Model{
		gitClient:           fakeGit{repoPath: dir},
		taskFilePath:        "tasks.md",
		taskStopMarker:      "### Long",
		taskDefaultBehavior: defaultBehavior,
	}
}

// mkTaskList builds a Task carrying the taskList command with the given AddText.
func mkTaskList(id int64, addText string) schema.Task {
	return schema.Task{
		Id: id,
		TaskData: schema.TaskData{
			Tn: schema.TaskNote{Command: schema.TaskNoteCmdTaskList, AddText: addText},
		},
	}
}

const taskListBaseFile = `### Today
- wash car
- buy food

### summer 2030
- plan trip due:2030-08-09
- oil change due:2000-01-01
`

// TestTaskList_Dispatch checks each AddText → filter mapping and the default
// behavior applied when AddText is empty.
func TestTaskList_Dispatch(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		addText   string
		behav     string
		wantParts []string
	}{
		{
			name:      "list returns all tasks",
			file:      taskListBaseFile,
			addText:   "list",
			behav:     "today",
			wantParts: []string{"wash car", "buy food", "plan trip due:2030-08-09", "oil change due:2000-01-01"},
		},
		{
			name:      "due returns only tasks with due date",
			file:      taskListBaseFile,
			addText:   "due",
			behav:     "today",
			wantParts: []string{"plan trip due:2030-08-09", "oil change due:2000-01-01"},
		},
		{
			name:    "today returns Today section plus overdue",
			file:    taskListBaseFile,
			addText: "today",
			behav:   "today",
			// plan trip (2030) is in the future and must be excluded.
			wantParts: []string{"wash car", "buy food", "oil change due:2000-01-01"},
		},
		{
			name:      "due today returns only today or overdue",
			file:      taskListBaseFile,
			addText:   "due today",
			behav:     "today",
			wantParts: []string{"oil change due:2000-01-01"},
		},
		{
			name: "warn returns warnings only",
			file: `### Today
- ok
- bad due:notadate
`,
			addText:   "warn",
			behav:     "today",
			wantParts: []string{"invalid due date"},
		},
		{
			name:      "empty AddText uses default behavior",
			file:      taskListBaseFile,
			addText:   "",
			behav:     "list",
			wantParts: []string{"plan trip due:2030-08-09"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTaskListModel(t, tt.file, tt.behav)
			got := m.taskList(mkTaskList(7, tt.addText))

			if got.Status != schema.TaskStatusDone {
				t.Fatalf("taskList() Status = %v, want %v (TextMsg=%q)", got.Status, schema.TaskStatusDone, got.TextMsg)
			}
			for _, want := range tt.wantParts {
				if !strings.Contains(got.TextMsg, want) {
					t.Errorf("taskList() TextMsg %q does not contain %q", got.TextMsg, want)
				}
			}
		})
	}
}

// TestTaskList_UnknownFilter verifies the default-case error, including that a
// misconfigured taskDefaultBehavior is surfaced (resolved behavior, not the
// empty AddText).
func TestTaskList_UnknownFilter(t *testing.T) {
	tests := []struct {
		name    string
		addText string
		behav   string
		wantMsg string
	}{
		{name: "explicit unknown argument", addText: "nope", behav: "today", wantMsg: "unknown task filter: nope"},
		{name: "misconfigured default behavior", addText: "", behav: "bogus", wantMsg: "unknown task filter: bogus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTaskListModel(t, taskListBaseFile, tt.behav)
			got := m.taskList(mkTaskList(7, tt.addText))

			if got.Status != schema.TaskStatusError {
				t.Fatalf("taskList() Status = %v, want %v", got.Status, schema.TaskStatusError)
			}
			if got.TextMsg != tt.wantMsg {
				t.Errorf("taskList() TextMsg = %q, want %q", got.TextMsg, tt.wantMsg)
			}
		})
	}
}

// TestTaskList_MissingFile verifies the read-error path returns TaskStatusError
// when the task file does not exist.
func TestTaskList_MissingFile(t *testing.T) {
	m := newTaskListModel(t, "", "today")
	got := m.taskList(mkTaskList(7, "list"))

	if got.Status != schema.TaskStatusError {
		t.Fatalf("taskList() Status = %v, want %v", got.Status, schema.TaskStatusError)
	}
	if !strings.Contains(got.TextMsg, "read task file") {
		t.Errorf("taskList() TextMsg = %q, want it to mention read task file", got.TextMsg)
	}
}

// TestTaskList_EmptyResult verifies the distinct empty messages for tasks and
// warnings.
func TestTaskList_EmptyResult(t *testing.T) {
	noDueTasks := `### summer 2030
- plan trip due:2030-08-09
`
	clean := `### Today
- ok
`

	m := newTaskListModel(t, noDueTasks, "today")
	got := m.taskList(mkTaskList(7, "due today"))
	if got.Status != schema.TaskStatusDone || got.TextMsg != "no tasks found" {
		t.Fatalf("taskList() = Status %v, TextMsg %q; want Done, %q", got.Status, got.TextMsg, "no tasks found")
	}

	m = newTaskListModel(t, clean, "today")
	got = m.taskList(mkTaskList(7, "warn"))
	if got.Status != schema.TaskStatusDone || got.TextMsg != "no warnings found" {
		t.Fatalf("taskList() = Status %v, TextMsg %q; want Done, %q", got.Status, got.TextMsg, "no warnings found")
	}
}

package taskmng

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

type stubRepo struct {
	added      schema.Task
	addTaskErr error
}

func (s *stubRepo) AddTask(task schema.Task) (int64, error) {
	s.added = task
	if s.addTaskErr != nil {
		return 0, s.addTaskErr
	}
	return 1, nil
}

func (s *stubRepo) GetFirstTaskByType(t schema.TaskType) (schema.Task, error) {
	return schema.Task{}, nil
}

func (s *stubRepo) GetTaskById(id int64) (schema.Task, error) {
	return schema.Task{}, nil
}

func (s *stubRepo) GetDialogById(id int64) (schema.Dialog, error) {
	return schema.Dialog{}, nil
}

func (s *stubRepo) UpdateTaskStatus(task schema.Task) error {
	return nil
}

func (s *stubRepo) UpdateDialog(d schema.Dialog) error {
	return nil
}

func TestParseSimplifications_nt(t *testing.T) {
	got, err := parseSimplifications([]string{"nt", "list"})
	if err != nil {
		t.Fatalf("parseSimplifications(nt list) unexpected error: %v", err)
	}
	want := []string{"/note", "task", "list"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseSimplifications(nt list) = %v, want %v", got, want)
	}

	got, err = parseSimplifications([]string{"Nt", "due", "today"})
	if err != nil {
		t.Fatalf("parseSimplifications(Nt due today) unexpected error: %v", err)
	}
	want = []string{"/note", "task", "due", "today"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseSimplifications(Nt due today) = %v, want %v", got, want)
	}
}

func TestParseNoteTaskArgs(t *testing.T) {
	tn, err := parseNoteTaskArgs([]string{"/note", "task", "list"})
	if err != nil {
		t.Fatalf("parseNoteTaskArgs(task list) unexpected error: %v", err)
	}
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "list" {
		t.Errorf("parseNoteTaskArgs(task list) = %+v, want taskList/list", tn)
	}

	tn, err = parseNoteTaskArgs([]string{"/note", "tasks", "due", "today"})
	if err != nil {
		t.Fatalf("parseNoteTaskArgs(tasks due today) unexpected error: %v", err)
	}
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "due today" {
		t.Errorf("parseNoteTaskArgs(tasks due today) = %+v, want taskList/due today", tn)
	}

	tn, err = parseNoteTaskArgs([]string{"/note", "task"})
	if err != nil {
		t.Fatalf("parseNoteTaskArgs(task) unexpected error: %v", err)
	}
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "" {
		t.Errorf("parseNoteTaskArgs(task) = %+v, want taskList/empty", tn)
	}

	tn, err = parseNoteTaskArgs([]string{"/note", "task", "warn"})
	if err != nil {
		t.Fatalf("parseNoteTaskArgs(task warn) unexpected error: %v", err)
	}
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "warn" {
		t.Errorf("parseNoteTaskArgs(task warn) = %+v, want taskList/warn", tn)
	}

	// Unknown filter passthrough.
	tn, err = parseNoteTaskArgs([]string{"/note", "task", "foo"})
	if err != nil {
		t.Fatalf("parseNoteTaskArgs(task foo) unexpected error: %v", err)
	}
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "foo" {
		t.Errorf("parseNoteTaskArgs(task foo) = %+v, want taskList/foo", tn)
	}

	// Whitespace normalization: extra spaces collapse into a clean filter value.
	tn, err = parseNoteTaskArgs([]string{"/note", "task", "due", "", "today", ""})
	if err != nil {
		t.Fatalf("parseNoteTaskArgs(task due  today ) unexpected error: %v", err)
	}
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "due today" {
		t.Errorf("parseNoteTaskArgs(task due  today ) = %+v, want taskList/due today", tn)
	}
}

func TestParseNoteTaskArgs_TooFewWords(t *testing.T) {
	// The pure helper must guard words[1] instead of panicking on short input.
	for _, words := range [][]string{{}, {"/note"}} {
		tn, err := parseNoteTaskArgs(words)
		if err == nil {
			t.Errorf("parseNoteTaskArgs(%v) expected error, got nil", words)
		}
		if tn.Command != "" {
			t.Errorf("parseNoteTaskArgs(%v) command = %q, want empty", words, tn.Command)
		}
	}
}

func TestParseNoteTaskArgs_ExistingCommands(t *testing.T) {
	cases := []struct {
		words   []string
		command schema.TaskNoteCmd
		addText string
	}{
		{[]string{"/note", "5bx", "line"}, schema.TaskNoteCmdAdd5bx, "line"},
		{[]string{"/note", "entry", "msg"}, schema.TaskNoteCmdAddEntry, "msg"},
		{[]string{"/note", "pull"}, schema.TaskNoteCmdPull, ""},
		{[]string{"/note", "weight", "78"}, schema.TaskNoteCmdAddWeight, "78"},
		{[]string{"/note", "bp", "120/80"}, schema.TaskNoteCmdAddBP, "120/80"},
		{[]string{"/note", "inbox", "add", "note1"}, schema.TaskNoteCmdAddInbox, "note1"},
		{[]string{"/note", "inbox", "read"}, schema.TaskNoteReadInbox, ""},
	}
	for _, c := range cases {
		tn, err := parseNoteTaskArgs(c.words)
		if err != nil {
			t.Fatalf("parseNoteTaskArgs(%v) unexpected error: %v", c.words, err)
		}
		if tn.Command != c.command || tn.AddText != c.addText {
			t.Errorf("parseNoteTaskArgs(%v) = {cmd:%s, text:%q}, want {cmd:%s, text:%q}",
				c.words, tn.Command, tn.AddText, c.command, c.addText)
		}
	}

	// Bare /note inbox must return the label-guard error instead of panicking.
	tn, err := parseNoteTaskArgs([]string{"/note", "inbox"})
	if err == nil || err.Error() != "for inbox need label" {
		t.Errorf("parseNoteTaskArgs(/note inbox) err = %v, want \"for inbox need label\"", err)
	}
	if tn.Command != "" {
		t.Errorf("parseNoteTaskArgs(/note inbox) command = %q, want empty", tn.Command)
	}
}

func TestCreateNoteTask_BareNt(t *testing.T) {
	for _, alias := range []string{"nt", "Nt"} {
		repo := &stubRepo{}
		m := NewTaskMng(repo)
		res, err := m.createNoteTask(123, alias)
		if err != nil {
			t.Fatalf("createNoteTask(%q) unexpected error: %v", alias, err)
		}
		if res != "task note created" {
			t.Errorf("createNoteTask(%q) result = %q, want \"task note created\"", alias, res)
		}
		tn := repo.added.TaskData.Tn
		if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "" {
			t.Errorf("createNoteTask(%q) task = %+v, want taskList/empty", alias, tn)
		}
	}

	// Other bare aliases must still be rejected for lack of a command.
	for _, alias := range []string{"n5", "nd", "nw", "nbp", "ni"} {
		m := NewTaskMng(&stubRepo{})
		_, err := m.createNoteTask(123, alias)
		if err == nil {
			t.Errorf("createNoteTask(%q) expected error, got nil", alias)
		} else if err.Error() != "for note need command" {
			t.Errorf("createNoteTask(%q) err = %q, want \"for note need command\"", alias, err)
		}
	}
}

func TestCreateNoteTask_BareNote(t *testing.T) {
	// Bare /note without a subcommand must be rejected, not create a task.
	repo := &stubRepo{}
	m := NewTaskMng(repo)
	_, err := m.createNoteTask(123, "/note")
	if err == nil {
		t.Fatalf("createNoteTask(/note) expected error, got nil")
	}
	if err.Error() != "for note need command" {
		t.Errorf("createNoteTask(/note) err = %q, want \"for note need command\"", err)
	}
	if repo.added.TaskData.Tn.Command != "" {
		t.Errorf("createNoteTask(/note) created a task, got %+v", repo.added)
	}
}

func TestCreateNoteTask_CaseSensitiveTask(t *testing.T) {
	// /note Task is not recognized as taskList (case-sensitive).
	repo := &stubRepo{}
	m := NewTaskMng(repo)
	_, err := m.createNoteTask(123, "/note Task")
	if err == nil {
		t.Fatalf("createNoteTask(/note Task) expected error, got nil")
	}
	if repo.added.TaskData.Tn.Command != "" {
		t.Errorf("createNoteTask(/note Task) created a task, got %+v", repo.added)
	}
}

func TestCreateNoteTask_AddTaskError(t *testing.T) {
	// Repo failure must propagate the wrapped error.
	repo := &stubRepo{addTaskErr: fmt.Errorf("boom")}
	m := NewTaskMng(repo)
	_, err := m.createNoteTask(123, "/note task list")
	if err == nil {
		t.Fatalf("createNoteTask(/note task list) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "taskMng add task") {
		t.Errorf("createNoteTask error = %q, want it to contain \"taskMng add task\"", err)
	}
	tn := repo.added.TaskData.Tn
	if tn.Command != schema.TaskNoteCmdTaskList || tn.AddText != "list" {
		t.Errorf("createNoteTask task = %+v, want taskList/list", repo.added)
	}
}

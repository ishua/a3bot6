package taskparser

import (
	"testing"
	"time"
)

// mkItem is a test helper that builds a TaskItem with the given text and,
// optionally, nested subtasks.
func mkItem(text string, subtasks ...TaskItem) TaskItem {
	return TaskItem{Text: text, Subtasks: subtasks}
}

// withDue returns a copy of it with the Due field set to midnight UTC on the
// given date. Used to set a due date without hand-building a time.Time.
func (it TaskItem) withDue(year int, month time.Month, day int) TaskItem {
	it.Due = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return it
}

func TestFilterAll(t *testing.T) {
	tests := []struct {
		name     string
		tf       *TaskFile
		expected []string
	}{
		{
			name:     "nil taskfile",
			tf:       nil,
			expected: []string{},
		},
		{
			name:     "no sections",
			tf:       &TaskFile{},
			expected: []string{},
		},
		{
			name:     "empty section",
			tf:       &TaskFile{Sections: []Section{{Title: "Today", Tasks: []TaskItem{}}}},
			expected: []string{},
		},
		{
			name: "single section",
			tf: &TaskFile{Sections: []Section{{
				Title: "Today",
				Tasks: []TaskItem{mkItem("a"), mkItem("b"), mkItem("c")},
			}}},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "multiple sections order preserved",
			tf: &TaskFile{Sections: []Section{
				{Title: "Today", Tasks: []TaskItem{mkItem("wash car"), mkItem("buy food")}},
				{Title: "summer 2030", Tasks: []TaskItem{mkItem("plan trip"), mkItem("oil change")}},
				{Title: "fall 2030", Tasks: []TaskItem{mkItem("renew passport")}},
			}},
			expected: []string{"wash car", "buy food", "plan trip", "oil change", "renew passport"},
		},
		{
			name: "subtasks stay nested in parent",
			tf: &TaskFile{Sections: []Section{{
				Title: "Today",
				Tasks: []TaskItem{
					mkItem("parent", mkItem("child1"), mkItem("child2")),
					mkItem("other"),
				},
			}}},
			expected: []string{"parent", "other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterAll(tt.tf)
			if got == nil {
				t.Fatalf("FilterAll() returned nil (want non-nil empty slice)")
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("FilterAll() returned %d tasks, want %d: %v", len(got), len(tt.expected), got)
			}
			for i, want := range tt.expected {
				if got[i].Text != want {
					t.Errorf("FilterAll()[%d].Text = %q, want %q", i, got[i].Text, want)
				}
			}
		})
	}
}

// TestFilterAll_SubtasksNested verifies that returned top-level tasks keep
// their original subtask structure untouched.
func TestFilterAll_SubtasksNested(t *testing.T) {
	child := mkItem("child", mkItem("grandchild"))
	parent := mkItem("parent", child)
	tf := &TaskFile{Sections: []Section{{Title: "Today", Tasks: []TaskItem{parent}}}}

	got := FilterAll(tf)
	if len(got) != 1 {
		t.Fatalf("FilterAll() returned %d tasks, want 1", len(got))
	}
	if len(got[0].Subtasks) != 1 {
		t.Fatalf("FilterAll() flattened subtasks for %q, want 1 nested subtask", got[0].Text)
	}
	if got[0].Subtasks[0].Text != "child" {
		t.Errorf("subtask text = %q, want %q", got[0].Subtasks[0].Text, "child")
	}
	if len(got[0].Subtasks[0].Subtasks) != 1 {
		t.Fatalf("grandchild lost: got %d nested levels, want 1", len(got[0].Subtasks[0].Subtasks))
	}
}

// TestFilterDue verifies that FilterDue returns only top-level tasks carrying
// a non-zero Due, sorted ascending by due date, and never includes no-due or
// subtask-only entries.
func TestFilterDue(t *testing.T) {
	tests := []struct {
		name     string
		tf       *TaskFile
		expected []string
	}{
		{
			name:     "nil taskfile",
			tf:       nil,
			expected: []string{},
		},
		{
			name:     "no sections",
			tf:       &TaskFile{},
			expected: []string{},
		},
		{
			name: "no due dates",
			tf: &TaskFile{Sections: []Section{{
				Title: "Today",
				Tasks: []TaskItem{mkItem("a"), mkItem("b")},
			}}},
			expected: []string{},
		},
		{
			name: "single section sorted ascending",
			tf: &TaskFile{Sections: []Section{{
				Title: "summer 2030",
				Tasks: []TaskItem{
					mkItem("later").withDue(2030, 8, 9),
					mkItem("sooner").withDue(2030, 7, 6),
					mkItem("middle").withDue(2030, 7, 13),
				},
			}}},
			expected: []string{"sooner", "middle", "later"},
		},
		{
			name: "across sections grouped then sorted",
			tf: &TaskFile{Sections: []Section{
				{Title: "fall 2030", Tasks: []TaskItem{mkItem("fall").withDue(2030, 10, 13)}},
				{Title: "summer 2030", Tasks: []TaskItem{mkItem("summer").withDue(2030, 7, 6)}},
			}},
			expected: []string{"summer", "fall"},
		},
		{
			name: "no-due tasks excluded",
			tf: &TaskFile{Sections: []Section{{
				Title: "Today",
				Tasks: []TaskItem{
					mkItem("no due"),
					mkItem("due task").withDue(2030, 7, 6),
					mkItem("other no due"),
				},
			}}},
			expected: []string{"due task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterDue(tt.tf)
			if got == nil {
				t.Fatalf("FilterDue() returned nil (want non-nil empty slice)")
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("FilterDue() returned %d tasks, want %d: %v", len(got), len(tt.expected), got)
			}
			for i, want := range tt.expected {
				if got[i].Text != want {
					t.Errorf("FilterDue()[%d].Text = %q, want %q", i, got[i].Text, want)
				}
			}
		})
	}
}

// TestFilterDue_Stable verifies that tasks sharing the same due date keep
// their original file order after sorting.
func TestFilterDue_Stable(t *testing.T) {
	tf := &TaskFile{Sections: []Section{{
		Title: "summer 2030",
		Tasks: []TaskItem{
			mkItem("first").withDue(2030, 7, 6),
			mkItem("second").withDue(2030, 7, 6),
			mkItem("third").withDue(2030, 7, 6),
		},
	}}}

	got := FilterDue(tf)
	if len(got) != 3 {
		t.Fatalf("FilterDue() returned %d tasks, want 3", len(got))
	}
	if got[0].Text != "first" || got[1].Text != "second" || got[2].Text != "third" {
		t.Errorf("FilterDue() not stable: got %q, %q, %q; want first, second, third", got[0].Text, got[1].Text, got[2].Text)
	}
}

// TestFilterNoDueExcluded verifies that tasks without a due date are never
// included in FilterDue. (FilterDueToday coverage is added in Task 4.)
func TestFilterNoDueExcluded(t *testing.T) {
	withNoDue := mkItem("chore")
	withDue := mkItem("appointment").withDue(2030, 7, 6)
	tf := &TaskFile{Sections: []Section{{
		Title: "Today",
		Tasks: []TaskItem{withNoDue, withDue, mkItem("errand")},
	}}}

	got := FilterDue(tf)
	if len(got) != 1 {
		t.Fatalf("FilterDue() returned %d tasks, want 1 (only the due one)", len(got))
	}
	if got[0].Text != "appointment" {
		t.Errorf("FilterDue() included no-due task: got %q, want %q", got[0].Text, "appointment")
	}
}

// TestFilterAll_DoesNotSort asserts that FilterAll preserves file order even
// when tasks carry due dates (I/O over sorted data must not reorder).
func TestFilterAll_DoesNotSort(t *testing.T) {
	// Attach due in reverse order: "later" has the earlier due date (2030-07-06)
	// but must stay first because it appears first in the file, while "earlier"
	// has the later due date (2030-07-13).
	tf := &TaskFile{Sections: []Section{{
		Title: "Today",
		Tasks: []TaskItem{
			mkItem("later").withDue(2030, 7, 6),
			mkItem("earlier").withDue(2030, 7, 13),
		},
	}}}

	got := FilterAll(tf)
	if len(got) != 2 {
		t.Fatalf("FilterAll() returned %d tasks, want 2", len(got))
	}
	if got[0].Text != "later" || got[1].Text != "earlier" {
		t.Errorf("FilterAll() reordered tasks: got %q, %q; want %q, %q", got[0].Text, got[1].Text, "later", "earlier")
	}
}

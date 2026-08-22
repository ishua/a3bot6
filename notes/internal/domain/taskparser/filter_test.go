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

// TestFilterDueToday verifies that FilterDueToday returns only top-level
// tasks carrying a non-zero Due whose due date is on or before now (today or
// overdue), sorted ascending by due date, and excludes no-due and future tasks.
func TestFilterDueToday(t *testing.T) {
	now := time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		tf       *TaskFile
		now      time.Time
		expected []string
	}{
		{
			name:     "nil taskfile",
			tf:       nil,
			now:      now,
			expected: []string{},
		},
		{
			name:     "no sections",
			tf:       &TaskFile{},
			now:      now,
			expected: []string{},
		},
		{
			name: "no due dates",
			tf: &TaskFile{Sections: []Section{{
				Title: "Today",
				Tasks: []TaskItem{mkItem("a"), mkItem("b")},
			}}},
			now:      now,
			expected: []string{},
		},
		{
			name: "only today/overdue included, sorted ascending",
			tf: &TaskFile{Sections: []Section{
				{Title: "Today", Tasks: []TaskItem{
					mkItem("overdue later").withDue(2030, 7, 14),
					mkItem("overdue earlier").withDue(2030, 7, 6),
					mkItem("future").withDue(2030, 8, 9),
					mkItem("no due"),
				}},
				{Title: "summer 2030", Tasks: []TaskItem{
					mkItem("other overdue").withDue(2030, 7, 13),
					mkItem("other future").withDue(2031, 1, 15),
				}},
			}},
			now:      now,
			expected: []string{"overdue earlier", "other overdue", "overdue later"},
		},
		{
			name: "today inclusive regardless of time of day",
			tf: &TaskFile{Sections: []Section{{
				Title: "summer 2030",
				Tasks: []TaskItem{mkItem("due today late hour").withDue(2030, 7, 14)},
			}}},
			now:      now,
			expected: []string{"due today late hour"},
		},
		{
			name: "tasks in the future excluded",
			tf: &TaskFile{Sections: []Section{{
				Title: "summer 2030",
				Tasks: []TaskItem{
					mkItem("future soon").withDue(2030, 7, 15),
					mkItem("future far").withDue(2031, 1, 15),
				},
			}}},
			now:      now,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterDueToday(tt.tf, tt.now)
			if got == nil {
				t.Fatalf("FilterDueToday() returned nil (want non-nil empty slice)")
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("FilterDueToday() returned %d tasks, want %d: %v", len(got), len(tt.expected), got)
			}
			for i, want := range tt.expected {
				if got[i].Text != want {
					t.Errorf("FilterDueToday()[%d].Text = %q, want %q", i, got[i].Text, want)
				}
			}
		})
	}
}

// TestFilterToday_Stable verifies that tasks sharing the same due date keep
// their original file order after sorting by FilterDueToday.
func TestFilterDueToday_Stable(t *testing.T) {
	tf := &TaskFile{Sections: []Section{{
		Title: "summer 2030",
		Tasks: []TaskItem{
			mkItem("second").withDue(2030, 7, 6),
			mkItem("first").withDue(2030, 7, 6),
			mkItem("third").withDue(2030, 7, 6),
		},
	}}}
	now := time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC)

	got := FilterDueToday(tf, now)
	if len(got) != 3 {
		t.Fatalf("FilterDueToday() returned %d tasks, want 3", len(got))
	}
	if got[0].Text != "second" || got[1].Text != "first" || got[2].Text != "third" {
		t.Errorf("FilterDueToday() not stable: got %q, %q, %q; want second, first, third", got[0].Text, got[1].Text, got[2].Text)
	}
}

// TestFilterNoDueExcluded verifies that tasks without a due date are never
// included in FilterDue or FilterDueToday.
func TestFilterNoDueExcluded(t *testing.T) {
	withNoDue := mkItem("chore")
	withDue := mkItem("appointment").withDue(2030, 7, 6)
	tf := &TaskFile{Sections: []Section{{
		Title: "Today",
		Tasks: []TaskItem{withNoDue, withDue, mkItem("errand")},
	}}}

	now := time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC)

	gotDue := FilterDue(tf)
	if len(gotDue) != 1 {
		t.Fatalf("FilterDue() returned %d tasks, want 1 (only the due one)", len(gotDue))
	}
	if gotDue[0].Text != "appointment" {
		t.Errorf("FilterDue() included no-due task: got %q, want %q", gotDue[0].Text, "appointment")
	}

	gotDueToday := FilterDueToday(tf, now)
	if len(gotDueToday) != 1 {
		t.Fatalf("FilterDueToday() returned %d tasks, want 1 (only the due one)", len(gotDueToday))
	}
	if gotDueToday[0].Text != "appointment" {
		t.Errorf("FilterDueToday() included no-due task: got %q, want %q", gotDueToday[0].Text, "appointment")
	}
}

// TestFilterToday verifies that FilterToday returns the entire "Today" section
// (regardless of due date) plus overdue/today tasks from other sections
// (non-zero Due and due <= now), preserving file order.
func TestFilterToday(t *testing.T) {
	tests := []struct {
		name     string
		tf       *TaskFile
		now      time.Time
		expected []string
	}{
		{
			name:     "nil taskfile",
			tf:       nil,
			now:      time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC),
			expected: []string{},
		},
		{
			name:     "no sections",
			tf:       &TaskFile{},
			now:      time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC),
			expected: []string{},
		},
		{
			name: "no Today section returns only overdue from others",
			tf: &TaskFile{Sections: []Section{{
				Title: "summer 2030",
				Tasks: []TaskItem{
					mkItem("overdue").withDue(2030, 7, 6),
					mkItem("future").withDue(2030, 8, 9),
					mkItem("no due"),
				},
			}}},
			now:      time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC),
			expected: []string{"overdue"},
		},
		{
			name: "Today section with no-due tasks all included",
			tf: &TaskFile{Sections: []Section{
				{Title: "Today", Tasks: []TaskItem{mkItem("chore a"), mkItem("chore b")}},
				{Title: "summer 2030", Tasks: []TaskItem{mkItem("future").withDue(2030, 8, 9)}},
			}},
			now:      time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC),
			expected: []string{"chore a", "chore b"},
		},
		{
			name: "Today section tasks included regardless of due",
			tf: &TaskFile{Sections: []Section{{
				Title: "Today",
				Tasks: []TaskItem{
					mkItem("with due").withDue(2030, 7, 6),
					mkItem("future due").withDue(2030, 8, 9),
					mkItem("no due"),
				},
			}}},
			now:      time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC),
			expected: []string{"with due", "future due", "no due"},
		},
		{
			name: "other sections included only if due <= now",
			tf: &TaskFile{Sections: []Section{
				{Title: "Today", Tasks: []TaskItem{mkItem("wash car")}},
				{Title: "summer 2030", Tasks: []TaskItem{
					mkItem("overdue same day").withDue(2030, 7, 14),
					mkItem("overdue earlier").withDue(2030, 7, 6),
					mkItem("future").withDue(2030, 8, 9),
					mkItem("no due"),
				}},
			}},
			now:      time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC),
			expected: []string{"wash car", "overdue same day", "overdue earlier"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterToday(tt.tf, tt.now)
			if got == nil {
				t.Fatalf("FilterToday() returned nil (want non-nil empty slice)")
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("FilterToday() returned %d tasks, want %d: %v", len(got), len(tt.expected), got)
			}
			for i, want := range tt.expected {
				if got[i].Text != want {
					t.Errorf("FilterToday()[%d].Text = %q, want %q", i, got[i].Text, want)
				}
			}
		})
	}
}

// TestFilterToday_TodayFirst verifies the design ordering: "Today" section
// tasks come first (in file order), then overdue/today tasks from other
// sections (in file order), regardless of section title sort.
func TestFilterToday_TodayFirst(t *testing.T) {
	tf := &TaskFile{Sections: []Section{
		{Title: "summer 2030", Tasks: []TaskItem{
			mkItem("overdue summer").withDue(2030, 7, 6),
		}},
		{Title: "Today", Tasks: []TaskItem{
			mkItem("wash car"),
			mkItem("buy food").withDue(2030, 7, 6),
		}},
		{Title: "fall 2030", Tasks: []TaskItem{
			mkItem("balance somethin").withDue(2030, 7, 13),
			mkItem("future").withDue(2030, 8, 9),
		}},
	}}
	now := time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC)

	got := FilterToday(tf, now)
	expected := []string{"wash car", "buy food", "overdue summer", "balance somethin"}
	if len(got) != len(expected) {
		t.Fatalf("FilterToday() returned %d tasks, want %d: %v", len(got), len(expected), got)
	}
	for i, want := range expected {
		if got[i].Text != want {
			t.Errorf("FilterToday()[%d].Text = %q, want %q", i, got[i].Text, want)
		}
	}
}

// TestFilterToday_DoesNotSort asserts that FilterToday preserves file order
// rather than sorting by due date.
func TestFilterToday_DoesNotSort(t *testing.T) {
	tf := &TaskFile{Sections: []Section{
		{Title: "Today", Tasks: []TaskItem{mkItem("a")}},
		{Title: "summer 2030", Tasks: []TaskItem{
			mkItem("later in file").withDue(2030, 7, 6),
			mkItem("earlier in file").withDue(2030, 7, 13),
		}},
	}}
	now := time.Date(2030, 7, 14, 10, 0, 0, 0, time.UTC)

	got := FilterToday(tf, now)
	expected := []string{"a", "later in file", "earlier in file"}
	if len(got) != len(expected) {
		t.Fatalf("FilterToday() returned %d tasks, want %d: %v", len(got), len(expected), got)
	}
	for i, want := range expected {
		if got[i].Text != want {
			t.Errorf("FilterToday()[%d].Text = %q, want %q", i, got[i].Text, want)
		}
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

package taskparser

import (
	"strings"
	"testing"
)

func TestFormatTasks_Empty(t *testing.T) {
	got := FormatTasks([]TaskItem{}, 100)
	if len(got) != 0 {
		t.Fatalf("FormatTasks(empty) returned %d parts, want 0: %v", len(got), got)
	}
}

func TestFormatTasks_NilInput(t *testing.T) {
	got := FormatTasks(nil, 100)
	if got == nil {
		t.Fatalf("FormatTasks(nil) returned nil (want non-nil empty slice)")
	}
	if len(got) != 0 {
		t.Fatalf("FormatTasks(nil) returned %d parts, want 0: %v", len(got), got)
	}
}

func TestFormatTasks_SingleTask(t *testing.T) {
	got := FormatTasks([]TaskItem{mkItem("wash car")}, 100)
	want := []string{"wash car"}
	assertParts(t, got, want)
}

func TestFormatTasks_Priority(t *testing.T) {
	got := FormatTasks([]TaskItem{
		{Text: "wash car", Priority: "A"},
		{Text: "buy food", Priority: "B"},
		{Text: "no priority"},
	}, 100)
	want := []string{"(A) wash car\n(B) buy food\nno priority"}
	assertParts(t, got, want)
}

func TestFormatTasks_DueDate(t *testing.T) {
	got := FormatTasks([]TaskItem{mkItem("pay").withDue(2030, 7, 6)}, 100)
	want := []string{"pay due:2030-07-06"}
	assertParts(t, got, want)
}

func TestFormatTasks_NoDueDate(t *testing.T) {
	got := FormatTasks([]TaskItem{mkItem("chore")}, 100)
	if len(got) != 1 {
		t.Fatalf("FormatTasks returned %d parts, want 1: %v", len(got), got)
	}
	if strings.Contains(got[0], "due:") {
		t.Errorf("FormatTasks emitted a due token for a task without a due date: %q", got[0])
	}
	if got[0] != "chore" {
		t.Errorf("FormatTasks = %q, want %q", got[0], "chore")
	}
}

func TestFormatTasks_Emojis(t *testing.T) {
	got := FormatTasks([]TaskItem{
		{Text: "clean kitchen", Emojis: []string{"🏡"}},
		{Text: "pay rent", Emojis: []string{"💰", "🏢"}},
	}, 100)
	want := []string{"clean kitchen 🏡\npay rent 💰🏢"}
	assertParts(t, got, want)
}

func TestFormatTasks_Wikilinks(t *testing.T) {
	got := FormatTasks([]TaskItem{
		{Text: "wash car", Wikilinks: []string{"link car"}},
		{Text: "deploy", Wikilinks: []string{"servers", "config"}},
	}, 100)
	want := []string{"wash car [[link car]]\ndeploy [[servers]] [[config]]"}
	assertParts(t, got, want)
}

func TestFormatTasks_InlineLinks(t *testing.T) {
	got := FormatTasks([]TaskItem{
		{Text: "read docs", InlineLinks: []InlineLink{{Text: "docs", URL: "https://example.com/docs"}}},
	}, 100)
	want := []string{"read docs [docs](https://example.com/docs)"}
	assertParts(t, got, want)
}

func TestFormatTasks_EmptyTextWithAttributes(t *testing.T) {
	got := FormatTasks([]TaskItem{
		{Text: "", Wikilinks: []string{"link"}, Emojis: []string{"🏡"}},
	}, 100)
	want := []string{"[[link]] 🏡"}
	assertParts(t, got, want)
}

func TestFormatTasks_Subtasks(t *testing.T) {
	got := FormatTasks([]TaskItem{
		mkItem("parent", mkItem("child1"), mkItem("child2")),
	}, 100)
	want := []string{"parent\n\tchild1\n\tchild2"}
	assertParts(t, got, want)
}

func TestFormatTasks_NestedSubtasks(t *testing.T) {
	got := FormatTasks([]TaskItem{
		mkItem("level0", mkItem("level1", mkItem("level2"))),
	}, 100)
	want := []string{"level0\n\tlevel1\n\t\tlevel2"}
	assertParts(t, got, want)
}

func TestFormatTasks_AllFields(t *testing.T) {
	item := (TaskItem{
		Priority:  "A",
		Text:      "prepare report",
		Wikilinks: []string{"meeting"},
		InlineLinks: []InlineLink{
			{Text: "docs", URL: "https://example.com/report"},
		},
		Emojis: []string{"💰"},
		Subtasks: []TaskItem{
			mkItem("draft"),
			mkItem("review"),
		},
	}).withDue(2030, 7, 6)

	got := FormatTasks([]TaskItem{item}, 100)
	want := []string{"(A) prepare report [[meeting]] [docs](https://example.com/report) due:2030-07-06 💰\n\tdraft\n\treview"}
	assertParts(t, got, want)
}

func TestFormatTasks_LimitSplit(t *testing.T) {
	got := FormatTasks([]TaskItem{mkItem("a"), mkItem("b"), mkItem("cc")}, 4)
	want := []string{"a\nb", "cc"}
	assertParts(t, got, want)
}

func TestFormatTasks_LimitSingleTaskExceeds(t *testing.T) {
	text := "much longer than the limit"
	got := FormatTasks([]TaskItem{mkItem(text)}, 5)
	want := []string{text}
	assertParts(t, got, want)
}

func TestFormatTasks_LimitExact(t *testing.T) {
	// "aaaa\nbbbb" is exactly 9 bytes, equal to the boundary, so it stays in
	// one part.
	got := FormatTasks([]TaskItem{mkItem("aaaa"), mkItem("bbbb")}, 9)
	want := []string{"aaaa\nbbbb"}
	assertParts(t, got, want)
}

func TestFormatTasks_ZeroLimit(t *testing.T) {
	got := FormatTasks([]TaskItem{mkItem("a"), mkItem("b"), mkItem("c")}, 0)
	want := []string{"a\nb\nc"}
	assertParts(t, got, want)
}

func TestFormatTasks_NegativeLimit(t *testing.T) {
	got := FormatTasks([]TaskItem{mkItem("a"), mkItem("b"), mkItem("c")}, -1)
	want := []string{"a\nb\nc"}
	assertParts(t, got, want)
}

func TestFormatWarnings_Empty(t *testing.T) {
	got := FormatWarnings([]Warning{}, 100)
	if len(got) != 0 {
		t.Fatalf("FormatWarnings(empty) returned %d parts, want 0: %v", len(got), got)
	}
}

func TestFormatWarnings_NilInput(t *testing.T) {
	got := FormatWarnings(nil, 100)
	if got == nil {
		t.Fatalf("FormatWarnings(nil) returned nil (want non-nil empty slice)")
	}
	if len(got) != 0 {
		t.Fatalf("FormatWarnings(nil) returned %d parts, want 0: %v", len(got), got)
	}
}

func TestFormatWarnings_Formatted(t *testing.T) {
	got := FormatWarnings([]Warning{
		{Line: 3, Message: "unrecognized line", Text: "garbage here"},
	}, 100)
	want := []string{"line 3: unrecognized line | garbage here"}
	assertParts(t, got, want)
}

func TestFormatWarnings_LimitSplit(t *testing.T) {
	got := FormatWarnings([]Warning{
		{Line: 1, Message: "msgmsg", Text: "aaaa"},
		{Line: 2, Message: "errerr", Text: "bbbb"},
	}, 30)
	want := []string{"line 1: msgmsg | aaaa", "line 2: errerr | bbbb"}
	assertParts(t, got, want)
}

func TestFormatWarnings_LimitSingleWarningExceeds(t *testing.T) {
	w := Warning{Line: 9, Message: "no such thing as a free lunch here", Text: "ignore this line please"}
	got := FormatWarnings([]Warning{w}, 5)
	want := []string{"line 9: no such thing as a free lunch here | ignore this line please"}
	assertParts(t, got, want)
}

func TestFormatTasks_Integration(t *testing.T) {
	tf := loadTestTaskFile(t)
	var items []TaskItem
	for _, s := range tf.Sections {
		items = append(items, s.Tasks...)
	}

	parts := FormatTasks(items, 4096)
	joined := strings.Join(parts, "\n")

	for _, want := range []string{"[[link car]]", "due:2030-07-06", "🏡", "\t"} {
		if !strings.Contains(joined, want) {
			t.Errorf("formatted output is missing %q:\n%s", want, joined)
		}
	}
}

func assertParts(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d parts, want %d:\ngot:  %q\nwant: %q", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("part[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

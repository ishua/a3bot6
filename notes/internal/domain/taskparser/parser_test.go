package taskparser

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("bad time %q: %v", s, err)
	}
	return tm
}

func parseFile(t *testing.T, text string, stopMarker string) *TaskFile {
	t.Helper()
	return ParseTaskFile(strings.Split(text, "\n"), stopMarker)
}

func TestIsTaskLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"dash task", "- foo", true},
		{"star task", "* foo", true},
		{"plain text", "foo", false},
		{"dash no space", "-foo", false},
		{"leading space", " - foo", false},
		{"empty", "", false},
		{"heading", "### Today", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTaskLine(tt.line); got != tt.want {
				t.Errorf("isTaskLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsSubtaskLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"tab dash", "\t- foo", true},
		{"spaces dash", "  - foo", true},
		{"spaces star", "   * foo", true},
		{"task line", "- foo", false},
		{"indented text", "  foo", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSubtaskLine(tt.line); got != tt.want {
				t.Errorf("isSubtaskLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsSectionHeader(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"header", "### Today", true},
		{"double space", "###  winter 2031", true},
		{"h2", "## In Progress", false},
		{"h1", "# Long", false},
		{"no space", "###", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSectionHeader(tt.line); got != tt.want {
				t.Errorf("isSectionHeader(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsStopLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		stopMarker string
		want       bool
	}{
		{"matches marker", "### Long", "### Long", true},
		{"separator", "---", "### Long", true},
		{"separator no marker", "---", "", true},
		{"not stop", "something", "### Long", false},
		{"empty marker no match", "something", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStopLine(tt.line, tt.stopMarker); got != tt.want {
				t.Errorf("isStopLine(%q, %q) = %v, want %v", tt.line, tt.stopMarker, got, tt.want)
			}
		})
	}
}

func TestExtractPriority(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantPri  string
		wantRest string
	}{
		{"A", "(A) do thing", "A", "do thing"},
		{"B", "(B) do thing", "B", "do thing"},
		{"none", "do thing", "", "do thing"},
		{"not marker", "(AB) do thing", "", "(AB) do thing"},
		{"lowercase", "(a) do thing", "", "(a) do thing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPri, gotRest := extractPriority(tt.line)
			if gotPri != tt.wantPri || gotRest != tt.wantRest {
				t.Errorf("extractPriority(%q) = (%q, %q), want (%q, %q)",
					tt.line, gotPri, gotRest, tt.wantPri, tt.wantRest)
			}
		})
	}
}

func TestExtractDueDate(t *testing.T) {
	valid, _ := time.Parse("2006-01-02", "2030-07-13")
	tests := []struct {
		name     string
		text     string
		wantDue  time.Time
		wantRest string
		wantWarn bool
	}{
		{"valid only", "due:2030-07-13", valid, "", false},
		{"valid with text", "foo due:2030-07-13 bar", valid, "foo bar", false},
		{"no due", "plain text", time.Time{}, "plain text", false},
		{"invalid month", "due:2030-13-45", time.Time{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDue, gotRest, gotWarn := extractDueDate(tt.text)
			if !gotDue.Equal(tt.wantDue) {
				t.Errorf("extractDueDate(%q) due = %v, want %v", tt.text, gotDue, tt.wantDue)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractDueDate(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
			if (gotWarn != nil) != tt.wantWarn {
				t.Errorf("extractDueDate(%q) warn = %v, want warn=%v", tt.text, gotWarn, tt.wantWarn)
			}
		})
	}
}

func TestExtractEmojis(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantEmojis []string
		wantRest   string
	}{
		{"single", "clean kitchen 🏡", []string{"🏡"}, "clean kitchen"},
		{"emoji before due", "💰 due:2030-07-13", []string{"💰"}, "due:2030-07-13"},
		{"multiple", "🏡 📞", []string{"🏡", "📞"}, ""},
		{"none", "plain text", []string{}, "plain text"},
		{"duplicate", "🏡 x 🏡", []string{"🏡", "🏡"}, "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEmojis, gotRest := extractEmojis(tt.text)
			if !reflect.DeepEqual(gotEmojis, tt.wantEmojis) {
				t.Errorf("extractEmojis(%q) emojis = %v, want %v", tt.text, gotEmojis, tt.wantEmojis)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractEmojis(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
		})
	}
}

func TestExtractWikilinks(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantLinks []string
		wantRest  string
		wantWarn  bool
	}{
		{"single", "[[finance]] x", []string{"finance"}, "x", false},
		{"multiple", "a [[b]] c [[d]] e", []string{"b", "d"}, "a c e", false},
		{"embedded", "--[[x]]--", []string{"x"}, "----", false},
		{"unclosed", "[[unclosed", []string{}, "", true},
		{"none", "plain text", []string{}, "plain text", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, gotRest, gotWarns := extractWikilinks(tt.text)
			if !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("extractWikilinks(%q) links = %v, want %v", tt.text, gotLinks, tt.wantLinks)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractWikilinks(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
			if (len(gotWarns) > 0) != tt.wantWarn {
				t.Errorf("extractWikilinks(%q) warns = %v, want warn=%v", tt.text, gotWarns, tt.wantWarn)
			}
		})
	}
}

func TestExtractInlineLinks(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		want     []InlineLink
		wantRest string
	}{
		{"single", "[link](url)", []InlineLink{{Text: "link", URL: "url"}}, ""},
		{"with text", "see [here](http://x) ok", []InlineLink{{Text: "here", URL: "http://x"}}, "see ok"},
		{"multiple", "[a](b) [c](d)", []InlineLink{{Text: "a", URL: "b"}, {Text: "c", URL: "d"}}, ""},
		{"none", "plain text", []InlineLink{}, "plain text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, gotRest := extractInlineLinks(tt.text)
			if !reflect.DeepEqual(gotLinks, tt.want) {
				t.Errorf("extractInlineLinks(%q) links = %v, want %v", tt.text, gotLinks, tt.want)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractInlineLinks(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
		})
	}
}

func TestParseFullFile(t *testing.T) {
	data, err := os.ReadFile("testdata/tasklist.md")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	tf := ParseTaskFile(strings.Split(string(data), "\n"), "### Long")

	if len(tf.Sections) != 4 {
		t.Fatalf("got %d sections, want 4: %+v", len(tf.Sections), tf.Sections)
	}
	wantTitles := []string{"Today", "summer 2030", "fall 2030", "winter 2031"}
	for i, want := range wantTitles {
		if tf.Sections[i].Title != want {
			t.Errorf("section %d title = %q, want %q", i, tf.Sections[i].Title, want)
		}
	}
	if len(tf.Warnings) != 4 {
		t.Errorf("got %d warnings, want 4 (H1, bare wikilinks, H2): %+v", len(tf.Warnings), tf.Warnings)
	}

	today := tf.Sections[0].Tasks
	if len(today) < 1 {
		t.Fatalf("Today section has no tasks")
	}
	first := today[0]
	if first.Text != "wash car" {
		t.Errorf("first task text = %q, want %q", first.Text, "wash car")
	}
	if !first.Due.Equal(mustTime(t, "2030-07-06")) {
		t.Errorf("first task due = %v, want 2030-07-06", first.Due)
	}
	if !reflect.DeepEqual(first.Wikilinks, []string{"link car"}) {
		t.Errorf("first task wikilinks = %v, want [link car]", first.Wikilinks)
	}
	if len(first.Subtasks) != 2 {
		t.Fatalf("first task subtasks = %d, want 2: %+v", len(first.Subtasks), first.Subtasks)
	}
	if first.Subtasks[0].Text != "find washer" || first.Subtasks[1].Text != "wash" {
		t.Errorf("first task subtask texts = [%q %q], want [find washer wash]",
			first.Subtasks[0].Text, first.Subtasks[1].Text)
	}

	summer := tf.Sections[1].Tasks
	if summer[0].Text != "balance somethin" {
		t.Errorf("summer first task text = %q, want %q", summer[0].Text, "balance somethin")
	}
	if !reflect.DeepEqual(summer[0].Wikilinks, []string{"finance"}) {
		t.Errorf("summer first wikilinks = %v, want [finance]", summer[0].Wikilinks)
	}
	if !reflect.DeepEqual(summer[0].Emojis, []string{"💰"}) {
		t.Errorf("summer first emojis = %v, want [💰]", summer[0].Emojis)
	}
	if !summer[0].Due.Equal(mustTime(t, "2030-07-13")) {
		t.Errorf("summer first due = %v, want 2030-07-13", summer[0].Due)
	}
}

func TestParseEmptyFile(t *testing.T) {
	tf := ParseTaskFile([]string{}, "")
	if len(tf.Sections) != 0 || len(tf.Warnings) != 0 {
		t.Errorf("empty input: got sections=%d warnings=%d, want 0/0",
			len(tf.Sections), len(tf.Warnings))
	}
	tf2 := ParseTaskFile([]string{"", "", ""}, "stop")
	if len(tf2.Sections) != 0 || len(tf2.Warnings) != 0 {
		t.Errorf("blank-line input: got sections=%d warnings=%d, want 0/0",
			len(tf2.Sections), len(tf2.Warnings))
	}
}

func TestParsePriority(t *testing.T) {
	a, warns := parseTaskLine("- (A) do thing", 1)
	if a.Priority != "A" || a.Text != "do thing" || len(warns) != 0 {
		t.Errorf("(A) => priority=%q text=%q warns=%d, want A/do thing/0", a.Priority, a.Text, len(warns))
	}
	b, _ := parseTaskLine("- (B) other", 2)
	if b.Priority != "B" || b.Text != "other" {
		t.Errorf("(B) => priority=%q text=%q, want B/other", b.Priority, b.Text)
	}
	none, _ := parseTaskLine("- (C) thing", 3)
	if none.Priority != "" || none.Text != "(C) thing" {
		t.Errorf("(C) => priority=%q text=%q, want empty/(C) thing", none.Priority, none.Text)
	}
}

func TestParseDueDate(t *testing.T) {
	valid, warns := parseTaskLine("- do thing due:2030-07-13", 1)
	if !valid.Due.Equal(mustTime(t, "2030-07-13")) || len(warns) != 0 {
		t.Errorf("valid due: got due=%v warns=%d", valid.Due, len(warns))
	}
	missing, warns := parseTaskLine("- do thing", 2)
	if !missing.Due.IsZero() || len(warns) != 0 {
		t.Errorf("missing due: got due=%v warns=%d", missing.Due, len(warns))
	}
	invalid, warns := parseTaskLine("- do thing due:2030-13-45", 3)
	if !invalid.Due.IsZero() {
		t.Errorf("invalid due is not zero: %v", invalid.Due)
	}
	if len(warns) != 1 || warns[0].Line != 3 {
		t.Errorf("invalid due: got warns=%+v, want 1 warning at line 3", warns)
	}
}

func TestParseEmoji(t *testing.T) {
	task, _ := parseTaskLine("- clean kitchen 🏡", 1)
	if !reflect.DeepEqual(task.Emojis, []string{"🏡"}) || task.Text != "clean kitchen" {
		t.Errorf("emoji: got emojis=%v text=%q", task.Emojis, task.Text)
	}
	// Emoji before due date token.
	task2, _ := parseTaskLine("- pay 💰 due:2030-08-09", 2)
	if !reflect.DeepEqual(task2.Emojis, []string{"💰"}) {
		t.Errorf("emoji-before-due: got emojis=%v", task2.Emojis)
	}
	if !task2.Due.Equal(mustTime(t, "2030-08-09")) || task2.Text != "pay" {
		t.Errorf("emoji-before-due: got due=%v text=%q", task2.Due, task2.Text)
	}
}

func TestParseWikilink(t *testing.T) {
	task, warns := parseTaskLine("- [[finance]] do thing", 1)
	if !reflect.DeepEqual(task.Wikilinks, []string{"finance"}) || len(warns) != 0 {
		t.Errorf("wikilink: got %v warns=%d", task.Wikilinks, len(warns))
	}
	task2, _ := parseTaskLine("- do [[a]] and [[b]] thing", 2)
	if !reflect.DeepEqual(task2.Wikilinks, []string{"a", "b"}) {
		t.Errorf("multiple wikilinks: got %v", task2.Wikilinks)
	}
	unclosed, warns := parseTaskLine("- do [[unclosed", 3)
	if unclosed.Text != "do" {
		t.Errorf("unclosed wikilink text = %q, want %q", unclosed.Text, "do")
	}
	if len(warns) != 1 || warns[0].Line != 3 {
		t.Errorf("unclosed wikilink: got warns=%+v, want 1 at line 3", warns)
	}
}

func TestParseInlineLink(t *testing.T) {
	task, warns := parseTaskLine("- see [here](http://x) ok", 1)
	want := []InlineLink{{Text: "here", URL: "http://x"}}
	if !reflect.DeepEqual(task.InlineLinks, want) {
		t.Errorf("inline link: got %+v, want %+v", task.InlineLinks, want)
	}
	if task.Text != "see ok" {
		t.Errorf("inline link text = %q, want %q", task.Text, "see ok")
	}
	if len(warns) != 0 {
		t.Errorf("inline link: unexpected warnings %+v", warns)
	}
}

func TestParseSubtask(t *testing.T) {
	tf := parseFile(t, "### s\n- parent 🏡 due:2030-01-01\n\t- child [[wiki]]\n\t- child2\n\t\t- grand\n", "")
	if len(tf.Sections) != 1 || len(tf.Sections[0].Tasks) != 1 {
		t.Fatalf("unexpected structure: %+v", tf.Sections)
	}
	parent := tf.Sections[0].Tasks[0]
	if parent.Text != "parent" || !reflect.DeepEqual(parent.Emojis, []string{"🏡"}) {
		t.Errorf("parent: text=%q emojis=%v", parent.Text, parent.Emojis)
	}
	if !parent.Due.Equal(mustTime(t, "2030-01-01")) {
		t.Errorf("parent due = %v", parent.Due)
	}
	if len(parent.Subtasks) != 2 {
		t.Fatalf("parent subtasks = %d, want 2: %+v", len(parent.Subtasks), parent.Subtasks)
	}
	child1 := parent.Subtasks[0]
	if child1.Text != "child" || !reflect.DeepEqual(child1.Wikilinks, []string{"wiki"}) {
		t.Errorf("child1: text=%q wikilinks=%v", child1.Text, child1.Wikilinks)
	}
	if len(child1.Subtasks) != 0 {
		t.Errorf("child1 subtasks = %d, want 0", len(child1.Subtasks))
	}
	child2 := parent.Subtasks[1]
	if child2.Text != "child2" || len(child2.Subtasks) != 1 {
		t.Fatalf("child2: text=%q subtasks=%d, want child2/1", child2.Text, len(child2.Subtasks))
	}
	if child2.Subtasks[0].Text != "grand" {
		t.Errorf("grand text = %q, want grand", child2.Subtasks[0].Text)
	}
}

func TestParseSubtaskNested(t *testing.T) {
	tf := parseFile(t, "### s\n- root\n\t- l1a\n\t\t- l2a\n\t\t\t- l3a\n\t- l1b\n", "")
	root := tf.Sections[0].Tasks[0]
	if root.Text != "root" || len(root.Subtasks) != 2 {
		t.Fatalf("root: text=%q subtasks=%d", root.Text, len(root.Subtasks))
	}
	l1a := root.Subtasks[0]
	if l1a.Text != "l1a" || len(l1a.Subtasks) != 1 {
		t.Fatalf("l1a: text=%q subtasks=%d", l1a.Text, len(l1a.Subtasks))
	}
	l2a := l1a.Subtasks[0]
	if l2a.Text != "l2a" || len(l2a.Subtasks) != 1 {
		t.Fatalf("l2a: text=%q subtasks=%d", l2a.Text, len(l2a.Subtasks))
	}
	if l2a.Subtasks[0].Text != "l3a" {
		t.Errorf("l3a text = %q, want l3a", l2a.Subtasks[0].Text)
	}
	if root.Subtasks[1].Text != "l1b" {
		t.Errorf("l1b text = %q, want l1b", root.Subtasks[1].Text)
	}
}

func TestParseStopMarker(t *testing.T) {
	// Stop on the configured marker.
	tf := parseFile(t, "### Today\n- a\n### Long\n- b\n", "### Long")
	if len(tf.Sections) != 1 || len(tf.Sections[0].Tasks) != 1 {
		t.Fatalf("stop-marker: sections=%d tasks=%d, want 1/1", len(tf.Sections), len(tf.Sections[0].Tasks))
	}
	// Stop on "---".
	tf2 := parseFile(t, "### Today\n- a\n---\n- b\n", "### Long")
	if len(tf2.Sections) != 1 || len(tf2.Sections[0].Tasks) != 1 {
		t.Fatalf("--- : sections=%d tasks=%d, want 1/1", len(tf2.Sections), len(tf2.Sections[0].Tasks))
	}
}

func TestParseStopMarkerSilentSkip(t *testing.T) {
	tf := parseFile(t, "### Today\n- a\n### Long\n# should-not-warn\ngarbage [[x]]\n", "### Long")
	if len(tf.Warnings) != 0 {
		t.Errorf("got warnings after stop marker: %+v", tf.Warnings)
	}
	if len(tf.Sections) != 1 || len(tf.Sections[0].Tasks) != 1 {
		t.Errorf("stop-marker structures wrong: %+v", tf.Sections)
	}
}

func TestParseSectionHeader(t *testing.T) {
	tf := parseFile(t, "### Today\n- a\n###  winter 2031\n- b\n", "")
	if len(tf.Sections) != 2 {
		t.Fatalf("sections = %d, want 2", len(tf.Sections))
	}
	if tf.Sections[0].Title != "Today" || tf.Sections[1].Title != "winter 2031" {
		t.Errorf("titles = [%q %q], want [Today winter 2031]",
			tf.Sections[0].Title, tf.Sections[1].Title)
	}
	if len(tf.Sections[0].Tasks) != 1 || len(tf.Sections[1].Tasks) != 1 {
		t.Errorf("task distribution wrong: %+v", tf.Sections)
	}
}

func TestParseWarnings(t *testing.T) {
	tf := parseFile(t, "# H1\n## H2\n[[bare]]\nloose text\n\t- orphan\n### s\n- ok\n", "")
	wants := []struct {
		text    string
		message string
	}{
		{"# H1", "unrecognized line"},
		{"## H2", "unrecognized line"},
		{"[[bare]]", "unrecognized line"},
		{"loose text", "unrecognized line"},
		{"\t- orphan", "subtask without parent"},
	}
	if len(tf.Warnings) != len(wants) {
		t.Fatalf("warnings = %d, want %d: %+v", len(tf.Warnings), len(wants), tf.Warnings)
	}
	for i, w := range wants {
		if tf.Warnings[i].Text != w.text || tf.Warnings[i].Message != w.message {
			t.Errorf("warning[%d] = %+v, want text=%q message=%q", i, tf.Warnings[i], w.text, w.message)
		}
	}
	if len(tf.Sections) != 1 || len(tf.Sections[0].Tasks) != 1 || tf.Sections[0].Tasks[0].Text != "ok" {
		t.Errorf("valid section/task not parsed: %+v", tf.Sections)
	}
}

func TestParseWarningsLineNumbers(t *testing.T) {
	tf := parseFile(t, "# H1\n## H2\n[[bare]]\nloose\n\t- orphan\n### s\n- bad due:2030-13-45\n", "")
	wantLines := []int{1, 2, 3, 4, 5, 7}
	if len(tf.Warnings) != len(wantLines) {
		t.Fatalf("warnings = %d, want %d: %+v", len(tf.Warnings), len(wantLines), tf.Warnings)
	}
	for i, want := range wantLines {
		if tf.Warnings[i].Line != want {
			t.Errorf("warning[%d] line = %d, want %d", i, tf.Warnings[i].Line, want)
		}
	}
	// The invalid due date inside the section should have line 7.
	last := tf.Warnings[len(tf.Warnings)-1]
	if last.Message == "" || last.Line != 7 {
		t.Errorf("expected inline due-date warning at line 7, got %+v", last)
	}
}

package taskparser

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// subtaskIndent prefixes each subtask line, one tab per nesting level.
	subtaskIndent = "\t"

	// warningFormat renders a single warning line.
	warningFormat = "line %d: %s | %s"
)

// formatTask renders a single task (and, recursively, its subtasks) back to
// human-readable text. It restores the attributes the parser extracted out of
// the raw text: priority, wikilinks, inline links, due date and emojis.
//
// Order: priority → text → wikilinks → inline links → due → emojis. Each
// subtask's lines are indented with one tab per nesting level, so level-one
// subtasks get "\t" and their own subtasks get "\t\t", and so on. The result is
// a single string that may contain newlines when subtasks are present.
func formatTask(item TaskItem) string {
	var emojis strings.Builder
	for _, e := range item.Emojis {
		emojis.WriteString(e)
	}

	var parts []string
	if item.Priority != "" {
		parts = append(parts, "("+item.Priority+")")
	}
	if item.Text != "" {
		parts = append(parts, item.Text)
	}
	for _, w := range item.Wikilinks {
		parts = append(parts, "[["+w+"]]")
	}
	for _, l := range item.InlineLinks {
		parts = append(parts, "["+l.Text+"]("+l.URL+")")
	}
	if !item.Due.IsZero() {
		parts = append(parts, "due:"+item.Due.Format(dateLayout))
	}
	if emojis.Len() > 0 {
		parts = append(parts, emojis.String())
	}
	base := strings.Join(parts, " ")

	if len(item.Subtasks) == 0 {
		return base
	}

	lines := make([]string, 0, 1+len(item.Subtasks))
	lines = append(lines, base)
	for _, sub := range item.Subtasks {
		for _, l := range strings.Split(formatTask(sub), "\n") {
			// An empty subtask renders as a bare line; skip it rather than
			// emitting a line containing only the indent tab.
			if l == "" {
				continue
			}
			lines = append(lines, subtaskIndent+l)
		}
	}
	return strings.Join(lines, "\n")
}

// splitByLimit groups lines into parts no larger than limit characters.
// limit <= 0 disables the limit, returning a single part. Lines are never split
// mid-way: if adding a line would push the current part over the limit, the
// current part is closed and a new one is started. A single oversized line is
// kept as one oversized part. It always returns a non-nil []string, empty for
// empty input.
func splitByLimit(lines []string, limit int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	var result []string
	current := lines[0]
	for _, line := range lines[1:] {
		candidate := current + "\n" + line
		if limit > 0 && utf8.RuneCountInString(candidate) > limit {
			result = append(result, current)
			current = line
			continue
		}
		current = candidate
	}
	result = append(result, current)
	return result
}

// FormatTasks renders each task with formatTask and groups them into parts no
// larger than limit characters. limit <= 0 disables the limit, returning a
// single part. Tasks are never split mid-way: if adding a task would push the
// current part over the limit, the current part is closed and a new one is
// started. A single task (including its subtasks) larger than the limit is kept
// as one oversized part. It always returns a non-nil []string, empty for empty
// input.
func FormatTasks(items []TaskItem, limit int) []string {
	lines := make([]string, 0, len(items))
	for _, it := range items {
		lines = append(lines, formatTask(it))
	}
	return splitByLimit(lines, limit)
}

// FormatWarnings renders each warning as "line N: <message> | <text>" and
// groups them into parts no larger than limit characters. Splitting behaves
// like FormatTasks: limit <= 0 returns a single part and an oversized warning
// is kept as one oversized part. It always returns a non-nil []string, empty
// for empty input.
func FormatWarnings(warnings []Warning, limit int) []string {
	lines := make([]string, 0, len(warnings))
	for _, w := range warnings {
		lines = append(lines, fmt.Sprintf(warningFormat, w.Line, w.Message, w.Text))
	}
	return splitByLimit(lines, limit)
}

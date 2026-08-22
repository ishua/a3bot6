package taskparser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	// dateLayout is the date format used for due:YYYY-MM-DD tokens.
	dateLayout = "2006-01-02"

	// bullet markers introducing a task line.
	bulletDash = "- "
	bulletStar = "* "

	// stopSeparator marks the end of task collection.
	stopSeparator = "---"
)

// knownEmojis is the controlled emoji vocabulary present in task files.
var knownEmojis = map[rune]struct{}{
	'🏡': {},
	'💰': {},
	'📞': {},
	'🏢': {},
}

// Warning describes an unrecognized or partially-invalid line.
type Warning struct {
	Line    int    // 1-based line number
	Text    string // original line text
	Message string // human-readable description
}

// InlineLink is a markdown [text](url) link extracted from a task line.
type InlineLink struct {
	Text string
	URL  string
}

// TaskItem is a parsed task with its extracted inline attributes.
type TaskItem struct {
	Text        string
	Due         time.Time // zero if not specified
	Priority    string    // "A", "B", or ""
	Emojis      []string
	Wikilinks   []string
	Subtasks    []TaskItem
	InlineLinks []InlineLink
}

// Section is a markdown ### heading followed by its tasks.
type Section struct {
	Title string
	Tasks []TaskItem
}

// TaskFile is the result of parsing a task list file.
type TaskFile struct {
	Sections []Section
	Warnings []Warning
}

// ParseTaskFile parses a task list file into a TaskFile. It never returns an
// error: unrecognized lines are collected as warnings. Parsing stops at the
// configured stopMarker (when non-empty) or the "---" separator; all remaining
// lines are silently skipped.
func ParseTaskFile(lines []string, stopMarker string) *TaskFile {
	tf := &TaskFile{
		Sections: make([]Section, 0),
		Warnings: make([]Warning, 0),
	}
	var current *Section
	idx := 0
	for idx < len(lines) {
		line := lines[idx]
		lineNum := idx + 1
		if isStopLine(line, stopMarker) {
			return tf
		}
		if strings.TrimSpace(line) == "" {
			idx++
			continue
		}
		if isSectionHeader(line) {
			title := strings.TrimSpace(strings.TrimPrefix(line, "###"))
			tf.Sections = append(tf.Sections, Section{Title: title, Tasks: make([]TaskItem, 0)})
			current = &tf.Sections[len(tf.Sections)-1]
			endIdx, warns := parseTasksInSection(lines, idx+1, current, stopMarker)
			tf.Warnings = append(tf.Warnings, warns...)
			idx = endIdx
			continue
		}
		if isSubtaskLine(line) {
			if current == nil {
				tf.Warnings = append(tf.Warnings, Warning{Line: lineNum, Text: line, Message: "subtask without parent"})
				idx++
				continue
			}
			end, warns := attachSubtasksToLast(&current.Tasks, lines, idx, lineNum, stopMarker)
			tf.Warnings = append(tf.Warnings, warns...)
			idx = end
			continue
		}
		if isTaskLine(line) {
			task, warns := parseTaskLine(line, lineNum)
			if current == nil {
				tf.Sections = append(tf.Sections, Section{Tasks: make([]TaskItem, 0)})
				current = &tf.Sections[len(tf.Sections)-1]
			}
			current.Tasks = append(current.Tasks, task)
			tf.Warnings = append(tf.Warnings, warns...)
			idx++
			continue
		}
		tf.Warnings = append(tf.Warnings, Warning{Line: lineNum, Text: line, Message: "unrecognized line"})
		idx++
	}
	return tf
}

// parseTaskLine turns a single task/subtask line into a TaskItem, applying all
// inline extractions in order: priority, due date, wikilinks, inline links,
// emojis, then the remaining trimmed text. Warnings (invalid due date, unclosed
// wikilink) are returned with the line number filled in.
func parseTaskLine(line string, lineNum int) (TaskItem, []Warning) {
	warnings := make([]Warning, 0)
	content := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(content, bulletDash) || strings.HasPrefix(content, bulletStar) {
		content = content[2:]
	}

	priority, rest := extractPriority(content)

	due, rest, warns := extractDueDate(rest)
	decorateWarnings(warns, lineNum, line)
	warnings = append(warnings, warns...)

	wikilinks, rest, warns := extractWikilinks(rest)
	decorateWarnings(warns, lineNum, line)
	warnings = append(warnings, warns...)

	links, rest := extractInlineLinks(rest)

	emojis, rest := extractEmojis(rest)

	return TaskItem{
		Text:        rest,
		Due:         due,
		Priority:    priority,
		Emojis:      emojis,
		Wikilinks:   wikilinks,
		Subtasks:    make([]TaskItem, 0),
		InlineLinks: links,
	}, warnings
}

// decorateWarnings fills in the line number and original line text on warnings
// produced by inline extraction helpers that do not know their source line.
func decorateWarnings(warns []Warning, lineNum int, line string) {
	for i := range warns {
		warns[i].Line = lineNum
		warns[i].Text = line
	}
}

// indentDepth returns the number of leading whitespace (spaces or tabs)
// characters in line.
func indentDepth(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

// parseSubtasks parses a run of consecutive subtask lines starting at startIdx,
// recursing into deeper-indented descendants. It returns the collected subtasks
// and the index of the first line after the run. lineNumBase is the 1-based line
// number of the line at startIdx. Parsing silently stops at the configured
// stopMarker or "---" separator.
func parseSubtasks(lines []string, startIdx int, lineNumBase int, stopMarker string) (subtasks []TaskItem, endIdx int, warnings []Warning) {
	subtasks = make([]TaskItem, 0)
	warnings = make([]Warning, 0)
	myDepth := indentDepth(lines[startIdx])
	i := startIdx
	for i < len(lines) {
		line := lines[i]
		if isStopLine(line, stopMarker) {
			return subtasks, i, warnings
		}
		if !isSubtaskLine(line) {
			return subtasks, i, warnings
		}
		d := indentDepth(line)
		switch {
		case d < myDepth:
			// Belongs to an outer level; back off.
			return subtasks, i, warnings
		case d == myDepth:
			task, warns := parseTaskLine(line, lineNumBase+(i-startIdx))
			subtasks = append(subtasks, task)
			warnings = append(warnings, warns...)
			i++
		default: // d > myDepth: child of the last sibling.
			if len(subtasks) == 0 {
				warnings = append(warnings, Warning{Line: lineNumBase + (i - startIdx), Text: line, Message: "subtask without parent"})
				i++
				continue
			}
			last := &subtasks[len(subtasks)-1]
			childSubs, end, warns := parseSubtasks(lines, i, lineNumBase+(i-startIdx), stopMarker)
			last.Subtasks = append(last.Subtasks, childSubs...)
			warnings = append(warnings, warns...)
			i = end
		}
	}
	return subtasks, i, warnings
}

// attachSubtasksToLast appends a run of subtask lines starting at startIdx to
// the last task in tasks. If tasks is empty it records a "subtask without
// parent" warning. It returns the index of the next line plus any warnings.
func attachSubtasksToLast(tasks *[]TaskItem, lines []string, startIdx int, lineNum int, stopMarker string) (int, []Warning) {
	warnings := make([]Warning, 0)
	if len(*tasks) == 0 {
		warnings = append(warnings, Warning{Line: lineNum, Text: lines[startIdx], Message: "subtask without parent"})
		return startIdx + 1, warnings
	}
	last := &(*tasks)[len(*tasks)-1]
	subs, end, warns := parseSubtasks(lines, startIdx, startIdx+1, stopMarker)
	last.Subtasks = append(last.Subtasks, subs...)
	warnings = append(warnings, warns...)
	return end, warnings
}

// parseTasksInSection parses the body of a section, starting at startIdx, until
// the next section header, the configured stopMarker or "---" stop line, or the
// end of input. It fills section.Tasks (including subtask attachment) and
// returns the index of the first unprocessed line plus any warnings.
func parseTasksInSection(lines []string, startIdx int, section *Section, stopMarker string) (endIdx int, warnings []Warning) {
	warnings = make([]Warning, 0)
	i := startIdx
	for i < len(lines) {
		line := lines[i]
		lineNum := i + 1
		if isStopLine(line, stopMarker) || isSectionHeader(line) {
			return i, warnings
		}
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		if isSubtaskLine(line) {
			end, warns := attachSubtasksToLast(&section.Tasks, lines, i, lineNum, stopMarker)
			warnings = append(warnings, warns...)
			i = end
			continue
		}
		if isTaskLine(line) {
			task, warns := parseTaskLine(line, lineNum)
			section.Tasks = append(section.Tasks, task)
			warnings = append(warnings, warns...)
			i++
			continue
		}
		warnings = append(warnings, Warning{Line: lineNum, Text: line, Message: "unrecognized line"})
		i++
	}
	return i, warnings
}

// clean collapses runs of whitespace in text into single spaces and trims it.
func clean(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// isTaskLine reports whether line is a task line: it starts with "- " or "* ".
func isTaskLine(line string) bool {
	return strings.HasPrefix(line, bulletDash) || strings.HasPrefix(line, bulletStar)
}

// isSubtaskLine reports whether line is a subtask line: it has leading
// tabs/spaces, then starts with "- " or "* ".
func isSubtaskLine(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	return trimmed != line && isTaskLine(trimmed)
}

// isSectionHeader reports whether line is a markdown ### section header.
func isSectionHeader(line string) bool {
	return strings.HasPrefix(line, "### ")
}

// isStopLine reports whether line is a stop marker: it equals the configured
// stopMarker (when non-empty) or the "---" separator.
func isStopLine(line, stopMarker string) bool {
	if stopMarker != "" && line == stopMarker {
		return true
	}
	return line == stopSeparator
}

// extractPriority pulls a leading "(A)" or "(B)" marker off line, returning the
// priority and the remainder of the text.
func extractPriority(line string) (priority string, rest string) {
	switch {
	case strings.HasPrefix(line, "(A)"):
		return "A", strings.TrimSpace(strings.TrimPrefix(line, "(A)"))
	case strings.HasPrefix(line, "(B)"):
		return "B", strings.TrimSpace(strings.TrimPrefix(line, "(B)"))
	}
	return "", line
}

var dueDateRe = regexp.MustCompile(`\bdue:(\S+)`)

// extractDueDate parses a "due:YYYY-MM-DD" token out of text. On a valid date
// it returns the parsed time and the text minus the token. On a malformed date
// it returns a non-nil warning; the caller is responsible for filling in the
// line number.
func extractDueDate(text string) (due time.Time, rest string, warnings []Warning) {
	warnings = make([]Warning, 0)
	loc := dueDateRe.FindStringSubmatchIndex(text)
	if loc == nil {
		return time.Time{}, text, warnings
	}
	token := text[loc[0]:loc[1]]
	value := text[loc[2]:loc[3]]
	rest = clean(text[:loc[0]] + text[loc[1]:])
	dueParsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, rest, []Warning{{Message: fmt.Sprintf("invalid due date %q", token)}}
	}
	return dueParsed, rest, warnings
}

// extractEmojis pulls out any emoji in knownEmojis found anywhere in text,
// returning them in order of appearance and the text with them removed.
func extractEmojis(text string) (emojis []string, rest string) {
	emojis = make([]string, 0)
	var sb strings.Builder
	for _, r := range text {
		if _, ok := knownEmojis[r]; ok {
			emojis = append(emojis, string(r))
		} else {
			sb.WriteRune(r)
		}
	}
	return emojis, clean(sb.String())
}

// extractWikilinks pulls out all "[[wikilink]]" tokens. An unclosed "[["
// produces a warning; the caller fills in the line number.
func extractWikilinks(text string) (wikilinks []string, rest string, warnings []Warning) {
	wikilinks = make([]string, 0)
	warnings = make([]Warning, 0)
	rest = text
	for {
		start := strings.Index(rest, "[[")
		if start == -1 {
			break
		}
		end := strings.Index(rest[start+2:], "]]")
		if end == -1 {
			warnings = append(warnings, Warning{Message: "unclosed wikilink"})
			rest = rest[:start]
			break
		}
		wikilinks = append(wikilinks, rest[start+2:start+2+end])
		rest = rest[:start] + rest[start+2+end+2:]
	}
	return wikilinks, clean(rest), warnings
}

var inlineLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// extractInlineLinks pulls out all "[text](url)" links, leaving the remaining
// text.
func extractInlineLinks(text string) (links []InlineLink, rest string) {
	links = make([]InlineLink, 0)
	for _, m := range inlineLinkRe.FindAllStringSubmatch(text, -1) {
		links = append(links, InlineLink{Text: m[1], URL: m[2]})
	}
	return links, clean(inlineLinkRe.ReplaceAllString(text, ""))
}

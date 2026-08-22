package taskparser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

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
// error: unrecognized lines are collected as warnings.
func ParseTaskFile(lines []string, stopMarker string) *TaskFile {
	return &TaskFile{}
}

// clean collapses runs of whitespace in text into single spaces and trims it.
func clean(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// isTaskLine reports whether line is a task line: it starts with "- " or "* ".
func isTaskLine(line string) bool {
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")
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
	return line == "---"
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

var dueDateRe = regexp.MustCompile(`due:(\S+)`)

// extractDueDate parses a "due:YYYY-MM-DD" token out of text. On a valid date
// it returns the parsed time and the text minus the token. On a malformed date
// it returns a non-nil warning; the caller is responsible for filling in the
// line number.
func extractDueDate(text string) (due time.Time, rest string, warning *Warning) {
	loc := dueDateRe.FindStringSubmatchIndex(text)
	if loc == nil {
		return time.Time{}, text, nil
	}
	token := text[loc[0]:loc[1]]
	value := text[loc[2]:loc[3]]
	rest = clean(text[:loc[0]] + text[loc[1]:])
	dueParsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, rest, &Warning{Message: fmt.Sprintf("invalid due date %q", token)}
	}
	return dueParsed, rest, nil
}

// extractEmojis pulls out any known emoji characters (🏡 💰 📞 🏢) found anywhere
// in text, returning them in order of appearance and the text with them removed.
func extractEmojis(text string) (emojis []string, rest string) {
	emojis = make([]string, 0)
	var sb strings.Builder
	for _, r := range text {
		switch r {
		case '🏡', '💰', '📞', '🏢':
			emojis = append(emojis, string(r))
		default:
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

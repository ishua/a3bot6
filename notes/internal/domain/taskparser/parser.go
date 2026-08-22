package taskparser

import "time"

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
	Due         time.Time  // zero if not specified
	Priority    string     // "A", "B", or ""
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

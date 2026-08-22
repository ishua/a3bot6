# note-task-parser: Phase 1 — Tolerant markdown parser with warnings

## Overview

Implement Phase 1 of the note task list feature: a tolerant markdown parser for `notes/internal/domain/taskparser` package. The parser reads a markdown task list file, extracts sections and tasks, and collects warnings for any unrecognized lines. This is a standalone package with no dependencies on other internal packages.

## Context

- **Project**: a3bot6 — Go services communicating via REST API
- **Service**: notes (Go) — Git-based notes service
- **Feature**: `/note task` command for parsing a markdown task list file
- **Full feature plan**: `.opencode/plans/2026-06-17-note-task-list.md`
- **This plan**: Phase 1 only (parser)
- **Test data**: `notes/test_data/tasklist.md` — 40 lines with H1/H2 headers, `###` sections, tasks with due dates, emoji tags, wikilinks, subtasks, stop marker `### Long`, `---` separator, loose text after stop marker

## Development Approach

- **Testing approach**: TDD (write tests first, then implementation)
- Complete each task before moving to next
- **CRITICAL**: every task MUST include new/updated tests
- **CRITICAL**: all tests must pass before starting next task

## Testing Strategy

- Unit tests in `notes/internal/domain/taskparser/parser_test.go`
- Test data in `notes/internal/domain/taskparser/testdata/tasklist.md`
- Inline table-driven tests for edge cases
- Full file parsing test for integration validation

## Solution Overview

A tolerant parser that:
- Always reads the entire input and returns a `*TaskFile` (never returns an error)
- Collects warnings for unrecognized lines: line number, text, description
- Recognizes sections (`###`), tasks (`-`/`*`), subtasks (indented), stop markers
- Extracts inline attributes: priority, due dates, emojis, wikilinks, inline links
- Stops task collection at configured `stopMarker` OR `---`

## Technical Details

### Data Structures

```go
type Warning struct {
    Line    int    // 1-based line number
    Text    string // original line text
    Message string // human-readable description
}

type InlineLink struct {
    Text string
    URL  string
}

type TaskItem struct {
    Text      string
    Due       time.Time      // zero if not specified
    Priority  string         // "A", "B", or ""
    Emojis    []string
    Wikilinks []string
    Subtasks  []TaskItem
    InlineLinks []InlineLink
}

type Section struct {
    Title string
    Tasks []TaskItem
}

type TaskFile struct {
    Sections []Section
    Warnings []Warning
}
```

### Parser Signature

```go
func ParseTaskFile(lines []string, stopMarker string) *TaskFile
```

### Line Processing Rules

| Line Pattern | Action |
|--------------|--------|
| Empty line | Skip silently |
| `### ...` | New Section (title = trimmed after `###`) |
| `- ` or `* ` (task) | TaskItem in current section |
| Tab/space + `- ` or `* ` (subtask) | Subtask under last TaskItem |
| `stopMarker` OR `---` | Stop parsing tasks, ignore rest |
| Unrecognized | Warning |

### Inline Attribute Extraction (per task line)

| Pattern | Extraction |
|---------|------------|
| `^(A)` / `^(B)` | Priority = "A" or "B" |
| `due:YYYY-MM-DD` | Parse to `time.Time`; invalid → Warning |
| Trailing emoji (`🏡`, `💰`, `📞`, `🏢`, etc.) | Collect in `Emojis` |
| `[[wikilink]]` | Collect in `Wikilinks`; unclosed `[[` → Warning |
| `[text](url)` | `InlineLink{Text, URL}` |

## What Goes Where

- **Implementation**: `notes/internal/domain/taskparser/parser.go`
- **Tests**: `notes/internal/domain/taskparser/parser_test.go`
- **Test data**: `notes/internal/domain/taskparser/testdata/tasklist.md`
- **No dependencies**: package has no imports from `mcore/pkg/schema` or other internal packages

## Implementation Steps

### Task 1: Create package structure and data types
**Files:**
- Create: `notes/internal/domain/taskparser/parser.go`

- [ ] Create directory `notes/internal/domain/taskparser/`
- [ ] Define `Warning`, `InlineLink`, `TaskItem`, `Section`, `TaskFile` structs
- [ ] Define `ParseTaskFile(lines []string, stopMarker string) *TaskFile` function signature
- [ ] Add required imports (`time`, `fmt`)
- [ ] run `go build` — verify package compiles

### Task 2: Copy test data
**Files:**
- Create: `notes/internal/domain/taskparser/testdata/tasklist.md`

- [ ] Create directory `notes/internal/domain/taskparser/testdata/`
- [ ] Copy `notes/test_data/tasklist.md` → `notes/internal/domain/taskparser/testdata/tasklist.md`
- [ ] Verify file content matches (40 lines)

### Task 3: Write inline helper functions
**Files:**
- Modify: `notes/internal/domain/taskparser/parser.go`

- [ ] Write `isTaskLine(line string) bool` — returns true for lines starting with `- ` or `* `
- [ ] Write `isSubtaskLine(line string) bool` — returns true for lines with leading tab/spaces then `- ` or `* `
- [ ] Write `isSectionHeader(line string) bool` — returns true for lines starting with `### `
- [ ] Write `isStopLine(line, stopMarker string) bool` — returns true for `line == stopMarker` or `line == "---"`
- [ ] Write `extractPriority(line string) (priority string, rest string)` — extracts `(A)` or `(B)`
- [ ] Write `extractDueDate(text string) (time.Time, string, bool)` — parses `due:YYYY-MM-DD`
- [ ] Write `extractEmojis(text string) (emojis []string, rest string)` — extracts trailing emoji characters
- [ ] Write `extractWikilinks(text string) (wikilinks []string, rest string, warnings []Warning)` — extracts `[[wikilink]]`
- [ ] Write `extractInlineLinks(text string) (links []InlineLink, rest string)` — extracts `[text](url)`
- [ ] Write tests for all helper functions (table-driven)
- [ ] run tests — must pass before next task

### Task 4: Implement section and task parsing
**Files:**
- Modify: `notes/internal/domain/taskparser/parser.go`
- Modify: `notes/internal/domain/taskparser/parser_test.go`

- [ ] Write `parseTaskLine(line string) TaskItem` — applies all inline extractions
- [ ] Write `parseSubtasks(lines []string, startIdx int) (subtasks []TaskItem, endIdx int)` — recursive subtask parsing
- [ ] Write `parseTasksInSection(lines []string, startIdx int, section *Section) (endIdx int, warnings []Warning)`
- [ ] Write `ParseTaskFile(lines []string, stopMarker string) *TaskFile` — main parser
- [ ] Write `TestParseFullFile` — parse testdata/tasklist.md, verify 4 sections (Today, summer 2030, fall 2030, winter 2031)
- [ ] Write `TestParseEmptyFile` — empty input → empty TaskFile, no warnings
- [ ] Write `TestParsePriority` — `(A)` and `(B)` prefix extraction
- [ ] Write `TestParseDueDate` — valid due, missing due, invalid due date
- [ ] Write `TestParseEmoji` — emoji tags at end of line
- [ ] Write `TestParseWikilink` — `[[link]]` in different positions, unclosed wikilink
- [ ] Write `TestParseInlineLink` — `[text](url)` extraction
- [ ] Write `TestParseSubtask` — nested subtasks with all attributes
- [ ] Write `TestParseStopMarker` — stop on `### Long` and `---`
- [ ] Write `TestParseSectionHeader` — `### ...` creates sections
- [ ] Write `TestParseWarnings` — unrecognized lines: H1, H2, bare wikilink, loose text, subtask without parent
- [ ] Write `TestParseWarningsLineNumbers` — correct 1-based line numbers in warnings
- [ ] run tests — must all pass

### Task 5: Verify acceptance criteria
**Files:** (none)

- [ ] Parser returns `*TaskFile` for all inputs (never nil, never errors)
- [ ] Sections correctly parsed: Today, summer 2030, fall 2030, winter 2031
- [ ] Tasks contain: priority (A/B), due dates, emojis, wikilinks, inline links
- [ ] Subtasks parsed with correct parent relationship
- [ ] Stop marker (`### Long`) and `---` both stop task collection
- [ ] Warnings generated for: `#` H1, `##` H2, bare wikilink on line, loose text, invalid due date
- [ ] Warnings include correct 1-based line numbers
- [ ] run full test suite — all tests pass

### Task 6: Build verification and vet
**Files:** (none)

- [ ] `cd notes && go build ./internal/domain/taskparser/...` — compiles cleanly
- [ ] `cd notes && go test ./internal/domain/taskparser/...` — all tests pass
- [ ] `cd notes && go vet ./internal/domain/taskparser/...` — no issues

## Post-Completion

- No external systems need updating
- Manual checks: review generated warnings match expected output from test file
- Next phase (Phase 2): wire parser into notes service REST API
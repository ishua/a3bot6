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
  - Exception: Task 2 (data copy only)
- **CRITICAL**: all tests must pass before starting next task

## Testing Strategy

- **Test package**: `taskparser` (internal test, same package — needed to test unexported helpers)
- **Assertions**: stdlib `testing` only (no testify — project has zero test deps currently)
- **Style**: table-driven tests where applicable
- Unit tests in `notes/internal/domain/taskparser/parser_test.go`
- Test data in `notes/internal/domain/taskparser/testdata/tasklist.md`
- Inline table-driven tests for edge cases
- Full file parsing test for integration validation
- **File reading in tests**: `os.ReadFile` + `strings.Split(string(data), "\n")`

## Solution Overview

A tolerant parser that:
- Always reads the entire input and returns a `*TaskFile` (never returns an error)
- Collects warnings for unrecognized lines: line number, text, description. Warnings from inline extraction (invalid due dates, unclosed wikilinks) are propagated into `TaskFile.Warnings`
- Recognizes sections (`###`), tasks (`-`/`*`), subtasks (indented), stop markers
- Extracts inline attributes: priority, due dates, emojis, wikilinks, inline links
- Stops task collection at configured `stopMarker` OR `---`. **After stop marker: all remaining lines are silently skipped** — no further task collection, no further warnings

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

**Check order in main loop** (first match wins):

1. If `stopMarker != ""` AND `line == stopMarker` → stop (silent skip rest)
2. `line == "---"` → stop (silent skip rest)
3. Empty line (after trim) → skip silently
4. `### ` prefix → Section (title = `strings.TrimSpace(strings.TrimPrefix(line, "###"))`)
5. Indented `- `/`* ` → subtask under last task
6. `- `/`* ` at start → task in current section
7. Anything else → Warning

| Line Pattern | Action |
|--------------|--------|
| Empty line | Skip silently |
| `### ...` | New Section (title = `TrimSpace(TrimPrefix(line, "###"))` — handles double spaces) |
| `- ` or `* ` (task) | TaskItem in current section |
| Tab/space + `- ` or `* ` (subtask) | Subtask under last TaskItem |
| `stopMarker` (if non-empty) OR `---` | Stop parsing, **silently skip all remaining lines** (no warnings) |
| Unrecognized | Warning |

### Inline Attribute Extraction (per task line)

**Extraction order** (important — emoji is NOT always trailing in test data, e.g. `💰 due:2030-07-13`):

1. **Priority**: `(A)` or `(B)` at start of task text → strip from text
2. **Due date**: `due:YYYY-MM-DD` anywhere in text → parse to `time.Time`, strip from text. Invalid date → warning, Due stays zero
3. **Wikilinks**: `[[link]]` anywhere → extract, strip from text. Unclosed `[[` → warning
4. **Inline links**: `[text](url)` anywhere → extract, strip from text
5. **Emojis**: scan remaining text for known emoji characters (see detection below), collect, strip from text
6. **Text**: what remains after all extractions, trimmed — this is `TaskItem.Text`

Example: `- balance somethin [[finance]] 💰 due:2030-07-13` →
`TaskItem{Text: "balance somethin", Due: 2030-07-13, Wikilinks: ["finance"], Emojis: ["💰"]}`

**Emoji detection**: use a predefined set of known emoji from the task file: `🏡 💰 📞 🏢`.
Extract by scanning the text for any of these characters and removing them.
This avoids unicode-range complexity and matches the controlled vocabulary of the file.

| Pattern | Extraction |
|---------|------------|
| `^(A)` / `^(B)` | Priority = "A" or "B" |
| `due:YYYY-MM-DD` | Parse to `time.Time`; invalid → Warning |
| Emoji (`🏡`, `💰`, `📞`, `🏢`) anywhere in remaining text | Collect in `Emojis` |
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

- [x] Create directory `notes/internal/domain/taskparser/`
- [x] Define `Warning`, `InlineLink`, `TaskItem`, `Section`, `TaskFile` structs
- [x] Define `ParseTaskFile(lines []string, stopMarker string) *TaskFile` function signature
- [x] Add required import (`time` only at this stage; `fmt` added in Task 3)
- [x] run `go build` — verify package compiles

### Task 2: Copy test data
**Files:**
- Create: `notes/internal/domain/taskparser/testdata/tasklist.md`

- [x] Create directory `notes/internal/domain/taskparser/testdata/`
- [x] Copy `notes/test_data/tasklist.md` → `notes/internal/domain/taskparser/testdata/tasklist.md`
- [x] Verify file content matches (40 lines)

### Task 3: Write inline helper functions
**Files:**
- Modify: `notes/internal/domain/taskparser/parser.go`

- [x] Write `isTaskLine(line string) bool` — returns true for lines starting with `- ` or `* `
- [x] Write `isSubtaskLine(line string) bool` — returns true for lines with leading tab/spaces then `- ` or `* `
- [x] Write `isSectionHeader(line string) bool` — returns true for lines starting with `### `
- [x] Write `isStopLine(line, stopMarker string) bool` — returns true for `line == stopMarker` (only if `stopMarker != ""`) or `line == "---"`
- [x] Write `extractPriority(line string) (priority string, rest string)` — extracts `(A)` or `(B)`
- [x] Write `extractDueDate(text string) (due time.Time, rest string, warning *Warning)` — parses `due:YYYY-MM-DD`; invalid date → non-nil warning (caller fills line number)
- [x] Write `extractEmojis(text string) (emojis []string, rest string)` — extracts known emoji (`🏡 💰 📞 🏢`) from anywhere in text
- [x] Write `extractWikilinks(text string) (wikilinks []string, rest string, warnings []Warning)` — extracts `[[wikilink]]`
- [x] Write `extractInlineLinks(text string) (links []InlineLink, rest string)` — extracts `[text](url)`
- [x] Write tests for all helper functions (table-driven)
- [x] run tests — must pass before next task

### Task 4: Implement section and task parsing
**Files:**
- Modify: `notes/internal/domain/taskparser/parser.go`
- Modify: `notes/internal/domain/taskparser/parser_test.go`

- [x] Write `parseTaskLine(line string, lineNum int) (TaskItem, []Warning)` — applies all inline extractions in defined order, returns warnings from due-date and wikilink extraction (with line numbers filled in)
- [x] Write `parseSubtasks(lines []string, startIdx int, lineNumBase int) (subtasks []TaskItem, endIdx int, warnings []Warning)` — recursive subtask parsing
- [x] Write `parseTasksInSection(lines []string, startIdx int, section *Section) (endIdx int, warnings []Warning)`
- [x] Write `ParseTaskFile(lines []string, stopMarker string) *TaskFile` — main parser
- [x] Write `TestParseFullFile` — parse testdata/tasklist.md, verify 4 sections (Today, summer 2030, fall 2030, winter 2031)
- [x] Write `TestParseEmptyFile` — empty input → empty TaskFile, no warnings
- [x] Write `TestParsePriority` — `(A)` and `(B)` prefix extraction
- [x] Write `TestParseDueDate` — valid due, missing due, invalid due date
- [x] Write `TestParseEmoji` — emoji tags anywhere in line (including before due date)
- [x] Write `TestParseWikilink` — `[[link]]` in different positions, unclosed wikilink
- [x] Write `TestParseInlineLink` — `[text](url)` extraction
- [x] Write `TestParseSubtask` — nested subtasks with all attributes
- [x] Write `TestParseSubtaskNested` — multi-level indentation (inline test with 2+ levels)
- [x] Write `TestParseStopMarker` — stop on `### Long` and `---`
- [x] Write `TestParseStopMarkerSilentSkip` — lines after stop marker do NOT generate warnings
- [x] Write `TestParseSectionHeader` — `### ...` creates sections (including double-space in `###  winter 2031`)
- [x] Write `TestParseWarnings` — unrecognized lines: H1, H2, bare wikilink, loose text, subtask without parent
- [x] Write `TestParseWarningsLineNumbers` — correct 1-based line numbers in warnings
- [x] run tests — must all pass

### Task 5: Verify acceptance criteria
**Files:** (none)

- [x] Parser returns `*TaskFile` for all inputs (never nil, never errors)
- [x] Sections correctly parsed: Today, summer 2030, fall 2030, winter 2031
- [x] Tasks contain: priority (A/B), due dates, emojis, wikilinks, inline links
- [x] Subtasks parsed with correct parent relationship
- [x] Stop marker (`### Long`) and `---` both stop task collection
- [x] Lines after stop marker are silently skipped (no warnings)
- [x] Warnings generated for: `#` H1, `##` H2, bare wikilink on line, loose text, invalid due date
- [x] Warnings from inline extraction (invalid due date, unclosed wikilink) propagated to `TaskFile.Warnings`
- [x] Warnings include correct 1-based line numbers
- [x] run full test suite — all tests pass

### Task 6: Build verification and vet
**Files:** (none)

- [x] `cd notes && go build ./internal/domain/taskparser/...` — compiles cleanly
- [x] `cd notes && go test ./internal/domain/taskparser/...` — all tests pass
- [x] `cd notes && go vet ./internal/domain/taskparser/...` — no issues

## Post-Completion

- No external systems need updating
- Manual checks: review generated warnings match expected output from test file
- Next phase (Phase 2): wire parser into notes service REST API
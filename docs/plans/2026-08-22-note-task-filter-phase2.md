# note-task-filter: Phase 2 — task filtering

## Overview

Implement Phase 2 of the note task list feature: four standalone filter functions in `notes/internal/domain/taskparser/`. Each filter extracts a specific subset of tasks from a parsed `*TaskFile`. This phase is purely computational — no I/O, no REST API wiring, no dependencies on other packages.

## Context

- **Project**: a3bot6 — Go services communicating via REST API
- **Service**: notes (Go) — Git-based notes service
- **Feature**: `/note task` command — Phase 2 is filtering only
- **Phase 1 (DONE)**: `notes/internal/domain/taskparser/parser.go` — `ParseTaskFile(lines []string, stopMarker string) *TaskFile` with types `TaskFile{Sections []Section, Warnings []Warning}`, `Section{Title string, Tasks []TaskItem}`, `TaskItem{Text string, Due time.Time, Priority string, Emojis []string, Wikilinks []string, Subtasks []TaskItem, InlineLinks []InlineLink}`. Due is zero `time.Time` when not specified. Package has no deps on mcore/schema or other internal packages.
- **Test data**: `notes/internal/domain/taskparser/testdata/tasklist.md` — 4 sections (Today, summer 2030, fall 2030, winter 2031), tasks with due dates, emojis, wikilinks, subtasks, stop marker `### Long` / `---`. File has 40 lines.
- **Full feature plan**: `.opencode/plans/2026-06-17-note-task-list.md` (Phase 2 section, lines 58-68)
- **Testing**: stdlib `testing` only (no testify), table-driven, internal test package `package taskparser`

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task before moving to next
- **CRITICAL**: every task MUST include new/updated tests
- **CRITICAL**: all tests must pass before starting next task

## Testing Strategy

- **Test package**: `taskparser` (internal test, same package — needed to access unexported helpers if any)
- **Assertions**: stdlib `testing` only (no testify — project has zero test deps currently)
- **Style**: table-driven tests where applicable
- Unit tests in `notes/internal/domain/taskparser/filter_test.go`
- Integration test that parses `testdata/tasklist.md` and verifies all 4 filter outputs against known content
- `os.ReadFile` + `strings.Split` for file loading in integration test
- All filter functions are exported (capitalized), so tests can live in the same package

## Solution Overview

Four independent filter functions:

1. **FilterAll** — all top-level tasks from all sections, in file order, no sorting
2. **FilterToday** — entire "Today" section (all tasks, regardless of due) PLUS tasks from other sections where `due <= now`. File order. If no "Today" section exists, returns only overdue tasks from other sections
3. **FilterDue** — all tasks (any section) with non-zero Due, sorted ascending by Due
4. **FilterDueToday** — all tasks with non-zero Due AND `due <= now`, sorted ascending by Due

Subtasks stay nested inside their parent `TaskItem.Subtasks` — they are never flattened, never filtered independently. Empty results return non-nil empty `[]TaskItem{}`.

## Technical Details

### Date Comparison

- `due` values are parsed as `time.Time` at midnight UTC (via `time.Parse("2006-01-02", ...)`)
- `now` is truncated to date: `now.Truncate(24 * time.Hour)`
- Inclusion test: `!due.After(now)` (equivalent to `due.Before(now) || due.Equal(now)`)
- This means a task due on today's date is included when `now` is today (or any time today)

### Today Section Matching

- Exact string match: `section.Title == "Today"` (case-sensitive, no normalization)
- If no section with title "Today" exists, `FilterToday` returns only overdue tasks from other sections

### Sorting

- `FilterDue` and `FilterDueToday` sort by `Due` ascending using `sort.Slice` with a stable compare
- Tasks with non-zero due sort before tasks with zero due (though zero-due tasks are excluded from these two filters anyway)
- `FilterAll` and `FilterToday` preserve file order (no sorting)

### Subtask Handling

- All four filters operate on top-level tasks only — `tf.Sections[i].Tasks`
- Subtasks (`TaskItem.Subtasks`) are returned together with their parent, unchanged
- Subtasks are never flattened, never filtered independently

### Empty Results

- Return `[]TaskItem{}` (non-nil empty slice) for consistency with Phase 1 convention

## Data Structures

No new types — reuse all types from Phase 1 (`TaskFile`, `Section`, `TaskItem`, `InlineLink`, `Warning`).

## Function Signatures

```go
// FilterAll returns all top-level tasks from every section, in file order.
func FilterAll(tf *TaskFile) []TaskItem

// FilterToday returns tasks for the "Today" section (all tasks, regardless
// of due date) plus overdue/today tasks from other sections (due <= now).
// File order is preserved; no sorting.
func FilterToday(tf *TaskFile, now time.Time) []TaskItem

// FilterDue returns all tasks (any section) that have a non-zero Due,
// sorted ascending by Due date.
func FilterDue(tf *TaskFile) []TaskItem

// FilterDueToday returns all tasks that have a non-zero Due and
// due <= now, sorted ascending by Due date.
func FilterDueToday(tf *TaskFile, now time.Time) []TaskItem
```

## What Goes Where

| File | Purpose |
|------|---------|
| `notes/internal/domain/taskparser/filter.go` | Implementation: the 4 filter functions + sort helper |
| `notes/internal/domain/taskparser/filter_test.go` | Tests for all 4 functions |
| `notes/internal/domain/taskparser/parser.go` | Phase 1 — no changes needed |

No new dependencies. Only stdlib: `time`, `sort`.

## Implementation Steps

### Task 1: Implement FilterAll + tests
**Files:**
- Create: `notes/internal/domain/taskparser/filter.go`
- Create: `notes/internal/domain/taskparser/filter_test.go`

- [ ] Write `FilterAll(tf *TaskFile) []TaskItem` — iterate `tf.Sections`, collect all `section.Tasks` into result slice in order
- [ ] Write test `TestFilterAll` — empty TaskFile → empty slice; single section; multiple sections; verify order preserved
- [ ] run tests — must pass before next task

### Task 2: Implement FilterDue + sort helper + tests
**Files:**
- Modify: `notes/internal/domain/taskparser/filter.go`
- Modify: `notes/internal/domain/taskparser/filter_test.go`

- [ ] Write `sortByDue(tasks []TaskItem)` — stable sort ascending by `Due`, zero-due tasks last
- [ ] Write `FilterDue(tf *TaskFile) []TaskItem` — collect all top-level tasks with non-zero Due, apply `sortByDue`
- [ ] Write test `TestFilterDue` — empty; no due dates → empty; tasks with due dates → sorted ascending; verify non-zero due tasks only
- [ ] run tests — must pass before next task

### Task 3: Implement FilterToday + tests
**Files:**
- Modify: `notes/internal/domain/taskparser/filter.go`
- Modify: `notes/internal/domain/taskparser/filter_test.go`

- [ ] Write `FilterToday(tf *TaskFile, now time.Time) []TaskItem` — truncate `now` to date; find "Today" section; collect all its tasks; collect overdue/today tasks from other sections; preserve file order
- [ ] Write test `TestFilterToday` — no "Today" section → only overdue from others; "Today" section with no-due tasks → all included; "Today" section tasks included regardless of due; tasks from other sections included only if due <= now
- [ ] run tests — must pass before next task

### Task 4: Implement FilterDueToday + tests
**Files:**
- Modify: `notes/internal/domain/taskparser/filter.go`
- Modify: `notes/internal/domain/taskparser/filter_test.go`

- [ ] Write `FilterDueToday(tf *TaskFile, now time.Time) []TaskItem` — truncate `now`; collect all top-level tasks with non-zero Due AND `due <= now`; apply `sortByDue`
- [ ] Write test `TestFilterDueToday` — empty; no due dates → empty; only today/overdue tasks included; sorted ascending; tasks in the future excluded
- [ ] run tests — must pass before next task

### Task 5: Integration test — parse testdata/tasklist.md, run all 4 filters
**Files:**
- Modify: `notes/internal/domain/taskparser/filter_test.go`

- [ ] Load `testdata/tasklist.md`, parse with `ParseTaskFile(..., "### Long")`
- [ ] `TestFilterAll`: verify total count of top-level tasks from all 4 sections (Today: 8 tasks, summer: 3, fall: 1, winter: 1 → total 13); verify first and last task text
- [ ] `TestFilterDue`: verify tasks with due dates are returned; verify ordering (2030-07-06, 2030-07-13, 2030-08-09, 2030-10-13, 2031-01-15)
- [ ] `TestFilterToday`: pick a `now` of 2030-07-10 — should include all Today section tasks (8) plus the 2030-07-06 overdue task from Today; summer task 2030-07-13 is future → excluded; verify count
- [ ] `TestFilterDueToday`: same `now` 2030-07-10 — should include: Today task with due 2030-07-06 (overdue), and summer task 2030-07-13 is future → excluded; verify count and sorting
- [ ] `TestFilterDueToday_noDueIncluded`: verify tasks without due dates are never included in `FilterDue` / `FilterDueToday`
- [ ] run tests — must all pass

### Task 6: Build verification and vet
**Files:** (none)

- [ ] `cd notes && go build ./internal/domain/taskparser/...` — compiles cleanly
- [ ] `cd notes && go test ./internal/domain/taskparser/...` — all tests pass
- [ ] `cd notes && go vet ./internal/domain/taskparser/...` — no issues

### Task 7: Verify acceptance criteria
**Files:** (none)

- [ ] `FilterAll(tf)` returns all top-level tasks, file order preserved, subtasks nested
- [ ] `FilterToday(tf, now)` returns full "Today" section + overdue/today tasks from other sections
- [ ] `FilterDue(tf)` returns only tasks with non-zero Due, sorted ascending
- [ ] `FilterDueToday(tf, now)` returns tasks with due AND due <= now, sorted ascending
- [ ] Subtasks are never flattened, never filtered independently
- [ ] Empty input → non-nil `[]TaskItem{}`
- [ ] No "Today" section → `FilterToday` returns only overdue from other sections
- [ ] run full test suite — all tests pass

## Post-Completion

- No external systems need updating
- Manual checks: verify filter results match expected output from `testdata/tasklist.md`
- Next phase (Phase 3): task list formatting (text output, CLI-friendly output)
- No REST API wiring, no mcore changes in this phase
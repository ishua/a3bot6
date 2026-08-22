package taskparser

import (
	"sort"
	"time"
)

const (
	// todaySectionTitle is the case-sensitive section title treated specially
	// by FilterToday. Tasks in this section are always returned regardless of
	// due date.
	todaySectionTitle = "Today"
)

// FilterAll returns all top-level tasks from every section, in file order.
// Tasks from each section are appended in order they appear; subtasks stay
// nested inside their parent TaskItem and are never flattened. Returns a
// non-nil empty slice for a TaskFile with no sections or no tasks.
func FilterAll(tf *TaskFile) []TaskItem {
	result := make([]TaskItem, 0)
	if tf == nil {
		return result
	}
	for _, section := range tf.Sections {
		result = append(result, section.Tasks...)
	}
	return result
}

// sortByDue stably sorts tasks in ascending order by Due date. Tasks with a
// zero (unset) Due sort last; tasks sharing a Due keep their original relative
// order.
func sortByDue(tasks []TaskItem) {
	sort.SliceStable(tasks, func(i, j int) bool {
		di, dj := tasks[i].Due, tasks[j].Due
		// Zero (unset) Due always sorts last, even when both are non-zero.
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}
		return di.Before(dj)
	})
}

// startOfDay truncates t to the start of its day (midnight UTC).
func startOfDay(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

// dueOnOrBefore reports whether task carries a non-zero Due on or before day
// (i.e. due today or overdue).
func dueOnOrBefore(task TaskItem, day time.Time) bool {
	return !task.Due.IsZero() && !task.Due.After(day)
}

// collectDue returns the top-level tasks from all sections, skipping the
// section titled skipTitle (which is ignored when empty), for which due
// returns true. The result preserves file order and is non-nil when empty.
func collectDue(tf *TaskFile, skipTitle string, due func(TaskItem) bool) []TaskItem {
	result := make([]TaskItem, 0)
	for _, section := range tf.Sections {
		if skipTitle != "" && section.Title == skipTitle {
			continue
		}
		for _, task := range section.Tasks {
			if due(task) {
				result = append(result, task)
			}
		}
	}
	return result
}

// FilterDue returns all top-level tasks (from any section) that have a
// non-zero Due, sorted ascending by Due date. Tasks with no due date are
// excluded. Returns a non-nil empty slice when there is nothing to return.
func FilterDue(tf *TaskFile) []TaskItem {
	if tf == nil {
		return make([]TaskItem, 0)
	}
	result := collectDue(tf, "", func(task TaskItem) bool {
		return !task.Due.IsZero()
	})
	sortByDue(result)
	return result
}

// FilterToday returns the entire "Today" section (all its tasks, regardless of
// due date) followed by overdue/today tasks from other sections (non-zero Due
// and due <= now). "Today" is matched by exact case-sensitive title. File order
// is preserved within each group; no sorting. If no "Today" section exists,
// only overdue/today tasks from other sections are returned. Returns a non-nil
// empty slice when there is nothing to return.
func FilterToday(tf *TaskFile, now time.Time) []TaskItem {
	result := make([]TaskItem, 0)
	if tf == nil {
		return result
	}
	day := startOfDay(now)
	for _, section := range tf.Sections {
		if section.Title == todaySectionTitle {
			result = append(result, section.Tasks...)
		}
	}
	result = append(result, collectDue(tf, todaySectionTitle, func(task TaskItem) bool {
		return dueOnOrBefore(task, day)
	})...)
	return result
}

// FilterDueToday returns all top-level tasks (from any section) that have a
// non-zero Due and due <= now, sorted ascending by Due date. Tasks with no due
// date and tasks due after now are excluded. now is truncated to the day, so a
// task due today is included regardless of the current time of day. Returns a
// non-nil empty slice when there is nothing to return.
func FilterDueToday(tf *TaskFile, now time.Time) []TaskItem {
	if tf == nil {
		return make([]TaskItem, 0)
	}
	day := startOfDay(now)
	result := collectDue(tf, "", func(task TaskItem) bool {
		return dueOnOrBefore(task, day)
	})
	sortByDue(result)
	return result
}

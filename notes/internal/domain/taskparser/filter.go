package taskparser

import "sort"

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
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}
		return di.Before(dj)
	})
}

// FilterDue returns all top-level tasks (from any section) that have a
// non-zero Due, sorted ascending by Due date. Tasks with no due date are
// excluded. Returns a non-nil empty slice when there is nothing to return.
func FilterDue(tf *TaskFile) []TaskItem {
	result := make([]TaskItem, 0)
	if tf == nil {
		return result
	}
	for _, section := range tf.Sections {
		for _, task := range section.Tasks {
			if !task.Due.IsZero() {
				result = append(result, task)
			}
		}
	}
	sortByDue(result)
	return result
}

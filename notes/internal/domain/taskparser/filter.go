package taskparser

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

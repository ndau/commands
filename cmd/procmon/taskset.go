package main

// TaskSet maintains a simple Set structure for tasks; it's cleaner than
// direct manipulation of a map.
type TaskSet map[*Task]struct{}

// Add puts a task in the TaskSet
func (t *TaskSet) Add(tasks ...*Task) {
	for _, task := range tasks {
		(*t)[task] = struct{}{}
	}
}

// Has indicates whether a task is in the TaskSet
func (t *TaskSet) Has(task *Task) bool {
	_, ok := (*t)[task]
	return ok
}

// Delete removes a task from the TaskSet
func (t *TaskSet) Delete(task *Task) {
	delete(*t, task)
}

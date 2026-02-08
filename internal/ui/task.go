package ui

import (
	"fmt"
	"os"
)

type taskItem struct {
	label string
	value string
}

// TaskList collects all planned tasks upfront and prints each
// with a checkmark as it executes.
type TaskList struct {
	items []taskItem
}

// NewTaskList creates an empty TaskList.
func NewTaskList() *TaskList {
	return &TaskList{}
}

// Add appends a task to the list.
func (t *TaskList) Add(label, value string) {
	t.items = append(t.items, taskItem{label: label, value: value})
}

// MarkDone prints the task at index as completed: - [x] label value
func (t *TaskList) MarkDone(index int) {
	item := t.items[index]
	fmt.Fprintf(os.Stderr, "- %s %s %s\n", green.Sprint("[x]"), cyan.Sprint(item.label), item.value)
}

// MarkFailed prints the task at index as failed: - [ ] label value
func (t *TaskList) MarkFailed(index int) {
	item := t.items[index]
	fmt.Fprintf(os.Stderr, "- %s %s %s\n", red.Sprint("[ ]"), cyan.Sprint(item.label), item.value)
}

// FailRemaining prints all tasks from index onwards as failed.
func (t *TaskList) FailRemaining(from int) {
	for i := from; i < len(t.items); i++ {
		t.MarkFailed(i)
	}
}

// Len returns the number of tasks.
func (t *TaskList) Len() int {
	return len(t.items)
}

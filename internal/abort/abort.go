// Package abort abandons tasks without removing git artifacts.
package abort

import (
	"fmt"

	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/queue"
	"github.com/sunChenXuan/mergesiding/internal/store"
)

// Abort abandons a task: set ABORTED and dequeue. Does not remove worktrees or branches;
// callers must use cleanup after abort when disk cleanup is desired.
func Abort(p paths.Paths, taskSlug string) (*models.TaskRecord, error) {
	s := store.Store{Paths: p}
	task, err := s.Load(taskSlug)
	if err != nil {
		return nil, err
	}
	switch task.Status {
	case models.StatusActive, models.StatusReady, models.StatusAwaitingWriter,
		models.StatusBlocked, models.StatusBlockedPartial:
	default:
		return nil, fmt.Errorf("cannot abort task in status %s", task.Status)
	}
	if err := (queue.ReadyQueue{Paths: p}).Remove(taskSlug); err != nil {
		return nil, fmt.Errorf("dequeue before abort: %w", err)
	}
	msg := "aborted"
	if task.Error == nil {
		task.Error = &msg
	}
	task.Status = models.StatusAborted
	if err := s.Save(task); err != nil {
		return nil, err
	}
	return task, nil
}

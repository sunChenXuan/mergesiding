// Package ready marks tasks READY and enqueues them for integrate.
package ready

import (
	"fmt"

	"github.com/sunChenXuan/mergesiding/internal/gitops"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/queue"
	"github.com/sunChenXuan/mergesiding/internal/store"
)

// MarkReady marks a task ready when worktrees are clean and have commits ahead.
func MarkReady(p paths.Paths, taskSlug string) (*models.TaskRecord, error) {
	s := store.Store{Paths: p}
	task, err := s.Load(taskSlug)
	if err != nil {
		return nil, err
	}
	prevStatus := task.Status
	prevReadyAt := task.ReadyAt
	prevError := task.Error
	switch task.Status {
	case models.StatusActive, models.StatusAwaitingWriter, models.StatusBlocked:
	default:
		return nil, fmt.Errorf("cannot mark ready from status: %s", task.Status)
	}
	for _, rb := range task.Repos {
		clean, err := gitops.IsClean(rb.WorktreePath)
		if err != nil {
			return nil, err
		}
		if !clean {
			return nil, fmt.Errorf("worktree not clean: %s", rb.WorktreePath)
		}
		ahead, err := gitops.CommitsAhead(rb.Path, rb.IntegrationBranch, rb.Branch)
		if err != nil {
			return nil, err
		}
		if ahead < 1 {
			return nil, fmt.Errorf("no commits ahead of %s: %s", rb.IntegrationBranch, rb.Branch)
		}
	}
	now := models.NowUTC()
	task.Status = models.StatusReady
	task.ReadyAt = &now
	task.Error = nil
	if err := s.Save(task); err != nil {
		return nil, err
	}
	if err := (queue.ReadyQueue{Paths: p}).Enqueue(taskSlug); err != nil {
		task.Status = prevStatus
		task.ReadyAt = prevReadyAt
		task.Error = prevError
		if rbErr := s.Save(task); rbErr != nil {
			return nil, fmt.Errorf("enqueue failed: %v (also rollback save: %w)", err, rbErr)
		}
		return nil, fmt.Errorf("enqueue failed: %w", err)
	}
	return task, nil
}

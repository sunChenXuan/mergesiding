// Package cleanup removes worktrees and/or task branches for DONE/ABORTED tasks.
package cleanup

import (
	"fmt"

	"github.com/sunChenXuan/mergesiding/internal/gitops"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/store"
)

// CleanupTask optionally removes worktrees then branches.
func CleanupTask(task *models.TaskRecord, removeWorktree, deleteBranch bool) ([]string, error) {
	var actions []string
	if removeWorktree {
		for _, rb := range task.Repos {
			if err := gitops.RemoveWorktree(rb.Path, rb.WorktreePath); err != nil {
				return actions, err
			}
			actions = append(actions, "removed worktree "+rb.WorktreePath)
		}
	}
	if deleteBranch {
		for _, rb := range task.Repos {
			if err := gitops.DeleteBranch(rb.Path, rb.Branch); err != nil {
				return actions, err
			}
			actions = append(actions, "deleted branch "+rb.Branch+" in "+rb.Path)
		}
	}
	return actions, nil
}

// Run loads a DONE/ABORTED task and cleans git artifacts. Requires at least one flag.
func Run(p paths.Paths, slug string, removeWorktree, deleteBranch bool) ([]string, error) {
	if !removeWorktree && !deleteBranch {
		return nil, fmt.Errorf("no cleanup actions without flags; pass --remove-worktree and/or --delete-branch")
	}
	task, err := (store.Store{Paths: p}).Load(slug)
	if err != nil {
		return nil, err
	}
	if task.Status != models.StatusDone && task.Status != models.StatusAborted {
		return nil, fmt.Errorf("cannot cleanup task in status %s", task.Status)
	}
	return CleanupTask(task, removeWorktree, deleteBranch)
}

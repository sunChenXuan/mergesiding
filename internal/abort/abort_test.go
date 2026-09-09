package abort_test

import (
	"testing"

	"github.com/sunChenXuan/mergesiding/internal/abort"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/queue"
	"github.com/sunChenXuan/mergesiding/internal/store"
)

// TestAbortSetsAbortedWithoutCleanup verifies Abort only updates status and dequeues.
func TestAbortSetsAbortedWithoutCleanup(t *testing.T) {
	p := paths.Paths{Root: t.TempDir()}
	s := store.Store{Paths: p}
	task := &models.TaskRecord{
		Slug:   "t",
		Status: models.StatusActive,
		Repos: []models.RepoBinding{{
			Path:              "/repo",
			WorktreePath:      "/wt",
			Branch:            "task/t",
			IntegrationBranch: "main",
		}},
		MergedRepos: []string{},
	}
	if err := s.Save(task); err != nil {
		t.Fatal(err)
	}
	q := queue.ReadyQueue{Paths: p}
	if err := q.Enqueue("t"); err != nil {
		t.Fatal(err)
	}

	got, err := abort.Abort(p, "t")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != models.StatusAborted {
		t.Fatalf("status = %s, want aborted", got.Status)
	}
	if got.Repos[0].WorktreePath != "/wt" {
		t.Fatalf("worktree path changed: %+v", got.Repos)
	}
	reloaded, err := s.Load("t")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Status != models.StatusAborted {
		t.Fatalf("reloaded status = %s", reloaded.Status)
	}
	items, err := q.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("queue still has %v", items)
	}
}

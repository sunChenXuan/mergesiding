// Package start creates isolated task worktrees and ACTIVE task records.
package start

import (
	"fmt"
	"path/filepath"

	"github.com/sunChenXuan/mergesiding/internal/config"
	"github.com/sunChenXuan/mergesiding/internal/gitops"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/store"
	"github.com/sunChenXuan/mergesiding/internal/worktree"
)

// Options configures Start.
type Options struct {
	Slug           string
	Repos          []string
	SharedRoot     *string
	WriterID       *string
	BriefPath      *string
}

// Start creates task branches/worktrees for each repo and persists an ACTIVE task.
func Start(p paths.Paths, opt Options) (*models.TaskRecord, error) {
	if opt.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if len(opt.Repos) == 0 {
		return nil, fmt.Errorf("at least one repo is required")
	}
	s := store.Store{Paths: p}
	if s.Exists(opt.Slug) {
		return nil, fmt.Errorf("task %s already exists", opt.Slug)
	}
	multi := len(opt.Repos) > 1
	bindings := make([]models.RepoBinding, 0, len(opt.Repos))
	for _, repo := range opt.Repos {
		repoAbs, err := filepath.Abs(repo)
		if err != nil {
			return nil, err
		}
		cfg, err := config.Load(repoAbs)
		if err != nil {
			return nil, err
		}
		integration := cfg.IntegrationBranch
		if integration == "" {
			integration, err = gitops.CurrentBranch(repoAbs)
			if err != nil {
				return nil, err
			}
		}
		var cfgRoot *string
		if cfg.WorktreeRoot != "" {
			v := cfg.WorktreeRoot
			cfgRoot = &v
		}
		wt, err := worktree.ResolveWorktreePath(worktree.ResolveInput{
			RepoRoot:       repoAbs,
			Slug:           opt.Slug,
			SharedRoot:     opt.SharedRoot,
			RepoConfigRoot: cfgRoot,
			MultiRepo:      multi && opt.SharedRoot != nil,
		})
		if err != nil {
			return nil, err
		}
		branch := "task/" + opt.Slug
		if err := gitops.CreateTaskWorktree(repoAbs, wt, branch, integration); err != nil {
			return nil, err
		}
		bindings = append(bindings, models.RepoBinding{
			Path:              repoAbs,
			WorktreePath:      wt,
			Branch:            branch,
			IntegrationBranch: integration,
		})
	}
	now := models.NowUTC()
	task := &models.TaskRecord{
		Slug:        opt.Slug,
		Status:      models.StatusActive,
		WriterID:    opt.WriterID,
		BriefPath:   opt.BriefPath,
		Repos:       bindings,
		MergedRepos: []string{},
		CreatedAt:   &now,
	}
	if err := s.Save(task); err != nil {
		return nil, err
	}
	return task, nil
}

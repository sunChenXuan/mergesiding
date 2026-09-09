// Package status builds human and JSON status views for mergesiding tasks.
package status

import (
	"fmt"
	"strings"

	"github.com/sunChenXuan/mergesiding/internal/escalate"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/queue"
	"github.com/sunChenXuan/mergesiding/internal/store"
)

// Payload returns list summaries or one detail map for --json consumers.
func Payload(p paths.Paths, slug *string) (any, error) {
	s := store.Store{Paths: p}
	q, err := (queue.ReadyQueue{Paths: p}).List()
	if err != nil {
		return nil, err
	}
	if slug != nil {
		task, err := s.Load(*slug)
		if err != nil {
			return nil, err
		}
		return detailDict(p, task, q), nil
	}
	tasks, err := s.List()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, summaryDict(p, task, q))
	}
	return out, nil
}

// FormatLines returns human-readable status lines.
func FormatLines(p paths.Paths, slug *string) ([]string, error) {
	s := store.Store{Paths: p}
	q, err := (queue.ReadyQueue{Paths: p}).List()
	if err != nil {
		return nil, err
	}
	if slug != nil {
		task, err := s.Load(*slug)
		if err != nil {
			return nil, err
		}
		return formatDetail(p, task, q), nil
	}
	tasks, err := s.List()
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(tasks))
	for _, task := range tasks {
		lines = append(lines, formatListLine(p, task, q))
	}
	return lines, nil
}

func queueIndex(q []string, slug string) any {
	for i, s := range q {
		if s == slug {
			return i
		}
	}
	return nil
}

func summaryDict(p paths.Paths, task *models.TaskRecord, q []string) map[string]any {
	esc := escalate.PathFor(p, task.Slug)
	var escAny any
	if esc != "" {
		escAny = esc
	}
	return map[string]any{
		"slug":          task.Slug,
		"status":        string(task.Status),
		"queue_index":   queueIndex(q, task.Slug),
		"writer_id":     task.WriterID,
		"error":         task.Error,
		"escalate_path": escAny,
	}
}

func detailDict(p paths.Paths, task *models.TaskRecord, q []string) map[string]any {
	d := summaryDict(p, task, q)
	d["brief_path"] = task.BriefPath
	d["merged_repos"] = task.MergedRepos
	repos := make([]map[string]any, 0, len(task.Repos))
	for _, rb := range task.Repos {
		repos = append(repos, map[string]any{
			"path":               rb.Path,
			"worktree_path":      rb.WorktreePath,
			"branch":             rb.Branch,
			"integration_branch": rb.IntegrationBranch,
		})
	}
	d["repos"] = repos
	d["recovery_checklist"] = RecoveryChecklist(task)
	return d
}

func formatListLine(p paths.Paths, task *models.TaskRecord, q []string) string {
	qi := queueIndex(q, task.Slug)
	queueStr := "-"
	if qi != nil {
		queueStr = fmt.Sprintf("%v", qi)
	}
	writer := "-"
	if task.WriterID != nil {
		writer = *task.WriterID
	}
	err := "-"
	if task.Error != nil {
		err = truncate(*task.Error, 80)
	}
	esc := escalate.PathFor(p, task.Slug)
	if esc == "" {
		esc = "-"
	}
	return fmt.Sprintf("%s %s queue=%s writer=%s error=%s escalate=%s",
		task.Slug, task.Status, queueStr, writer, err, esc)
}

func formatDetail(p paths.Paths, task *models.TaskRecord, q []string) []string {
	lines := []string{formatListLine(p, task, q)}
	for _, rb := range task.Repos {
		lines = append(lines, fmt.Sprintf("  repo=%s wt=%s branch=%s integration=%s",
			rb.Path, rb.WorktreePath, rb.Branch, rb.IntegrationBranch))
	}
	if len(task.MergedRepos) > 0 {
		lines = append(lines, "  merged_repos="+strings.Join(task.MergedRepos, ","))
	}
	for _, c := range RecoveryChecklist(task) {
		lines = append(lines, "  - "+c)
	}
	return lines
}

// RecoveryChecklist returns short recovery hints for blocked statuses.
func RecoveryChecklist(task *models.TaskRecord) []string {
	switch task.Status {
	case models.StatusBlocked:
		return []string{
			"Inspect error and escalate markdown if present",
			"Fix verify/dirty-base issues in worktree or base",
			"Run mergesiding ready --slug " + task.Slug,
			"Scheduler re-runs mergesiding integrate",
		}
	case models.StatusBlockedPartial:
		return []string{
			"Review merged_repos for already-merged bases",
			"Do not re-ready solely for remaining repos",
			"Run mergesiding integrate --slug " + task.Slug,
		}
	case models.StatusAwaitingWriter:
		return []string{
			"Resolve rebase conflicts in the worktree only",
			"git add && git rebase --continue",
			"mergesiding ready --slug " + task.Slug,
			"Scheduler runs mergesiding integrate",
		}
	default:
		return nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Package integrate runs serial rebase → verify → merge under an exclusive lock.
package integrate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sunChenXuan/mergesiding/internal/config"
	"github.com/sunChenXuan/mergesiding/internal/escalate"
	"github.com/sunChenXuan/mergesiding/internal/gitops"
	"github.com/sunChenXuan/mergesiding/internal/lock"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/queue"
	"github.com/sunChenXuan/mergesiding/internal/store"
	"github.com/sunChenXuan/mergesiding/internal/verify"
)

// Outcome is the result of a single IntegrateOne invocation.
type Outcome struct {
	Slug         string
	FinalStatus  *models.TaskStatus
	Error        *string
	Empty        bool
	StopBatch    bool
	Message      string
	EscalatePath string
}

// ToDict returns stable fields for CLI --json / MCP consumers.
func (o Outcome) ToDict() map[string]any {
	var status any
	if o.FinalStatus != nil {
		status = string(*o.FinalStatus)
	}
	var err any
	if o.Error != nil {
		err = *o.Error
	}
	var msg any
	if o.Message != "" {
		msg = o.Message
	}
	var esc any
	if o.EscalatePath != "" {
		esc = o.EscalatePath
	}
	return map[string]any{
		"slug":          nullIfEmpty(o.Slug),
		"status":        status,
		"error":         err,
		"empty":         o.Empty,
		"stop_batch":    o.StopBatch,
		"message":       msg,
		"escalate_path": esc,
	}
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func stopBatch(st models.TaskStatus) bool {
	switch st {
	case models.StatusAwaitingWriter, models.StatusBlocked, models.StatusBlockedPartial:
		return true
	default:
		return false
	}
}

func outcomeFromTask(p paths.Paths, slug string, task *models.TaskRecord, message string) Outcome {
	st := task.Status
	esc := escalate.PathFor(p, slug)
	return Outcome{
		Slug:         slug,
		FinalStatus:  &st,
		Error:        task.Error,
		StopBatch:    stopBatch(st),
		Message:      message,
		EscalatePath: esc,
	}
}

func resolvedRepoPath(bindingPath string) string {
	abs, err := filepath.Abs(bindingPath)
	if err != nil {
		return bindingPath
	}
	return abs
}

func runPhaseA(p paths.Paths, s store.Store, q queue.ReadyQueue, slug string, task *models.TaskRecord) *Outcome {
	for _, binding := range task.Repos {
		base := binding.Path
		wt := binding.WorktreePath
		clean, err := gitops.IsClean(base)
		if err != nil || !clean {
			msg := fmt.Sprintf("base checkout dirty: %s", base)
			task.Status = models.StatusBlocked
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			o := outcomeFromTask(p, slug, task, "")
			return &o
		}
		if _, err := gitops.Run(base, "checkout", binding.IntegrationBranch); err != nil {
			msg := err.Error()
			task.Status = models.StatusBlocked
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			o := outcomeFromTask(p, slug, task, "")
			return &o
		}
		upstreamOut, err := gitops.Run(base, "rev-parse", binding.IntegrationBranch)
		if err != nil {
			msg := err.Error()
			task.Status = models.StatusBlocked
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			o := outcomeFromTask(p, slug, task, "")
			return &o
		}
		result := gitops.RebaseOnto(wt, strings.TrimSpace(upstreamOut))
		if !result.OK {
			if len(result.Conflicts) > 0 {
				esc, _ := escalate.WriteConflictEscalation(p, task, binding, result.Conflicts)
				escalate.TryResumeWriter(task.WriterID)
				msg := fmt.Sprintf("conflicts: %v; see %s", result.Conflicts, esc)
				task.Status = models.StatusAwaitingWriter
				task.Error = &msg
				_ = s.Save(task)
				_ = q.Remove(slug)
				o := outcomeFromTask(p, slug, task, "")
				return &o
			}
			msg := result.Message
			task.Status = models.StatusBlocked
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			o := outcomeFromTask(p, slug, task, "")
			return &o
		}
		cfg, _ := config.Load(base)
		vr := verify.Run(wt, cfg.Verify)
		if !vr.OK {
			esc, _ := escalate.WriteVerifyEscalation(p, task, binding, vr.Log)
			escalate.TryResumeWriter(task.WriterID)
			msg := fmt.Sprintf("verify failed; see %s", esc)
			task.Status = models.StatusBlocked
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			o := outcomeFromTask(p, slug, task, "")
			return &o
		}
	}
	return nil
}

func runPhaseB(p paths.Paths, s store.Store, q queue.ReadyQueue, slug string, task *models.TaskRecord, skipMerged bool) Outcome {
	merged := []string{}
	if skipMerged {
		for _, m := range task.MergedRepos {
			merged = append(merged, resolvedRepoPath(m))
		}
	}
	for _, binding := range task.Repos {
		resolved := resolvedRepoPath(binding.Path)
		if skipMerged {
			skip := false
			for _, m := range merged {
				if m == resolved {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		base := binding.Path
		if _, err := gitops.Run(base, "checkout", binding.IntegrationBranch); err != nil {
			task.Status = models.StatusBlockedPartial
			task.MergedRepos = merged
			msg := fmt.Sprintf("merged=%v; error=%v", merged, err)
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			return outcomeFromTask(p, slug, task, "")
		}
		if err := gitops.MergeFFOrNoFF(base, binding.Branch); err != nil {
			task.Status = models.StatusBlockedPartial
			task.MergedRepos = merged
			msg := fmt.Sprintf("merged=%v; error=%v", merged, err)
			task.Error = &msg
			_ = s.Save(task)
			_ = q.Remove(slug)
			return outcomeFromTask(p, slug, task, "")
		}
		merged = append(merged, resolved)
	}
	task.Status = models.StatusDone
	task.Error = nil
	task.MergedRepos = []string{}
	_ = s.Save(task)
	_ = q.Remove(slug)
	return outcomeFromTask(p, slug, task, "")
}

// IntegrateOne integrates one task: peek queue or target a slug for resume/recovery.
func IntegrateOne(p paths.Paths, slug *string) Outcome {
	s := store.Store{Paths: p}
	q := queue.ReadyQueue{Paths: p}
	lk := &lock.IntegrateLock{Paths: p}
	if err := lk.Acquire(false, 0); err != nil {
		msg := "lock held"
		return Outcome{Error: &msg, Message: msg, StopBatch: true}
	}
	defer lk.Release()

	explicit := slug != nil
	var target string
	if explicit {
		target = *slug
	} else {
		peek, ok, err := q.Peek()
		if err != nil {
			msg := err.Error()
			return Outcome{Error: &msg, Message: msg, StopBatch: true}
		}
		if !ok {
			return Outcome{Empty: true}
		}
		target = peek
	}

	task, err := s.Load(target)
	if err != nil {
		msg := err.Error()
		return Outcome{Slug: target, Error: &msg, Message: msg, StopBatch: true}
	}

	if task.Status == models.StatusIntegrating {
		msg := "crash recovery: task was integrating; manual review required"
		task.Status = models.StatusBlocked
		task.Error = &msg
		_ = s.Save(task)
		_ = q.Remove(target)
		return outcomeFromTask(p, target, task, "")
	}

	if explicit {
		if task.Status == models.StatusBlockedPartial {
			return runPhaseB(p, s, q, target, task, true)
		}
		if task.Status != models.StatusReady {
			return outcomeFromTask(p, target, task, fmt.Sprintf("cannot integrate task in status %s", task.Status))
		}
	} else if task.Status != models.StatusReady {
		_ = q.Remove(target)
		return outcomeFromTask(p, target, task, "")
	}

	task.Status = models.StatusIntegrating
	_ = s.Save(task)

	if phaseA := runPhaseA(p, s, q, target, task); phaseA != nil {
		return *phaseA
	}
	return runPhaseB(p, s, q, target, task, false)
}

// IntegrateAll processes the queue until empty or stop_batch.
func IntegrateAll(p paths.Paths) ([]Outcome, map[string]any) {
	var results []Outcome
	for {
		out := IntegrateOne(p, nil)
		if out.Empty {
			break
		}
		results = append(results, out)
		if out.StopBatch {
			break
		}
	}
	summary := map[string]any{
		"count": len(results),
		"stopped": len(results) > 0 && results[len(results)-1].StopBatch,
	}
	return results, summary
}

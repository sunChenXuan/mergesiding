// Package models defines task status and persisted task records for mergesiding.
package models

import "time"

// TaskStatus is the lifecycle state of a mergesiding task.
type TaskStatus string

const (
	StatusActive          TaskStatus = "active"
	StatusReady           TaskStatus = "ready"
	StatusIntegrating     TaskStatus = "integrating"
	StatusAwaitingWriter  TaskStatus = "awaiting_writer"
	StatusBlocked         TaskStatus = "blocked"
	StatusBlockedPartial  TaskStatus = "blocked_partial"
	StatusDone            TaskStatus = "done"
	StatusAborted         TaskStatus = "aborted"
)

// RepoBinding links a human base checkout to its task worktree and branch.
type RepoBinding struct {
	Path              string `json:"path"`
	WorktreePath      string `json:"worktree_path"`
	Branch            string `json:"branch"`
	IntegrationBranch string `json:"integration_branch"`
}

// TaskRecord is the on-disk JSON task document under MERGESIDING_HOME/tasks.
type TaskRecord struct {
	Slug        string       `json:"slug"`
	Status      TaskStatus   `json:"status"`
	WriterID    *string      `json:"writer_id"`
	BriefPath   *string      `json:"brief_path"`
	Repos       []RepoBinding `json:"repos"`
	MergedRepos []string     `json:"merged_repos"`
	CreatedAt   *string      `json:"created_at"`
	ReadyAt     *string      `json:"ready_at"`
	Error       *string      `json:"error"`
}

// NowUTC returns an RFC3339 UTC timestamp string for created_at / ready_at.
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

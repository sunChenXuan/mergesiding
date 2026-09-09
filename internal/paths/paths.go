// Package paths resolves mergesiding home directories for tasks, queue, locks, and escalate.
package paths

import (
	"os"
	"path/filepath"
)

// EnvHome is the preferred environment override for mergesiding metadata root.
const EnvHome = "MERGESIDING_HOME"

// EnvHomeLegacy is accepted for migration from agent-git.
const EnvHomeLegacy = "AGENT_GIT_HOME"

// Paths holds the mergesiding home layout.
type Paths struct {
	Root string
}

// Default resolves home: MERGESIDING_HOME, else AGENT_GIT_HOME, else ~/.mergesiding.
func Default() Paths {
	if v := os.Getenv(EnvHome); v != "" {
		return Paths{Root: filepath.Clean(v)}
	}
	if v := os.Getenv(EnvHomeLegacy); v != "" {
		return Paths{Root: filepath.Clean(v)}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{Root: ".mergesiding"}
	}
	return Paths{Root: filepath.Join(home, ".mergesiding")}
}

// TasksDir returns the directory of per-slug task JSON files.
func (p Paths) TasksDir() string { return filepath.Join(p.Root, "tasks") }

// LocksDir returns the lock directory.
func (p Paths) LocksDir() string { return filepath.Join(p.Root, "locks") }

// QueueDir returns the queue directory.
func (p Paths) QueueDir() string { return filepath.Join(p.Root, "queue") }

// EscalateDir returns the escalation markdown directory.
func (p Paths) EscalateDir() string { return filepath.Join(p.Root, "escalate") }

// LogsDir returns the logs directory.
func (p Paths) LogsDir() string { return filepath.Join(p.Root, "logs") }

// IntegrateLock returns the integrate lock file path.
func (p Paths) IntegrateLock() string { return filepath.Join(p.LocksDir(), "integrate.lock") }

// ReadyQueueLock returns the lock file guarding ready-queue read-modify-write.
func (p Paths) ReadyQueueLock() string {
	return filepath.Join(p.LocksDir(), "ready-queue.lock")
}

// ReadyQueue returns the ready queue JSON path.
func (p Paths) ReadyQueue() string { return filepath.Join(p.QueueDir(), "ready.json") }

// TaskFile returns the task JSON path for a slug.
func (p Paths) TaskFile(slug string) string {
	return filepath.Join(p.TasksDir(), slug+".json")
}

// Ensure creates home subdirectories if missing.
func (p Paths) Ensure() error {
	for _, d := range []string{p.TasksDir(), p.LocksDir(), p.QueueDir(), p.EscalateDir(), p.LogsDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

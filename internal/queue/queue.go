// Package queue implements the FIFO ready queue persisted as JSON.
package queue

import (
	"encoding/json"
	"os"
	"time"

	"github.com/sunChenXuan/mergesiding/internal/lock"
	"github.com/sunChenXuan/mergesiding/internal/paths"
)

// ReadyQueue is a FIFO list of task slugs waiting for integrate.
type ReadyQueue struct {
	Paths paths.Paths
}

func (q ReadyQueue) withLock(fn func() error) error {
	return lock.HoldExclusive(q.Paths.ReadyQueueLock(), true, 30*time.Second, fn)
}

func (q ReadyQueue) read() ([]string, error) {
	if err := q.Paths.Ensure(); err != nil {
		return nil, err
	}
	path := q.Paths.ReadyQueue()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []string{}
	}
	return items, nil
}

func (q ReadyQueue) write(items []string) error {
	if err := q.Paths.Ensure(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	path := q.Paths.ReadyQueue()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	_ = os.Remove(path)
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Enqueue appends slug if not already present (idempotent).
func (q ReadyQueue) Enqueue(slug string) error {
	return q.withLock(func() error {
		items, err := q.read()
		if err != nil {
			return err
		}
		for _, s := range items {
			if s == slug {
				return nil
			}
		}
		items = append(items, slug)
		return q.write(items)
	})
}

// List returns all queued slugs in FIFO order.
func (q ReadyQueue) List() ([]string, error) {
	var items []string
	err := q.withLock(func() error {
		var err error
		items, err = q.read()
		return err
	})
	return items, err
}

// Peek returns the next slug without removing it.
func (q ReadyQueue) Peek() (string, bool, error) {
	var peek string
	var ok bool
	err := q.withLock(func() error {
		items, err := q.read()
		if err != nil {
			return err
		}
		if len(items) == 0 {
			ok = false
			return nil
		}
		peek = items[0]
		ok = true
		return nil
	})
	return peek, ok, err
}

// Remove deletes a slug from the queue.
func (q ReadyQueue) Remove(slug string) error {
	return q.withLock(func() error {
		items, err := q.read()
		if err != nil {
			return err
		}
		out := make([]string, 0, len(items))
		for _, s := range items {
			if s != slug {
				out = append(out, s)
			}
		}
		return q.write(out)
	})
}

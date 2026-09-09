// Package queue implements the FIFO ready queue persisted as JSON.
package queue

import (
	"encoding/json"
	"os"

	"github.com/sunChenXuan/mergesiding/internal/paths"
)

// ReadyQueue is a FIFO list of task slugs waiting for integrate.
type ReadyQueue struct {
	Paths paths.Paths
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
	return os.WriteFile(q.Paths.ReadyQueue(), append(data, '\n'), 0o644)
}

// Enqueue appends slug if not already present (idempotent).
func (q ReadyQueue) Enqueue(slug string) error {
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
}

// List returns all queued slugs in FIFO order.
func (q ReadyQueue) List() ([]string, error) {
	return q.read()
}

// Peek returns the next slug without removing it.
func (q ReadyQueue) Peek() (string, bool, error) {
	items, err := q.read()
	if err != nil {
		return "", false, err
	}
	if len(items) == 0 {
		return "", false, nil
	}
	return items[0], true, nil
}

// Remove deletes a slug from the queue.
func (q ReadyQueue) Remove(slug string) error {
	items, err := q.read()
	if err != nil {
		return err
	}
	out := items[:0]
	for _, s := range items {
		if s != slug {
			out = append(out, s)
		}
	}
	return q.write(out)
}

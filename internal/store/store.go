// Package store persists TaskRecord JSON files under mergesiding home.
package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
)

// ErrNotFound is returned when a task slug has no JSON file.
var ErrNotFound = errors.New("task not found")

// Store reads and writes task JSON documents.
type Store struct {
	Paths paths.Paths
}

// Exists reports whether a task file exists for slug.
func (s Store) Exists(slug string) bool {
	_, err := os.Stat(s.Paths.TaskFile(slug))
	return err == nil
}

// Save writes a task record atomically-ish via temp + rename.
func (s Store) Save(task *models.TaskRecord) error {
	if err := s.Paths.Ensure(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	path := s.Paths.TaskFile(task.Slug)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads a task by slug.
func (s Store) Load(slug string) (*models.TaskRecord, error) {
	data, err := os.ReadFile(s.Paths.TaskFile(slug))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var task models.TaskRecord
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// List returns all tasks found under tasks/.
func (s Store) List() ([]*models.TaskRecord, error) {
	if err := s.Paths.Ensure(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.Paths.TasksDir())
	if err != nil {
		return nil, err
	}
	out := make([]*models.TaskRecord, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".json")
		task, err := s.Load(slug)
		if err != nil {
			return nil, err
		}
		out = append(out, task)
	}
	_ = filepath.Separator // keep filepath imported for Windows clarity
	return out, nil
}

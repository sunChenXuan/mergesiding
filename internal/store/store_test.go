package store

import (
	"testing"

	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	p := paths.Paths{Root: t.TempDir()}
	s := Store{Paths: p}
	writer := "w1"
	task := &models.TaskRecord{
		Slug:     "foo",
		Status:   models.StatusActive,
		WriterID: &writer,
		Repos: []models.RepoBinding{{
			Path:              "/repo",
			WorktreePath:      "/wt",
			Branch:            "task/foo",
			IntegrationBranch: "main",
		}},
		MergedRepos: []string{},
	}
	if err := s.Save(task); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("foo")
	if err != nil {
		t.Fatal(err)
	}
	if got.Slug != "foo" || got.Status != models.StatusActive {
		t.Fatalf("got %+v", got)
	}
	if !s.Exists("foo") {
		t.Fatal("expected exists")
	}
}

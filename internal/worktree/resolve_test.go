package worktree

import (
	"path/filepath"
	"testing"
)

func TestDefaultSiblingLayout(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "projects", "my-app")
	if err := osMkdirAll(repo); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveWorktreePath(ResolveInput{RepoRoot: repo, Slug: "feat-foo"})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(filepath.Dir(repo), ".agent-git-worktrees", "my-app", "feat-foo")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestAbsoluteSharedRoot(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	if err := osMkdirAll(repo); err != nil {
		t.Fatal(err)
	}
	shared := t.TempDir()
	got, err := ResolveWorktreePath(ResolveInput{
		RepoRoot:   repo,
		Slug:       "s",
		SharedRoot: &shared,
		MultiRepo:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(shared, "repo", "s")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRelativeRepoConfigRoot(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "my-app")
	if err := osMkdirAll(repo); err != nil {
		t.Fatal(err)
	}
	rel := "../.agent-git-worktrees/my-app"
	got, err := ResolveWorktreePath(ResolveInput{
		RepoRoot:       repo,
		Slug:           "x",
		RepoConfigRoot: &rel,
	})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(filepath.Join(repo, rel, "x"))
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRejectsHostileSlug(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "projects", "my-app")
	if err := osMkdirAll(repo); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveWorktreePath(ResolveInput{RepoRoot: repo, Slug: "../escape"}); err == nil {
		t.Fatal("expected hostile slug to fail")
	}
}

func TestMultiRepoEnvNestsByRepoName(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "my-app")
	if err := osMkdirAll(repo); err != nil {
		t.Fatal(err)
	}
	shared := t.TempDir()
	t.Setenv(EnvWorktreeRoot, shared)
	t.Setenv(EnvWorktreeRootLegacy, "")
	got, err := ResolveWorktreePath(ResolveInput{
		RepoRoot:  repo,
		Slug:      "feat",
		MultiRepo: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(shared, "my-app", "feat")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func osMkdirAll(p string) error {
	return mkdirAll(p)
}

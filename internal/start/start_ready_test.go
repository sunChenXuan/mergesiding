package start_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sunChenXuan/mergesiding/internal/gitops"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/ready"
	"github.com/sunChenXuan/mergesiding/internal/start"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "my-app")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := gitops.Run(dir, "init", "-b", "main"); err != nil {
		t.Fatal(err)
	}
	_, _ = gitops.Run(dir, "config", "user.email", "t@example.com")
	_, _ = gitops.Run(dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _ = gitops.Run(dir, "add", "README.md")
	_, _ = gitops.Run(dir, "commit", "-m", "init")
	return dir
}

func TestStartSiblingAndReady(t *testing.T) {
	repo := initRepo(t)
	home := paths.Paths{Root: t.TempDir()}
	task, err := start.Start(home, start.Options{Slug: "feat", Repos: []string{repo}})
	if err != nil {
		t.Fatal(err)
	}
	if len(task.Repos) != 1 {
		t.Fatal(task.Repos)
	}
	wt := task.Repos[0].WorktreePath
	wantParent := filepath.Join(filepath.Dir(repo), ".agent-git-worktrees", "my-app", "feat")
	if wt != wantParent {
		t.Fatalf("worktree=%q want %q", wt, wantParent)
	}
	if err := os.WriteFile(filepath.Join(wt, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _ = gitops.Run(wt, "add", "f.txt")
	_, _ = gitops.Run(wt, "commit", "-m", "work")
	if _, err := ready.MarkReady(home, "feat"); err != nil {
		t.Fatal(err)
	}
}

package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := Run(dir, "init", "-b", "main"); err != nil {
		t.Fatal(err)
	}
	_, _ = Run(dir, "config", "user.email", "t@example.com")
	_, _ = Run(dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(dir, "add", "README.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(dir, "commit", "-m", "init"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCreateTaskWorktreeAndClean(t *testing.T) {
	repo := initRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	if err := CreateTaskWorktree(repo, wt, "task/demo", "main"); err != nil {
		t.Fatal(err)
	}
	clean, err := IsClean(wt)
	if err != nil || !clean {
		t.Fatalf("clean=%v err=%v", clean, err)
	}
	br, err := CurrentBranch(wt)
	if err != nil || br != "task/demo" {
		t.Fatalf("branch=%q err=%v", br, err)
	}
}

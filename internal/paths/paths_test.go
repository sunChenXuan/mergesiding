package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultHomeUsesEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, dir)
	t.Setenv(EnvHomeLegacy, "")
	p := Default()
	if p.Root != filepath.Clean(dir) {
		t.Fatalf("Root = %q want %q", p.Root, dir)
	}
}

func TestDefaultFallsBackToLegacyEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, "")
	t.Setenv(EnvHomeLegacy, dir)
	p := Default()
	if p.Root != filepath.Clean(dir) {
		t.Fatalf("Root = %q want %q", p.Root, dir)
	}
}

func TestEnsureCreatesDirs(t *testing.T) {
	root := t.TempDir()
	p := Paths{Root: root}
	if err := p.Ensure(); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{p.TasksDir(), p.LocksDir(), p.QueueDir(), p.EscalateDir(), p.LogsDir()} {
		if st, err := os.Stat(d); err != nil || !st.IsDir() {
			t.Fatalf("missing dir %s: %v", d, err)
		}
	}
}

package verify

import "testing"

func TestRunEmptyOK(t *testing.T) {
	r := Run(t.TempDir(), nil)
	if !r.OK {
		t.Fatal(r.Log)
	}
}

func TestRunFailingCommand(t *testing.T) {
	r := Run(t.TempDir(), []string{"exit 1"})
	if r.OK {
		t.Fatal("expected failure")
	}
}

package lock

import (
	"testing"
	"time"

	"github.com/sunChenXuan/mergesiding/internal/paths"
)

func TestLockExclusive(t *testing.T) {
	p := paths.Paths{Root: t.TempDir()}
	a := &IntegrateLock{Paths: p}
	b := &IntegrateLock{Paths: p}
	if err := a.Acquire(false, 0); err != nil {
		t.Fatal(err)
	}
	defer a.Release()
	if err := b.Acquire(false, 0); err == nil {
		t.Fatal("expected second acquire to fail")
	}
	if err := a.Release(); err != nil {
		t.Fatal(err)
	}
	if err := b.Acquire(true, time.Second); err != nil {
		t.Fatal(err)
	}
	_ = b.Release()
}

package queue

import (
	"testing"

	"github.com/sunChenXuan/mergesiding/internal/paths"
)

func TestEnqueuePeekRemove(t *testing.T) {
	q := ReadyQueue{Paths: paths.Paths{Root: t.TempDir()}}
	if err := q.Enqueue("a"); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue("a"); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue("b"); err != nil {
		t.Fatal(err)
	}
	items, err := q.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0] != "a" || items[1] != "b" {
		t.Fatalf("list = %v", items)
	}
	peek, ok, err := q.Peek()
	if err != nil || !ok || peek != "a" {
		t.Fatalf("peek = %q ok=%v err=%v", peek, ok, err)
	}
	if err := q.Remove("a"); err != nil {
		t.Fatal(err)
	}
	peek, ok, err = q.Peek()
	if err != nil || !ok || peek != "b" {
		t.Fatalf("peek2 = %q ok=%v err=%v", peek, ok, err)
	}
}

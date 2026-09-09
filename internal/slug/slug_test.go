package slug

import "testing"

func TestValidateAcceptsSafe(t *testing.T) {
	for _, s := range []string{"feat", "feat-foo", "a1", "Feat_2.3"} {
		if err := Validate(s); err != nil {
			t.Fatalf("%q: %v", s, err)
		}
	}
}

func TestValidateRejectsTraversal(t *testing.T) {
	for _, s := range []string{"", "..", "../x", "a/b", `a\b`, "a/../b", "feat/../../evil"} {
		if err := Validate(s); err == nil {
			t.Fatalf("expected reject %q", s)
		}
	}
}

func TestValidateWriterID(t *testing.T) {
	if err := ValidateWriterID("writer-1"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateWriterID(`x"; rm -rf /; #`); err == nil {
		t.Fatal("expected reject shell metacharacters")
	}
}

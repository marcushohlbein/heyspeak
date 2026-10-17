package holdshortcut

import "testing"

func TestValidateAccelerator(t *testing.T) {
	if err := ValidateAccelerator(DefaultAccelerator); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAccelerator(""); err == nil {
		t.Fatal("expected error for empty")
	}
	if err := ValidateAccelerator("NotAReal+Chord"); err == nil {
		t.Fatal("expected error for invalid")
	}
}

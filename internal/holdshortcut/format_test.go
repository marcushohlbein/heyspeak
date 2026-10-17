package holdshortcut

import "testing"

func TestFormatAccelerator(t *testing.T) {
	got := FormatAccelerator(true, true, false, false, "Space")
	if got != "Ctrl+Shift+Space" {
		t.Fatalf("got %q", got)
	}
	if err := ValidateAccelerator(got); err != nil {
		t.Fatal(err)
	}
}

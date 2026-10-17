package holdshortcut

import "testing"

func TestParseChord(t *testing.T) {
	c, err := parseChord("Ctrl+Shift+Space")
	if err != nil {
		t.Fatal(err)
	}
	if !c.ctrl || !c.shift || c.keyCode != 0x31 {
		t.Fatalf("unexpected chord: %+v", c)
	}
}

func TestParseChordRequiresModifier(t *testing.T) {
	if _, err := parseChord("Space"); err == nil {
		t.Fatal("expected error for accelerator without modifier")
	}
}

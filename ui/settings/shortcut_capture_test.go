package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/holdshortcut"
)

func TestAcceleratorFromKeyDown(t *testing.T) {
	acc, ok := acceleratorFromKeyDown(ui.InputEvent{
		Kind: ui.InputKeyDown,
		Key:  ui.KeySpace,
		Mods: ui.Ctrl | ui.Shift,
	})
	if !ok || acc != holdshortcut.DefaultAccelerator {
		t.Fatalf("got %q ok=%v", acc, ok)
	}
	_, ok = acceleratorFromKeyDown(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeySpace, Mods: 0})
	if ok {
		t.Fatal("space without modifiers should not capture")
	}
}

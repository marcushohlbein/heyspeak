//go:build darwin

package paste

import (
	"os"
	"testing"

	"github.com/egoist/mygo"
)

// TestMain runs the mygo application loop on the main thread: mygo.Clipboard
// routes every pasteboard access through the running loop, so the clipboard
// tests below touch the live macOS pasteboard from a real (accessory) app.

func TestMain(m *testing.M) {
	code := make(chan int, 1)
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.WhenReady(func() {
		go func() {
			code <- m.Run()
			mygo.App.Quit()
		}()
	})
	if err := mygo.App.Run(); err != nil {
		os.Exit(1)
	}
	os.Exit(<-code)
}

func TestRestorePlainClipboard(t *testing.T) {
	// (a) clipboard still holds the inserted text: prior content is restored.
	writePlainClipboard("inserted")
	restorePlainClipboard("inserted", "prior")
	if got := readPlainClipboard(); got != "prior" {
		t.Fatalf("clipboard %q, want %q", got, "prior")
	}

	// (b) empty prior content: clipboard is cleared.
	writePlainClipboard("inserted")
	restorePlainClipboard("inserted", "")
	if got := readPlainClipboard(); got != "" {
		t.Fatalf("clipboard %q, want cleared", got)
	}

	// (c) clipboard no longer holds the inserted text: untouched.
	writePlainClipboard("unrelated")
	restorePlainClipboard("inserted", "prior")
	if got := readPlainClipboard(); got != "unrelated" {
		t.Fatalf("clipboard %q, want %q left untouched", got, "unrelated")
	}
}

func TestUndoTranscriptClipboard(t *testing.T) {
	writePlainClipboard("image-stand-in")
	if snapshotPasteboard() < 1 {
		t.Fatal("snapshot captured nothing")
	}
	writePlainClipboard("transcript")
	undoTranscriptClipboard("transcript", "image-stand-in", 1)
	if got := readPlainClipboard(); got != "image-stand-in" {
		t.Fatalf("clipboard %q, want prior contents", got)
	}

	writePlainClipboard("prior")
	snapshotPasteboard()
	writePlainClipboard("transcript")
	writePlainClipboard("user-replaced")
	undoTranscriptClipboard("transcript", "prior", 1)
	if got := readPlainClipboard(); got != "user-replaced" {
		t.Fatalf("clipboard %q, want the replacement left untouched", got)
	}
}

func TestPasteboardSnapshotRoundtrip(t *testing.T) {
	writePlainClipboard("snapshot-source")
	if n := snapshotPasteboard(); n < 1 {
		t.Fatalf("snapshot captured %d items, want >= 1", n)
	}
	writePlainClipboard("overwritten")
	restorePasteboardSnapshot()
	if got := readPlainClipboard(); got != "snapshot-source" {
		t.Fatalf("clipboard %q after restore, want %q", got, "snapshot-source")
	}
}

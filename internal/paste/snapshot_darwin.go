//go:build darwin

package paste

/*
#cgo darwin LDFLAGS: -framework AppKit
#include "snapshot.h"
*/
import "C"

// snapshotPasteboard copies the full system pasteboard (every item and type,
// including images and files) and returns the number of items captured.
func snapshotPasteboard() int {
	return int(C.speek_pasteboard_snapshot())
}

// restorePasteboardSnapshot writes the last snapshot back, replacing the
// current pasteboard contents.
func restorePasteboardSnapshot() {
	C.speek_pasteboard_restore()
}

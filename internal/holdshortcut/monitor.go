package holdshortcut

import "errors"

// Listener receives press and release for a held chord (not toggle).
type Listener interface {
	OnPress()
	OnRelease()
}

// Monitor watches a global keyboard chord and reports hold duration.
type Monitor interface {
	Start(accelerator string, listener Listener) error
	Stop() error
}

// ErrUnsupported is returned on platforms without an implementation.
var ErrUnsupported = errors.New("holdshortcut: unsupported on this platform")

// New returns a platform monitor.
func New() Monitor {
	return newMonitor()
}

// DefaultAccelerator is the spike default chord.
const DefaultAccelerator = "Ctrl+Shift+Space"

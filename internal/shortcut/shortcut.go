// Package shortcut registers global hold-to-talk shortcuts and forwards raw
// press/release events.
//
// HoldMonitor.Start returns before the OS event tap is actually registered
// (registration happens on a locked OS thread). A registration failure is
// logged and clears the thread so Start can be retried; callers that need
// synchronous confirmation should validate the accelerator up front with
// holdshortcut.ValidateAccelerator.
package shortcut

// Events are push-to-talk press and release notifications.
type Events struct {
	OnPress   func()
	OnRelease func()
}

// Monitor watches a global hold-to-talk chord (key down and key up).
type Monitor interface {
	Start(accelerator string, events Events) error
	Stop() error
}

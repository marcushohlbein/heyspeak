package shortcut

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/speek/speek/internal/holdshortcut"
)

// HoldMonitor runs the Phase 1 CGEvent tap adapter on a locked OS thread.
type HoldMonitor struct {
	mu      sync.Mutex
	monitor holdshortcut.Monitor
	thread  *holdThread
}

type holdThread struct {
	monitor holdshortcut.Monitor
	acc     string
	events  Events
}

func (h *holdThread) OnPress() {
	if h.events.OnPress != nil {
		h.events.OnPress()
	}
}

func (h *holdThread) OnRelease() {
	if h.events.OnRelease != nil {
		h.events.OnRelease()
	}
}

// NewHoldMonitor returns a platform hold-to-talk monitor.
func NewHoldMonitor() *HoldMonitor {
	return &HoldMonitor{monitor: holdshortcut.New()}
}

// Start registers the accelerator and begins listening.
func (h *HoldMonitor) Start(accelerator string, events Events) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.thread != nil {
		return fmt.Errorf("shortcut: already started")
	}
	if runtime.GOOS != "darwin" {
		return holdshortcut.ErrUnsupported
	}
	t := &holdThread{monitor: h.monitor, acc: accelerator, events: events}
	h.thread = t
	go func() {
		runtime.LockOSThread()
		if err := t.monitor.Start(t.acc, t); err != nil {
			// Start runs asynchronously on the locked OS thread, so the error
			// cannot travel back to the caller; surface it in the log and reset
			// the thread so StartShortcut can be retried.
			log.Printf("shortcut start: %v", err)
			h.mu.Lock()
			h.thread = nil
			h.mu.Unlock()
			return
		}
		_ = holdshortcut.Run(t.monitor)
	}()
	return nil
}

// Stop tears down the event tap.
func (h *HoldMonitor) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.thread == nil {
		return nil
	}
	err := h.monitor.Stop()
	h.thread = nil
	return err
}

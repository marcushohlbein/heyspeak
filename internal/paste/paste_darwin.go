//go:build darwin

package paste

import (
	"sync"
	"time"

	"github.com/ebitengine/purego"
)

const (
	kVK_ANSI_V              = 0x09
	kCGSessionEventTap      = 1
	kCGEventFlagMaskCommand = 1 << 20
)

var (
	cgEventCreateKeyboardEvent func(source uintptr, keyCode uint16, keyDown bool) uintptr
	cgEventSetFlags            func(event uintptr, flags uint64)
	cgEventPost                func(tap int, event uintptr)
)

var pasteInit sync.Once
var pasteInitErr error

func initPaste() error {
	pasteInit.Do(func() {
		cg, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_GLOBAL)
		if err != nil {
			pasteInitErr = err
			return
		}
		purego.RegisterLibFunc(&cgEventCreateKeyboardEvent, cg, "CGEventCreateKeyboardEvent")
		purego.RegisterLibFunc(&cgEventSetFlags, cg, "CGEventSetFlags")
		purego.RegisterLibFunc(&cgEventPost, cg, "CGEventPost")
	})
	return pasteInitErr
}

func paste() error {
	if err := initPaste(); err != nil {
		return err
	}
	// Let the pasteboard settle before posting the shortcut.
	time.Sleep(40 * time.Millisecond)

	down := cgEventCreateKeyboardEvent(0, kVK_ANSI_V, true)
	up := cgEventCreateKeyboardEvent(0, kVK_ANSI_V, false)
	if down == 0 || up == 0 {
		return ErrPostFailed
	}
	cgEventSetFlags(down, kCGEventFlagMaskCommand)
	cgEventSetFlags(up, kCGEventFlagMaskCommand)

	// Session tap delivers to the foreground app; HID tap often does not.
	cgEventPost(kCGSessionEventTap, down)
	time.Sleep(30 * time.Millisecond)
	cgEventPost(kCGSessionEventTap, up)
	return nil
}

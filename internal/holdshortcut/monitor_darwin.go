//go:build darwin

package holdshortcut

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	kCGEventKeyDown      = 10
	kCGEventKeyUp        = 11
	kCGEventFlagsChanged = 12

	kCGHIDEventTap           = 0
	kCGHeadInsertEventTap    = 0
	kCGEventTapOptionDefault = 0

	kCGEventFlagMaskControl   = 1 << 18
	kCGEventFlagMaskShift     = 1 << 17
	kCGEventFlagMaskAlternate = 1 << 19
	kCGEventFlagMaskCommand   = 1 << 20
)

var (
	cgEventTapCreate              func(tapLocation, placement, options int, mask uint64, callback, userInfo uintptr) uintptr
	cgEventTapEnable              func(tap uintptr, enable bool)
	cfMachPortCreateRunLoopSource func(allocator uintptr, port uintptr, order int) uintptr
	cfRunLoopGetCurrent           func() uintptr
	cfRunLoopAddSource            func(rl, source uintptr, mode uintptr)
	cfRunLoopRun                  func()
	cfRunLoopStop                 func(rl uintptr)
	cgEventGetIntegerValueField   func(event uintptr, field int) int64
	cgEventGetFlags               func(event uintptr) uint64
	cgEventGetType                func(event uintptr) uint32

	axIsProcessTrusted func() bool

	kCFRunLoopDefaultMode uintptr
)

func initDarwin() error {
	cg, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	ax, err := purego.Dlopen("/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices", purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}

	purego.RegisterLibFunc(&cgEventTapCreate, cg, "CGEventTapCreate")
	purego.RegisterLibFunc(&cgEventTapEnable, cg, "CGEventTapEnable")
	purego.RegisterLibFunc(&cgEventGetIntegerValueField, cg, "CGEventGetIntegerValueField")
	purego.RegisterLibFunc(&cgEventGetFlags, cg, "CGEventGetFlags")
	purego.RegisterLibFunc(&cgEventGetType, cg, "CGEventGetType")

	purego.RegisterLibFunc(&cfMachPortCreateRunLoopSource, cf, "CFMachPortCreateRunLoopSource")
	purego.RegisterLibFunc(&cfRunLoopGetCurrent, cf, "CFRunLoopGetCurrent")
	purego.RegisterLibFunc(&cfRunLoopAddSource, cf, "CFRunLoopAddSource")
	purego.RegisterLibFunc(&cfRunLoopRun, cf, "CFRunLoopRun")
	purego.RegisterLibFunc(&cfRunLoopStop, cf, "CFRunLoopStop")
	modeVar, err := purego.Dlsym(cf, "kCFRunLoopDefaultMode")
	if err != nil {
		return err
	}
	// Dlsym returns the address of the CFStringRef global; pass the string ref itself.
	kCFRunLoopDefaultMode = *(*uintptr)(unsafe.Pointer(modeVar))
	purego.RegisterLibFunc(&axIsProcessTrusted, ax, "AXIsProcessTrusted")
	return nil
}

var darwinInitOnce sync.Once
var darwinInitErr error

func ensureDarwin() error {
	darwinInitOnce.Do(func() { darwinInitErr = initDarwin() })
	return darwinInitErr
}

type darwinMonitor struct {
	mu       sync.Mutex
	chord    chord
	listener Listener
	tap      uintptr
	rl       uintptr
	active   bool
	started  bool
}

func newMonitor() Monitor { return &darwinMonitor{} }

func (m *darwinMonitor) Start(accelerator string, listener Listener) error {
	if m.started {
		return errors.New("holdshortcut: already started")
	}
	c, err := parseChord(accelerator)
	if err != nil {
		return err
	}
	if err := ensureDarwin(); err != nil {
		return err
	}
	if !axIsProcessTrusted() {
		return errors.New("holdshortcut: enable Accessibility for this app in System Settings → Privacy & Security → Accessibility (required for global key-up)")
	}

	m.mu.Lock()
	m.chord = c
	m.listener = listener
	m.mu.Unlock()

	mask := (uint64(1) << kCGEventKeyDown) | (uint64(1) << kCGEventKeyUp) | (uint64(1) << kCGEventFlagsChanged)
	tap := cgEventTapCreate(kCGHIDEventTap, kCGHeadInsertEventTap, kCGEventTapOptionDefault, mask, tapCallback, uintptr(unsafe.Pointer(m)))
	if tap == 0 {
		return errors.New("holdshortcut: CGEventTapCreate failed (check Accessibility permission)")
	}
	cgEventTapEnable(tap, true)

	m.rl = cfRunLoopGetCurrent()
	src := cfMachPortCreateRunLoopSource(0, tap, 0)
	if src == 0 {
		return errors.New("holdshortcut: CFMachPortCreateRunLoopSource failed")
	}
	cfRunLoopAddSource(m.rl, src, kCFRunLoopDefaultMode)

	m.mu.Lock()
	m.tap = tap
	m.started = true
	m.mu.Unlock()
	return nil
}

func (m *darwinMonitor) Stop() error {
	m.mu.Lock()
	if m.tap != 0 {
		cgEventTapEnable(m.tap, false)
	}
	if m.rl != 0 {
		cfRunLoopStop(m.rl)
	}
	m.started = false
	m.mu.Unlock()
	return nil
}

func runMonitor(m Monitor) error {
	dm, ok := m.(*darwinMonitor)
	if !ok {
		return ErrUnsupported
	}
	if !dm.started {
		return errors.New("holdshortcut: call Start before Run")
	}
	cfRunLoopRun()
	return nil
}

const (
	kCGKeyboardEventKeycode    = 9
	kCGKeyboardEventAutorepeat = 11
)

var tapCallback = purego.NewCallback(func(proxy, typ, event, userInfo uintptr) uintptr {
	m := (*darwinMonitor)(unsafe.Pointer(userInfo))
	if m.swallow(event) {
		return 0
	}
	return event
})

// swallow reports whether this event is part of the held chord. Those events
// are removed: Ctrl+Shift+Space otherwise repeats as NUL (^@) into the focused app.
func (m *darwinMonitor) swallow(event uintptr) bool {
	typ := cgEventGetType(event)
	keyCode := uint16(cgEventGetIntegerValueField(event, kCGKeyboardEventKeycode))
	flags := cgEventGetFlags(event)

	m.mu.Lock()
	c := m.chord
	listener := m.listener
	active := m.active
	m.mu.Unlock()

	if listener == nil {
		return false
	}

	switch typ {
	case kCGEventKeyDown:
		if keyCode != c.keyCode || !flagsMatch(flags, c) {
			return false
		}
		if cgEventGetIntegerValueField(event, kCGKeyboardEventAutorepeat) != 0 {
			return true
		}
		if !active {
			m.mu.Lock()
			m.active = true
			m.mu.Unlock()
			m.dispatch(listener.OnPress)
		}
		return true
	case kCGEventKeyUp:
		if keyCode != c.keyCode || !active {
			return false
		}
		m.mu.Lock()
		m.active = false
		m.mu.Unlock()
		m.dispatch(listener.OnRelease)
		return true
	case kCGEventFlagsChanged:
		if active && !flagsMatch(flags, c) {
			m.mu.Lock()
			m.active = false
			m.mu.Unlock()
			m.dispatch(listener.OnRelease)
		}
		return false
	default:
		return false
	}
}

func (m *darwinMonitor) dispatch(fn func()) {
	if fn == nil {
		return
	}
	go fn()
}

func flagsMatch(flags uint64, c chord) bool {
	if c.ctrl && flags&kCGEventFlagMaskControl == 0 {
		return false
	}
	if c.shift && flags&kCGEventFlagMaskShift == 0 {
		return false
	}
	if c.alt && flags&kCGEventFlagMaskAlternate == 0 {
		return false
	}
	if c.super && flags&kCGEventFlagMaskCommand == 0 {
		return false
	}
	return true
}

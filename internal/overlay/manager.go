package overlay

import (
	"errors"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/overlayglow"
)

type sessionMode int

const (
	modeHidden       sessionMode = 0
	modeRecording    sessionMode = 1
	modeTranscribing sessionMode = 2
)

const doublePulseGap = 100 * time.Millisecond

// Manager owns the glow window, animations, and locked display geometry.
type Manager struct {
	mu sync.Mutex

	win            *mygo.Window
	windowBounds   mygo.Rectangle
	lockedID       int64
	lockedBounds   mygo.Rectangle
	mode           sessionMode
	glow           overlayglow.State
	opacity        float64
	opacityFrom    float64
	opacityTarget  float64
	animStart      time.Time
	animDuration   time.Duration
	dimFrom        float64
	dimStart       time.Time
	dimTarget      float64
	hiding         bool
	transcribeAt   time.Time
	showTranscribe bool
	prepared       bool
	displaysHooked bool
	micPeak        atomic.Uint32
}

// NewManager creates an overlay manager. Call RegisterDisplayChanges after the app is ready.
func NewManager() *Manager {
	return &Manager{
		glow: overlayglow.State{
			Stops:         overlayglow.NewStops(),
			Start:         time.Now(),
			TranscribeDim: 1,
			GPU:           true,
		},
	}
}

// RegisterDisplayChanges refreshes placement when monitors change.
func (m *Manager) RegisterDisplayChanges() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		if m.displaysHooked {
			m.mu.Unlock()
			return
		}
		m.displaysHooked = true
		m.mu.Unlock()
		mygo.Screen.OnDisplaysChanged(func() {
			m.onDisplaysChanged()
		})
	})
}

// prepareLocked locks the display nearest the pointer. Call with m.mu held.
func (m *Manager) prepareLocked() {
	d, bounds := lockDisplay()
	m.lockedID = d.ID
	m.lockedBounds = bounds
	m.prepared = true
}

// PrepareSession locks the display nearest the pointer for the current dictation take.
func (m *Manager) PrepareSession() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		m.prepareLocked()
		m.mu.Unlock()
	})
}

func (m *Manager) OnMicPeak(peak int) {
	level := overlayglow.PeakToPulseLevel(peak)
	bits := math.Float32bits(level)
	for {
		old := m.micPeak.Load()
		if level <= math.Float32frombits(old) {
			return
		}
		if m.micPeak.CompareAndSwap(old, bits) {
			return
		}
	}
}

func (m *Manager) Pulse(strength float32) {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		m.glow.Voice.Queue(strength, 0)
		m.mu.Unlock()
		m.requestFrame()
	})
}

func (m *Manager) DoublePulse() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		m.glow.Voice.Queue(0.45, 0)
		m.glow.Voice.Queue(0.45, doublePulseGap)
		m.mu.Unlock()
		m.requestFrame()
	})
}

// ShowRecording fades the recording glow in on the prepared display.
func (m *Manager) ShowRecording() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		if !m.prepared {
			m.prepareLocked()
		}
		m.mode = modeRecording
		m.hiding = false
		m.showTranscribe = false
		m.glow.Start = time.Now()
		m.glow.Voice.Reset()
		m.micPeak.Store(0)
		m.glow.TranscribeDim = 1
		m.dimFrom, m.dimStart, m.dimTarget = 1, time.Time{}, 1
		m.glow.Visible = true
		if m.opacity <= 0.001 {
			m.opacity = 0
			m.opacityFrom = 0
		} else {
			m.opacityFrom = m.opacity
		}
		m.opacityTarget = 1
		m.animStart = time.Now()
		m.animDuration = fadeInDuration
		m.mu.Unlock()
		if err := m.ensureWindow(); err != nil {
			log.Printf("overlay: %v", err)
			m.teardown()
			return
		}
		m.present()
	})
}

// ShowTranscribing keeps the overlay on the locked display; after a short delay a subtler glow appears if still processing.
func (m *Manager) ShowTranscribing() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		m.mode = modeTranscribing
		m.transcribeAt = time.Now()
		m.showTranscribe = false
		m.hiding = false
		m.glow.Visible = true
		m.mu.Unlock()
		if m.win == nil {
			if err := m.ensureWindow(); err != nil {
				log.Printf("overlay transcribing: %v", err)
				m.teardown()
				return
			}
		}
		m.present()
	})
}

// Hide fades the glow out and dismisses the window when the animation completes.
func (m *Manager) Hide() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		if m.mode == modeHidden || m.hiding {
			m.mu.Unlock()
			return
		}
		m.hiding = true
		m.opacityFrom = m.opacity
		m.opacityTarget = 0
		m.animStart = time.Now()
		m.animDuration = fadeOutDuration
		m.mu.Unlock()
		m.requestFrame()
	})
}

// Close tears down the overlay immediately (application shutdown).
func (m *Manager) Close() {
	mygo.RunOnMain(func() {
		m.teardown()
	})
}

func (m *Manager) onDisplaysChanged() {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		if m.mode == modeHidden {
			m.mu.Unlock()
			return
		}
		if d, ok := displayByID(m.lockedID); ok {
			m.lockedBounds = overlayglow.DisplayRect(d)
		} else {
			m.mu.Unlock()
			m.teardown()
			return
		}
		m.mu.Unlock()
		if err := m.ensureWindow(); err != nil {
			log.Printf("overlay displays changed: %v", err)
			m.teardown()
			return
		}
		m.present()
	})
}

func (m *Manager) view(c *ui.Context) {
	now := time.Now()
	m.mu.Lock()
	fadeDur := fadeDuration(m.hiding, m.opacityTarget, m.opacityFrom, c.Preferences().ReduceMotion, m.animDuration)
	ease := easeInOut
	if m.opacityTarget > m.opacityFrom {
		ease = easeOutShow
	}
	m.opacity = advanceOpacity(now, m.animStart, m.opacityFrom, m.opacityTarget, fadeDur, ease)
	m.glow.Opacity = m.opacity

	if m.mode == modeTranscribing && !m.showTranscribe && now.Sub(m.transcribeAt) >= TranscribeVisualDelay {
		m.showTranscribe = true
	}
	if m.showTranscribe {
		if m.dimTarget != 0.45 {
			m.dimFrom = m.glow.TranscribeDim
			m.dimStart = now
			m.dimTarget = 0.45
		}
	} else if m.dimTarget != 1 {
		m.dimFrom = m.glow.TranscribeDim
		m.dimStart = now
		m.dimTarget = 1
	}
	m.glow.TranscribeDim = advanceOpacity(now, m.dimStart, m.dimFrom, m.dimTarget, transcribeDimDuration, easeOutShow)

	mode := m.mode
	hiding := m.hiding
	recording := mode == modeRecording || mode == modeTranscribing
	showGlow := (recording && !hiding) || (hiding && m.opacity > 0.001)
	m.glow.Visible = showGlow
	if bits := m.micPeak.Swap(0); bits != 0 {
		m.glow.Voice.Push(math.Float32frombits(bits))
	}
	state := m.glow
	finishHide := hiding && m.opacity <= 0.001
	m.mu.Unlock()

	overlayglow.Draw(&state, c)
	m.mu.Lock()
	m.glow.GPU = state.GPU
	m.glow.Voice = state.Voice
	m.glow.LastFrame = state.LastFrame
	m.glow.GradientPhase = state.GradientPhase
	m.mu.Unlock()

	if (recording || hiding) && !finishHide {
		c.AnimationFrame()
	}

	if finishHide {
		mygo.RunOnMain(m.finishHide)
	}
}

func (m *Manager) finishHide() {
	m.mu.Lock()
	m.hiding = false
	m.mode = modeHidden
	m.glow.Visible = false
	m.glow.Voice.Reset()
	m.micPeak.Store(0)
	m.glow.TranscribeDim = 1
	m.dimFrom, m.dimStart, m.dimTarget = 1, time.Time{}, 1
	m.opacity = 0
	m.prepared = false
	m.mu.Unlock()
	if m.win != nil {
		m.win.Hide()
	}
}

func (m *Manager) teardown() {
	m.mu.Lock()
	m.mode = modeHidden
	m.hiding = false
	m.glow.Visible = false
	m.glow.Voice.Reset()
	m.micPeak.Store(0)
	m.glow.TranscribeDim = 1
	m.dimFrom, m.dimStart, m.dimTarget = 1, time.Time{}, 1
	m.opacity = 0
	m.showTranscribe = false
	m.prepared = false
	win := m.win
	m.win = nil
	m.mu.Unlock()
	if win != nil {
		win.Hide()
		win.Close()
	}
}

func (m *Manager) requestFrame() {
	if m.win != nil {
		m.win.Update(func() {})
	}
}

func (m *Manager) ensureWindow() error {
	m.mu.Lock()
	bounds := m.lockedBounds
	if m.win != nil && m.windowBounds == bounds {
		m.mu.Unlock()
		return nil
	}
	if m.win != nil {
		win := m.win
		m.win = nil
		m.mu.Unlock()
		win.Close()
	} else {
		m.mu.Unlock()
	}

	win := mygo.NewWindow(mygo.WindowOptions{
		Title:         "speek",
		X:             1,
		Y:             1,
		Width:         bounds.Width,
		Height:        bounds.Height,
		Hidden:        true,
		Panel:         true,
		Frameless:     true,
		Transparent:   true,
		DisableShadow: true,
		AlwaysOnTop:   true,
		SkipTaskbar:   true,
		DisableResize: true,
		Content:       ui.View(m.view),
	})
	if win == nil {
		return errors.New("overlay: could not create window")
	}
	m.mu.Lock()
	m.win = win
	m.windowBounds = bounds
	m.mu.Unlock()
	return nil
}

func (m *Manager) present() {
	m.mu.Lock()
	win := m.win
	b := m.lockedBounds
	m.mu.Unlock()
	if win == nil {
		return
	}
	win.ShowInactive()
	win.SetIgnoreMouseEvents(true)
	overlayglow.PlaceOverFrame(win.NativeHandle(), float64(b.X), float64(b.Y), float64(b.Width), float64(b.Height))
	win.Update(func() {})
}

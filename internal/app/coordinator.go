package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
	"github.com/speek/speek/internal/paste"
	"github.com/speek/speek/internal/shortcut"
	"github.com/speek/speek/internal/transcription"
)

// Coordinator routes push-to-talk input through the dictation state machine.
type Coordinator struct {
	svc   Services
	store *config.Store
	opts  Options

	mu sync.Mutex

	state              State
	releaseDuringStart bool
	blockPTTDown       bool
	chordHeld          bool
	shuttingDown       bool
	trayTooltip        string

	session       uint64
	sessionCtx    context.Context
	sessionCancel context.CancelFunc
	recordTimer   *time.Timer
}

// NewCoordinator creates the input coordinator.
func NewCoordinator(svc Services, store *config.Store, opts Options) *Coordinator {
	return &Coordinator{svc: svc, store: store, opts: opts, state: StateIdle}
}

// State returns the current coordinator state.
func (c *Coordinator) State() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// CanCancel reports whether the tray cancel action should be enabled.
func (c *Coordinator) CanCancel() bool {
	s := c.State()
	return s == StateRecording || s == StateTranscribing
}

// TrayTooltip returns a tray status override, or empty for the default label.
func (c *Coordinator) TrayTooltip() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.trayTooltip
}

// StartShortcut registers the configured hold chord.
func (c *Coordinator) StartShortcut() error {
	cfg := c.store.Get()
	return c.svc.Shortcut.Start(cfg.PushToTalkShortcut, c.events())
}

// StopShortcut unregisters the global chord.
func (c *Coordinator) StopShortcut() error {
	return c.svc.Shortcut.Stop()
}

// ReloadShortcut applies an updated shortcut from configuration.
func (c *Coordinator) ReloadShortcut() error {
	if err := c.StopShortcut(); err != nil {
		return err
	}
	return c.StartShortcut()
}

// Shutdown stops an active session when the application is quitting.
func (c *Coordinator) Shutdown() {
	c.mu.Lock()
	c.shuttingDown = true
	c.mu.Unlock()
	c.abortSession(true)
}

// CancelDictation stops capture or transcription and returns to idle without pasting.
func (c *Coordinator) CancelDictation() {
	c.mu.Lock()
	switch c.state {
	case StateRecording, StateTranscribing:
		// Only swallow the next key-down when the chord is still held; if it was
		// already released, a fresh press should start a new session.
		if c.chordHeld {
			c.blockPTTDown = true
		}
	default:
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()
	c.abortSession(false)
}

func (c *Coordinator) events() shortcut.Events {
	return shortcut.Events{
		OnPress:   c.onPress,
		OnRelease: c.onRelease,
	}
}

func (c *Coordinator) onPress() {
	c.mu.Lock()
	if c.shuttingDown || c.blockPTTDown {
		c.mu.Unlock()
		return
	}
	switch c.state {
	case StateIdle:
		c.chordHeld = true
		if msg, err := c.modelFailureMessage(); err != nil {
			gen := c.session
			c.mu.Unlock()
			c.userFailure(gen, msg, "", false)
			return
		}
		c.svc.Paste.CapturePasteTarget()
		c.state = StateStarting
		c.releaseDuringStart = false
		c.trayTooltip = ""
		c.session++
		gen := c.session
		ctx, cancel := context.WithCancel(context.Background())
		c.sessionCtx = ctx
		c.sessionCancel = cancel
		c.mu.Unlock()
		c.notifyState(StateStarting)
		go c.beginSession(gen, ctx)
	default:
		// Duplicate key-down while starting or already recording.
		c.mu.Unlock()
	}
}

func (c *Coordinator) onRelease() {
	c.mu.Lock()
	c.blockPTTDown = false
	c.chordHeld = false
	switch c.state {
	case StateStarting:
		c.releaseDuringStart = true
	case StateRecording:
		gen := c.session
		c.mu.Unlock()
		go c.finishRecording(gen)
		return
	default:
		// Idle or transcribing: release is a no-op.
	}
	c.mu.Unlock()
}

func (c *Coordinator) beginSession(gen uint64, ctx context.Context) {
	c.svc.Overlay.PrepareSession()
	if err := c.svc.Audio.Start(ctx); err != nil {
		c.failStart(gen, err)
		return
	}
	c.svc.Overlay.ShowRecording()
	c.mu.Lock()
	if c.shuttingDown || gen != c.session || c.state != StateStarting {
		c.mu.Unlock()
		_, _ = c.svc.Audio.Stop()
		return
	}
	release := c.releaseDuringStart
	c.mu.Unlock()

	if release {
		go c.finishRecording(gen)
		return
	}

	c.mu.Lock()
	if gen != c.session || c.state != StateStarting {
		c.mu.Unlock()
		return
	}
	c.state = StateRecording
	c.mu.Unlock()
	c.notifyState(StateRecording)
	c.armRecordTimer(gen)
}

func (c *Coordinator) armRecordTimer(gen uint64) {
	d := c.maxRecordingDuration()
	c.mu.Lock()
	if c.recordTimer != nil {
		c.recordTimer.Stop()
	}
	c.recordTimer = time.AfterFunc(d, func() {
		c.mu.Lock()
		ok := gen == c.session && c.state == StateRecording
		if ok && c.chordHeld {
			c.blockPTTDown = true
		}
		if ok {
			c.trayTooltip = "Recording reached the time limit"
		}
		c.mu.Unlock()
		if ok {
			c.finishRecording(gen)
		}
	})
	c.mu.Unlock()
}

func (c *Coordinator) finishRecording(gen uint64) {
	c.mu.Lock()
	current := gen == c.session && (c.state == StateRecording || c.state == StateStarting)
	c.mu.Unlock()
	if !current {
		// A stale goroutine (superseded by cancel or a new session) must not
		// stop the live session's capture or touch its state.
		return
	}
	c.stopRecordTimer()

	samples, err := c.svc.Audio.Stop()
	if err != nil {
		if errors.Is(err, audio.ErrNotRecording) {
			// Session already ended (timer-vs-release or cancel-vs-release race).
			return
		}
		msg := "Could not stop microphone capture."
		if errors.Is(err, audio.ErrPermissionDenied) || errors.Is(err, audio.ErrSilentInput) {
			msg = "The microphone produced silence. Quit this copy and launch Speek.app (make package, then open dist/Speek.app). Allow speek in System Settings → Privacy & Security → Microphone."
		}
		c.userFailure(gen, msg, "", true)
		return
	}
	if len(samples) == 0 {
		c.userFailure(gen, "No audio captured.", "", false)
		return
	}
	go c.transcribeAndPaste(gen, samples)
}

func (c *Coordinator) transcribeAndPaste(gen uint64, samples []int16) {
	c.mu.Lock()
	if c.shuttingDown {
		c.mu.Unlock()
		return
	}
	if gen != c.session || (c.state != StateStarting && c.state != StateRecording) {
		c.mu.Unlock()
		return
	}
	c.state = StateTranscribing
	ctx := c.sessionCtx
	c.mu.Unlock()
	c.notifyState(StateTranscribing)
	c.svc.Overlay.ShowTranscribing()

	modelID, err := c.svc.Models.EffectiveModelID(ctx)
	if err != nil {
		msg, _ := c.modelFailureMessage()
		if msg == "" {
			msg = err.Error()
		}
		c.userFailure(gen, msg, "", true)
		return
	}

	text, err := c.svc.Transcription.Transcribe(ctx, samples, modelID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			c.returnIdle(gen)
			return
		}
		if errors.Is(err, transcription.ErrNoSpeech) || errors.Is(err, transcription.ErrEmptyAudio) {
			c.svc.Overlay.DoublePulse()
			c.userFailure(gen, "No speech detected.", "", false)
			return
		}
		if errors.Is(err, transcription.ErrUnintelligible) {
			c.userFailure(gen, "The microphone picked up sound, but it was not recognized. Open Settings, choose the microphone, and use Test microphone.", "", true)
			return
		}
		c.userFailure(gen, "Transcription failed.", "", true)
		return
	}
	if strings.TrimSpace(text) == "" {
		c.svc.Overlay.DoublePulse()
		c.userFailure(gen, "No speech detected.", "", false)
		return
	}

	c.mu.Lock()
	if c.shuttingDown || gen != c.session || c.state != StateTranscribing {
		c.mu.Unlock()
		return
	}
	c.state = StatePasting
	c.mu.Unlock()
	c.notifyState(StatePasting)

	if err := c.svc.Paste.Paste(ctx, text); err != nil {
		if errors.Is(err, context.Canceled) {
			c.returnIdle(gen)
			return
		}
		c.userFailure(gen, pasteFailureMessage(err), text, true)
		return
	}

	c.mu.Lock()
	c.trayTooltip = ""
	c.mu.Unlock()
	c.svc.Overlay.Pulse(0.9)
	c.returnIdle(gen)
}

func (c *Coordinator) failStart(gen uint64, err error) {
	log.Printf("capture start: %v", err)
	c.mu.Lock()
	stale := gen != c.session || c.state != StateStarting
	c.mu.Unlock()
	if stale {
		return
	}
	var msg string
	switch {
	case errors.Is(err, audio.ErrPermissionDenied):
		msg = "Microphone access is off. Launch Speek.app so macOS lists speek, then allow it in System Settings → Privacy & Security → Microphone."
	case errors.Is(err, audio.ErrNoInputDevice), errors.Is(err, audio.ErrDeviceNotFound):
		msg = "No microphone available"
	default:
		msg = err.Error()
	}
	c.userFailure(gen, msg, "", true)
}

func (c *Coordinator) abortSession(shutdown bool) {
	c.stopRecordTimer()
	c.mu.Lock()
	cancel := c.sessionCancel
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	c.svc.Overlay.Hide()

	c.mu.Lock()
	state := c.state
	c.session++
	c.mu.Unlock()

	if state == StateRecording || state == StateStarting {
		_, _ = c.svc.Audio.Stop()
	}

	if shutdown {
		return
	}
	c.mu.Lock()
	c.state = StateIdle
	c.mu.Unlock()
	c.notifyState(StateIdle)
}

func (c *Coordinator) modelFailureMessage() (string, error) {
	_, err := c.svc.Models.EffectiveModelID(context.Background())
	if err == nil {
		return "", nil
	}
	id := strings.TrimSpace(c.store.Get().ModelID)
	if id == "" {
		id = models.DefaultModelID
	}
	switch {
	case errors.Is(err, models.ErrNoSelection):
		return "Multiple models installed. Open Settings and choose a model.", err
	case errors.Is(err, models.ErrNotInstalled), errors.Is(err, models.ErrChecksumMismatch):
		return fmt.Sprintf("Model \"%s\" is not installed. Open Settings to download it.", id), err
	default:
		return err.Error(), err
	}
}

func pasteFailureMessage(err error) string {
	switch {
	case errors.Is(err, paste.ErrFocusChanged):
		return "Another application is focused. The previous clipboard was kept."
	case errors.Is(err, paste.ErrPermissionDenied):
		return "Accessibility permission is required to paste automatically. The transcript is on the clipboard."
	case errors.Is(err, paste.ErrTargetLost):
		return "The original application is no longer available. The previous clipboard was kept."
	default:
		return err.Error()
	}
}

// userFailure shows tray/overlay feedback and optionally a dialog; never logs transcript text.
// A stale generation (session superseded by cancel or a new press) is a no-op so the
// old goroutine cannot clobber the live session's state, tooltip, or dialogs.
func (c *Coordinator) userFailure(gen uint64, message, recognized string, dialog bool) {
	log.Printf("dictation: %s", message)
	c.mu.Lock()
	if gen != c.session {
		c.mu.Unlock()
		return
	}
	c.trayTooltip = message
	c.state = StateError
	c.mu.Unlock()
	c.notifyState(StateError)
	// Status stays on the tray tooltip. A fullscreen overlay for this text
	// covers the display and ignores clicks.
	if dialog && c.opts.OnError != nil {
		c.opts.OnError(message, recognized)
	}
	c.mu.Lock()
	c.state = StateIdle
	c.mu.Unlock()
	c.notifyState(StateIdle)
	c.svc.Overlay.Hide()
}

func (c *Coordinator) returnIdle(gen uint64) {
	c.mu.Lock()
	stale := gen != c.session
	c.mu.Unlock()
	if stale {
		return
	}
	c.stopRecordTimer()
	c.mu.Lock()
	c.state = StateIdle
	c.mu.Unlock()
	c.notifyState(StateIdle)
	c.svc.Overlay.Hide()
}

func (c *Coordinator) stopRecordTimer() {
	c.mu.Lock()
	if c.recordTimer != nil {
		c.recordTimer.Stop()
		c.recordTimer = nil
	}
	c.mu.Unlock()
}

func (c *Coordinator) notifyState(s State) {
	if c.opts.OnStateChange != nil {
		c.opts.OnStateChange(s)
	}
}

func (c *Coordinator) maxRecordingDuration() time.Duration {
	sec := c.store.Get().MaxRecordingSeconds
	if sec <= 0 {
		return DefaultMaxRecording
	}
	return time.Duration(sec) * time.Second
}

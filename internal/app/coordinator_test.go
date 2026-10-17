package app

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
	"github.com/speek/speek/internal/overlay"
	"github.com/speek/speek/internal/paste"
	"github.com/speek/speek/internal/shortcut"
	"github.com/speek/speek/internal/transcription"
)

func testCoordinator(svc Services, store *config.Store, opts Options) (*Coordinator, *shortcut.FakeMonitor) {
	sc := &shortcut.FakeMonitor{}
	svc.Shortcut = sc
	c := NewCoordinator(svc, store, opts)
	if err := c.StartShortcut(); err != nil {
		panic(err)
	}
	return c, sc
}

func defaultTestServices() Services {
	m := &models.FakeManager{Selected: "ggml-tiny"}
	return Services{
		Overlay:       &overlay.FakeController{},
		Audio:         &audio.FakeCapturer{},
		Transcription: transcription.FakeEngine{Text: "hello"},
		Paste:         &paste.FakeService{},
		Models:        m,
	}
}

func TestTranscribingOverlay(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Transcription = stallEngine{}
	ov := &overlay.FakeController{}
	svc.Overlay = ov
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateTranscribing })
	if !ov.Transcribing {
		t.Fatal("expected transcribing overlay")
	}
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
}

func TestHappyPathPressRelease(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	pasteFake := svc.Paste.(*paste.FakeService)
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	if pasteFake.Captures != 1 {
		t.Fatalf("expected target capture on press, got %d", pasteFake.Captures)
	}
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if pasteFake.Last != "hello" {
		t.Fatalf("paste %q", pasteFake.Last)
	}
}

func TestKeyUpDuringStarting(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	unblock := make(chan struct{})
	svc := defaultTestServices()
	svc.Audio = &audio.BlockingFake{Unblock: unblock}
	pasteFake := svc.Paste.(*paste.FakeService)
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateStarting })
	sc.Events.OnRelease()
	close(unblock)
	waitFor(t, func() bool { return c.State() == StateIdle })
	if pasteFake.Last != "hello" {
		t.Fatalf("expected paste after early release, got %q", pasteFake.Last)
	}
}

func TestDuplicateKeyDownIgnoredWhileRecording(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	starts := 0
	svc := defaultTestServices()
	svc.Audio = &audio.FakeCapturer{
		OnStart: func(context.Context) error {
			starts++
			return nil
		},
	}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnPress()
	sc.Events.OnPress()
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if starts != 1 {
		t.Fatalf("expected one capture start, got %d", starts)
	}
}

func TestIgnoreKeyUpWhenIdle(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	c, sc := testCoordinator(defaultTestServices(), store, Options{})
	sc.Events.OnRelease()
	if c.State() != StateIdle {
		t.Fatal("unexpected state change")
	}
}

func TestCancelFromRecording(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	pasteFake := svc.Paste.(*paste.FakeService)
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if pasteFake.Last != "" {
		t.Fatal("paste should not run on cancel")
	}

	// Key-down blocked until physical release after cancel.
	sc.Events.OnPress()
	if c.State() != StateIdle {
		t.Fatal("expected blocked key-down")
	}
	sc.Events.OnRelease()
	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
}

func TestCancelFromTranscribing(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Transcription = stallEngine{}
	pasteFake := svc.Paste.(*paste.FakeService)
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateTranscribing })
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if pasteFake.Last != "" {
		t.Fatal("paste should not run after cancel")
	}
}

type stallEngine struct{}

func (stallEngine) Transcribe(ctx context.Context, _ []int16, _ string) (string, error) {
	<-ctx.Done()
	return "", ctx.Err()
}

// gatedEngine pins the first Transcribe call on a gate (so that goroutine goes
// stale) and then behaves as a normal engine for later sessions.
type gatedEngine struct {
	gate  chan struct{}
	mu    sync.Mutex
	first bool
}

func (g *gatedEngine) Transcribe(ctx context.Context, _ []int16, _ string) (string, error) {
	g.mu.Lock()
	first := g.first
	g.first = false
	g.mu.Unlock()
	if first {
		<-g.gate
		return "", ctx.Err()
	}
	return "hello", nil
}

func TestStaleTranscribeGoroutineDoesNotClobberNewSession(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	gate := make(chan struct{})
	svc := defaultTestServices()
	svc.Transcription = &gatedEngine{gate: gate, first: true}
	ov := &overlay.FakeController{}
	svc.Overlay = ov
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateTranscribing })

	// Cancel (key already up), then start a fresh session on the same engine.
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })

	// Release the old goroutine: it sees context.Canceled and must no-op.
	close(gate)
	time.Sleep(150 * time.Millisecond)
	if c.State() != StateRecording {
		t.Fatalf("state %v, want StateRecording", c.State())
	}
	if !ov.Visible {
		t.Fatal("overlay for the new session should still be visible")
	}

	// The new session still completes normally.
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	pasteFake := svc.Paste.(*paste.FakeService)
	if pasteFake.Last != "hello" {
		t.Fatalf("new session paste %q", pasteFake.Last)
	}
}

func TestCancelWithKeyUpDoesNotBlockNextPress(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Transcription = stallEngine{}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateTranscribing })

	// Chord is already up when cancel lands: the next press must work.
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
	sc.Events.OnPress()
	waitFor(t, func() bool {
		return c.State() == StateRecording || c.State() == StateTranscribing
	})
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateTranscribing })
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
}

func TestRecordingDurationLimit(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Transcription = stallEngine{}
	c, sc := testCoordinator(svc, store, Options{})
	cfg := store.Get()
	cfg.MaxRecordingSeconds = 1
	if err := store.Set(cfg); err != nil {
		t.Fatal(err)
	}

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c.State() == StateTranscribing {
			break
		}
		if c.State() == StateIdle || c.State() == StateError {
			t.Fatalf("unexpected state %v before limit", c.State())
		}
		time.Sleep(5 * time.Millisecond)
	}
	if c.State() != StateTranscribing {
		t.Fatalf("expected transcribing after limit, got %v", c.State())
	}

	sc.Events.OnPress()
	if c.State() != StateTranscribing {
		t.Fatal("key-down should be blocked while shortcut still held after limit")
	}
	if c.TrayTooltip() != "Recording reached the time limit" {
		t.Fatalf("tooltip %q", c.TrayTooltip())
	}
	c.CancelDictation()
	waitFor(t, func() bool { return c.State() == StateIdle })
}

func TestNoSpeechSkipsPaste(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	pasteFake := svc.Paste.(*paste.FakeService)
	svc.Transcription = transcription.FakeEngine{Err: transcription.ErrNoSpeech}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if pasteFake.Last != "" {
		t.Fatal("paste should be skipped for no speech")
	}
}

func TestSilenceShowsStatus(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	ov := &overlay.FakeController{}
	svc.Overlay = ov
	svc.Transcription = transcription.FakeEngine{Err: transcription.ErrNoSpeech}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if !strings.Contains(c.TrayTooltip(), "No speech") {
		t.Fatalf("tooltip %q", c.TrayTooltip())
	}
}

func TestPasteFocusChangedError(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	pasteFake := svc.Paste.(*paste.FakeService)
	pasteFake.Err = paste.ErrFocusChanged
	var errMsg string
	c, sc := testCoordinator(svc, store, Options{
		OnError: func(msg, recognized string) {
			errMsg = msg
			if recognized != "hello" {
				t.Fatalf("recognized %q", recognized)
			}
		},
	})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if !strings.Contains(errMsg, "clipboard") {
		t.Fatalf("msg %q", errMsg)
	}
}

func TestTranscriptionFailure(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	var errMsg string
	svc := defaultTestServices()
	svc.Transcription = transcription.FakeEngine{Err: errors.New("boom")}
	c, sc := testCoordinator(svc, store, Options{
		OnError: func(msg, _ string) { errMsg = msg },
	})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if errMsg == "" {
		t.Fatal("expected error callback")
	}
}

func TestCaptureStartFailureReturnsIdle(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Audio = &audio.FakeCapturer{
		OnStart: func(context.Context) error { return errors.New("mic denied") },
	}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateIdle })
	sc.Events.OnRelease()
}

func TestShutdownStopsRecording(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	stopped := make(chan struct{}, 1)
	svc := defaultTestServices()
	svc.Audio = &audio.FakeCapturer{
		OnStop: func() ([]int16, error) {
			stopped <- struct{}{}
			return nil, nil
		},
	}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	c.Shutdown()
	waitFor(t, func() bool {
		select {
		case <-stopped:
			return true
		default:
			return false
		}
	})
}

func TestNoModelError(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	ov := &overlay.FakeController{}
	svc.Overlay = ov
	svc.Models = &models.FakeManager{Selected: ""}
	var errMsg string
	c, sc := testCoordinator(svc, store, Options{
		OnError: func(msg, _ string) { errMsg = msg },
	})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if c.State() == StateRecording {
		t.Fatal("should not start recording without a model")
	}
	if !strings.Contains(errMsg, "not installed") && !strings.Contains(c.TrayTooltip(), "not installed") {
		t.Fatalf("msg %q tooltip %q", errMsg, c.TrayTooltip())
	}
}

func TestCoordinatorOverlayOnHold(t *testing.T) {
	store, err := config.Open(t.TempDir() + "/cfg.json")
	if err != nil {
		t.Fatal(err)
	}
	ov := &overlay.FakeController{}
	sc := &shortcut.FakeMonitor{}
	svc := defaultTestServices()
	svc.Overlay = ov
	svc.Shortcut = sc
	svc.Models = &models.FakeManager{Selected: "ggml-tiny"}
	coord := NewCoordinator(svc, store, Options{})
	if err := coord.StartShortcut(); err != nil {
		t.Fatal(err)
	}
	if sc.Accelerator != store.Get().PushToTalkShortcut {
		t.Fatalf("accelerator %q", sc.Accelerator)
	}

	sc.Events.OnPress()
	waitFor(t, func() bool { return ov.Visible })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return coord.State() == StateIdle })
}

func waitFor(t *testing.T, cond func() bool) {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func TestConcurrentTranscribeGuard(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	var transcribing sync.WaitGroup
	transcribing.Add(1)
	svc := defaultTestServices()
	svc.Transcription = slowEngine{&transcribing}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateTranscribing })
	sc.Events.OnPress()
	if c.State() != StateTranscribing {
		t.Fatalf("state %v", c.State())
	}
	transcribing.Done()
	waitFor(t, func() bool { return c.State() == StateIdle })
}

type slowEngine struct {
	wg *sync.WaitGroup
}

func (s slowEngine) Transcribe(ctx context.Context, samples []int16, modelID string) (string, error) {
	s.wg.Wait()
	return "ok", nil
}

func TestSilentInputShowsMicGuidance(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	pasteFake := svc.Paste.(*paste.FakeService)
	svc.Audio = &audio.FakeCapturer{
		OnStop: func() ([]int16, error) { return nil, audio.ErrSilentInput },
	}
	var errMsg string
	c, sc := testCoordinator(svc, store, Options{
		OnError: func(msg, _ string) { errMsg = msg },
	})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	tt := c.TrayTooltip()
	if !strings.Contains(tt, "silence") && !strings.Contains(strings.ToLower(tt), "microphone") {
		t.Fatalf("tooltip %q should mention silence/microphone", tt)
	}
	if errMsg == "" {
		t.Fatal("expected OnError for silent input")
	}
	if pasteFake.Last != "" {
		t.Fatal("no paste expected for silent input")
	}
}

func TestEmptySamplesNoPaste(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	pasteFake := svc.Paste.(*paste.FakeService)
	svc.Audio = &audio.FakeCapturer{
		OnStop: func() ([]int16, error) { return nil, nil },
	}
	var errFired bool
	c, sc := testCoordinator(svc, store, Options{
		OnError: func(string, string) { errFired = true },
	})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if tt := c.TrayTooltip(); tt != "No audio captured." {
		t.Fatalf("tooltip %q want %q", tt, "No audio captured.")
	}
	if pasteFake.Last != "" {
		t.Fatal("no paste expected for empty samples")
	}
	if errFired {
		t.Fatal("OnError should not fire for empty samples")
	}
}

func TestCaptureStartFailureMapsErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"permission", audio.ErrPermissionDenied, "Microphone access is off"},
		{"no device", audio.ErrNoInputDevice, "No microphone available"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _ := config.Open(t.TempDir() + "/cfg.json")
			svc := defaultTestServices()
			svc.Audio = &audio.FakeCapturer{
				OnStart: func(context.Context) error { return tc.err },
			}
			c, sc := testCoordinator(svc, store, Options{})

			sc.Events.OnPress()
			waitFor(t, func() bool { return c.State() == StateIdle })
			sc.Events.OnRelease()
			if !strings.Contains(c.TrayTooltip(), tc.want) {
				t.Fatalf("tooltip %q should mention %q", c.TrayTooltip(), tc.want)
			}
		})
	}
}

func TestMultipleModelsRequiresSelection(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Models = &models.FakeManager{
		Selected: "",
		Models:   []models.Model{{ID: "ggml-tiny"}, {ID: "ggml-base"}},
	}
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	if c.State() != StateIdle {
		t.Fatalf("state %v want StateIdle", c.State())
	}
	if !strings.Contains(c.TrayTooltip(), "Multiple models installed") {
		t.Fatalf("tooltip %q", c.TrayTooltip())
	}
	sc.Events.OnRelease()
}

func TestReloadShortcutAppliesNewAccelerator(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	cfg := store.Get()
	cfg.PushToTalkShortcut = "Alt+Shift+Space"
	if err := store.Set(cfg); err != nil {
		t.Fatal(err)
	}
	c, sc := testCoordinator(defaultTestServices(), store, Options{})
	if sc.Accelerator != "Alt+Shift+Space" {
		t.Fatalf("initial accelerator %q", sc.Accelerator)
	}

	cfg = store.Get()
	cfg.PushToTalkShortcut = "Ctrl+Shift+Space"
	if err := store.Set(cfg); err != nil {
		t.Fatal(err)
	}
	if err := c.ReloadShortcut(); err != nil {
		t.Fatal(err)
	}
	if sc.Accelerator != "Ctrl+Shift+Space" {
		t.Fatalf("accelerator after reload %q", sc.Accelerator)
	}
	if !sc.Started {
		t.Fatal("monitor not started after reload")
	}

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })
}

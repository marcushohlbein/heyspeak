package audio

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/speek/speek/internal/config"
)

// Recorder captures microphone audio using malgo with bounded buffers.
type Recorder struct {
	store *config.Store

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
	out    []int16
	err    error
}

// NewRecorder creates a capturer backed by configuration.
func NewRecorder(store *config.Store) *Recorder {
	return &Recorder{store: store}
}

// Start begins capture until Stop is called or ctx is cancelled.
func (r *Recorder) Start(ctx context.Context) error {
	if err := ensureMicPermission(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		return ErrAlreadyRecording
	}
	cfg := r.store.Get()
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	r.cancel = cancel
	r.done = done
	r.out = nil
	r.err = nil

	sessionCfg := SessionConfig{
		DeviceID:   cfg.InputDeviceID,
		MaxSeconds: cfg.MaxRecordingSeconds,
		SampleRate: uint32(SampleRate),
	}

	go func() {
		samples, err := recordSession(runCtx, sessionCfg)
		r.mu.Lock()
		if r.done == done {
			// Still the current session: publish results for Stop to consume.
			// A wedged previous session that timed out in Stop has been
			// detached (r.done replaced or cleared); never touch its fields.
			r.out = samples
			r.err = err
			close(done)
			r.cancel = nil
		}
		r.mu.Unlock()
	}()
	return nil
}

// Stop ends an active capture and returns recorded samples at SampleRate mono PCM16.
// A capture that finished on its own still holds results; Stop consumes them.
// If the capture goroutine is wedged, Stop gives up after 3s and resets so a
// later Start recovers.
func (r *Recorder) Stop() ([]int16, error) {
	r.mu.Lock()
	cancel := r.cancel
	done := r.done
	r.mu.Unlock()
	if cancel == nil && done == nil {
		return nil, ErrNotRecording
	}
	if cancel != nil {
		cancel()
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		r.mu.Lock()
		r.cancel = nil
		r.done = nil
		r.out = nil
		r.err = nil
		r.mu.Unlock()
		return nil, errors.New("audio: stop timed out")
	}
	r.mu.Lock()
	out, err := r.out, r.err
	r.done = nil
	r.out = nil
	r.err = nil
	r.mu.Unlock()
	return out, err
}

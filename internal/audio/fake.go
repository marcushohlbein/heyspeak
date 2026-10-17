package audio

import (
	"context"
	"sync"
)

// FakeCapturer is a test double for Capturer.
type FakeCapturer struct {
	mu      sync.Mutex
	active  bool
	OnStart func(ctx context.Context) error
	OnStop  func() ([]int16, error)
}

func (f *FakeCapturer) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.active {
		return ErrAlreadyRecording
	}
	if f.OnStart != nil {
		if err := f.OnStart(ctx); err != nil {
			return err
		}
	}
	f.active = true
	return nil
}

func (f *FakeCapturer) Stop() ([]int16, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.active {
		return nil, ErrNotRecording
	}
	f.active = false
	if f.OnStop != nil {
		return f.OnStop()
	}
	return []int16{1, 2, 3}, nil
}

// BlockingFake waits on unblock before Start returns, for race tests.
type BlockingFake struct {
	mu      sync.Mutex
	active  bool
	Unblock chan struct{}
	OnStop  func() ([]int16, error)
}

func (f *BlockingFake) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.Unblock:
	}
	f.mu.Lock()
	f.active = true
	f.mu.Unlock()
	return nil
}

func (f *BlockingFake) Stop() ([]int16, error) {
	f.mu.Lock()
	if !f.active {
		f.mu.Unlock()
		return nil, ErrNotRecording
	}
	f.active = false
	f.mu.Unlock()
	if f.OnStop != nil {
		return f.OnStop()
	}
	return []int16{9, 9, 9}, nil
}

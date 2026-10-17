package audio

import (
	"context"
	"errors"
	"testing"

	"github.com/speek/speek/internal/config"
)

func TestRecorderSessionLifecycle(t *testing.T) {
	store, err := config.Open(t.TempDir() + "/cfg.json")
	if err != nil {
		t.Fatal(err)
	}
	r := NewRecorder(store)
	if err := r.Start(context.Background()); err != nil {
		if errors.Is(err, ErrPermissionDenied) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	if err := r.Start(context.Background()); err != ErrAlreadyRecording {
		t.Fatalf("expected already recording, got %v", err)
	}
	_, err = r.Stop()
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Stop()
	if err != ErrNotRecording {
		t.Fatalf("expected not recording, got %v", err)
	}
}

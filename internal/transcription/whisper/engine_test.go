package whisper

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
	"github.com/speek/speek/internal/transcription"
)

func TestTranscribeNoSpeechShort(t *testing.T) {
	store, _ := config.Open(filepath.Join(t.TempDir(), "cfg.json"))
	e := NewEngine(models.NewStoreAt(store, t.TempDir()), store)
	_, err := e.Transcribe(context.Background(), []int16{1}, "ggml-tiny")
	if err != transcription.ErrNoSpeech {
		t.Fatalf("got %v", err)
	}
}

func TestTranscribeEmptyAudio(t *testing.T) {
	store, _ := config.Open(filepath.Join(t.TempDir(), "cfg.json"))
	e := NewEngine(models.NewStoreAt(store, t.TempDir()), store)
	_, err := e.Transcribe(context.Background(), nil, "ggml-tiny")
	if err != transcription.ErrEmptyAudio {
		t.Fatalf("got %v", err)
	}
}

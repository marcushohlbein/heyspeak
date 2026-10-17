package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speek/speek/internal/holdshortcut"
)

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Get().ModelID != Default().ModelID {
		t.Fatalf("default model %q", s.Get().ModelID)
	}
	if s.Get().PushToTalkShortcut != Default().PushToTalkShortcut {
		t.Fatalf("default shortcut: got %q", s.Get().PushToTalkShortcut)
	}

	want := Config{
		PushToTalkShortcut:  "Alt+Shift+Space",
		ModelID:             "ggml-base",
		Language:            "en",
		MaxRecordingSeconds: 300,
	}
	if err := s.Set(want); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	got := s2.Get()
	if got != want {
		t.Fatalf("reload: got %+v want %+v", got, want)
	}
}

func TestStoreMissingFileUsesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("expected no file until first save")
	}
	if err := s.Set(s.Get()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestStoreLoadLegacyConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"modelId":"ggml-base"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg := s.Get()
	if !cfg.RestoreClipboard {
		t.Fatal("legacy config should default RestoreClipboard to true")
	}
	if cfg.PushToTalkShortcut != holdshortcut.DefaultAccelerator {
		t.Fatalf("shortcut %q want %q", cfg.PushToTalkShortcut, holdshortcut.DefaultAccelerator)
	}
	if cfg.MaxRecordingSeconds != 300 {
		t.Fatalf("max recording seconds %d want 300", cfg.MaxRecordingSeconds)
	}
	if cfg.ModelID != "ggml-base" {
		t.Fatalf("model id %q want ggml-base", cfg.ModelID)
	}

	explicit := filepath.Join(dir, "explicit.json")
	if err := os.WriteFile(explicit, []byte(`{"modelId":"ggml-base","restoreClipboard":false}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(explicit)
	if err != nil {
		t.Fatal(err)
	}
	if s2.Get().RestoreClipboard {
		t.Fatal("explicit restoreClipboard:false must be preserved")
	}
}

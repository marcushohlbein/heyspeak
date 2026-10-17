package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/speek/speek/internal/holdshortcut"
)

const fileName = "config.json"

// DefaultModelID is the fallback model catalog entry (mirrored as
// models.DefaultModelID for callers that speak in catalog terms).
const DefaultModelID = "ggml-tiny"

// Config is persisted user settings.
type Config struct {
	PushToTalkShortcut  string `json:"pushToTalkShortcut"`
	ModelID             string `json:"modelId"`
	Language            string `json:"language"`
	MaxRecordingSeconds int    `json:"maxRecordingSeconds"`
	InputDeviceID       string `json:"inputDeviceId"`
	// RestoreClipboard restores plain-text clipboard contents after a successful paste.
	RestoreClipboard bool `json:"restoreClipboard"`
}

// Default returns factory defaults aligned with Phase 1 spikes.
func Default() Config {
	return Config{
		PushToTalkShortcut:  holdshortcut.DefaultAccelerator,
		ModelID:             DefaultModelID,
		Language:            "",
		MaxRecordingSeconds: 300,
		RestoreClipboard:    true,
	}
}

// Store loads and saves configuration under the speek config directory.
type Store struct {
	mu   sync.Mutex
	path string
	cfg  Config
}

// OpenDefault opens or creates the store in the platform user config directory.
func OpenDefault() (*Store, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	return Open(filepath.Join(dir, fileName))
}

// Open uses an explicit config file path (for tests).
func Open(path string) (*Store, error) {
	s := &Store{path: path, cfg: Default()}
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return s, nil
}

// Path returns the on-disk JSON path.
func (s *Store) Path() string { return s.path }

// Get returns a copy of the current configuration.
func (s *Store) Get() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

// Set replaces the configuration and persists it.
func (s *Store) Set(cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = normalize(cfg)
	return s.saveLocked()
}

// Reload re-reads the configuration from disk, replacing the in-memory copy.
// Callers merge their own field values back on top before saving.
func (s *Store) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("config: parse %s: %w", s.path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("config: parse %s: %w", s.path, err)
	}
	if _, ok := raw["restoreClipboard"]; !ok {
		cfg.RestoreClipboard = true
	}
	s.cfg = normalize(cfg)
	return nil
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func normalize(cfg Config) Config {
	def := Default()
	if cfg.PushToTalkShortcut == "" {
		cfg.PushToTalkShortcut = def.PushToTalkShortcut
	}
	if cfg.MaxRecordingSeconds <= 0 {
		cfg.MaxRecordingSeconds = def.MaxRecordingSeconds
	}
	return cfg
}

// ConfigDir is the directory for speek configuration files.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "speek"), nil
}

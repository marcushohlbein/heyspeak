package models

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/speek/speek/internal/config"
)

// Store manages the model catalog and local installations.
type Store struct {
	cfg *config.Store
	// dataRoot overrides the config directory parent for tests; empty uses config.ConfigDir().
	dataRoot string

	mu        sync.Mutex
	downloads map[string]context.CancelFunc
	// verified caches successful SHA-256 checks keyed by absolute path, so
	// EffectiveModelID/ModelPath skip re-hashing when size+mtime are unchanged.
	verified map[string]verifiedFile
}

// verifiedFile records the stat identity of a file whose digest was confirmed.
type verifiedFile struct {
	size  int64
	mtime time.Time
}

// NewStore creates a model manager backed by the application config and user-data directory.
func NewStore(cfg *config.Store) *Store {
	return &Store{cfg: cfg, downloads: map[string]context.CancelFunc{}, verified: map[string]verifiedFile{}}
}

// NewStoreAt creates a store with models under dataRoot (for tests).
func NewStoreAt(cfg *config.Store, dataRoot string) *Store {
	return &Store{cfg: cfg, dataRoot: dataRoot, downloads: map[string]context.CancelFunc{}, verified: map[string]verifiedFile{}}
}

func (s *Store) SelectedModelID() string {
	return s.cfg.Get().ModelID
}

func (s *Store) SetSelectedModelID(id string) {
	cfg := s.cfg.Get()
	cfg.ModelID = id
	_ = s.cfg.Set(cfg)
}

func (s *Store) ListInstalled(ctx context.Context) ([]Model, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	status, err := s.ListCatalog(ctx)
	if err != nil {
		return nil, err
	}
	var out []Model
	for _, st := range status {
		if !st.Installed {
			continue
		}
		out = append(out, Model{ID: st.Entry.ID, Name: st.Entry.Name})
	}
	return out, nil
}

func (s *Store) ListCatalog(ctx context.Context) ([]EntryStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m, err := loadManifest(s.dataRoot)
	if err != nil {
		return nil, err
	}
	out := make([]EntryStatus, 0, len(DefaultCatalog()))
	for _, entry := range DefaultCatalog() {
		st := EntryStatus{Entry: entry}
		rec, ok := m.Installed[entry.ID]
		if !ok {
			out = append(out, st)
			continue
		}
		if err := s.verifyInstalledRecord(rec, entry.SHA256); err != nil {
			out = append(out, st)
			continue
		}
		st.Installed = true
		st.Path = rec.Path
		out = append(out, st)
	}
	return out, nil
}

func (s *Store) Download(ctx context.Context, id string, progress func(DownloadProgress)) error {
	entry, ok := catalogByID(id)
	if !ok {
		return ErrUnknownModel
	}
	if path, err := s.ModelPath(id); err == nil {
		if _, statErr := os.Stat(path); statErr == nil {
			return ErrAlreadyInstalled
		}
	}

	s.mu.Lock()
	if s.downloads[id] != nil {
		s.mu.Unlock()
		return ErrDownloadInFlight
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.downloads[id] = cancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.downloads, id)
		s.mu.Unlock()
	}()

	temp, err := tempDownloadPathFor(s.dataRoot, id)
	if err != nil {
		return err
	}
	final, err := finalModelPathFor(s.dataRoot, id)
	if err != nil {
		return err
	}

	if err := downloadToTemp(runCtx, entry, temp, progress); err != nil {
		_ = os.Remove(temp)
		return err
	}
	if err := installVerifiedAt(s.dataRoot, temp, final, entry); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}

func (s *Store) CancelDownload(id string) {
	s.mu.Lock()
	cancel := s.downloads[id]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *Store) Remove(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := catalogByID(id); !ok {
		return ErrUnknownModel
	}
	m, err := loadManifest(s.dataRoot)
	if err != nil {
		return err
	}
	rec, ok := m.Installed[id]
	if !ok {
		return ErrNotInstalled
	}
	if rec.Path != "" {
		_ = os.Remove(rec.Path)
	}
	delete(m.Installed, id)
	if err := saveManifest(s.dataRoot, m); err != nil {
		return err
	}
	if s.SelectedModelID() == id {
		s.SetSelectedModelID("")
	}
	return nil
}

func (s *Store) EffectiveModelID(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	selected := strings.TrimSpace(s.SelectedModelID())
	if selected != "" {
		if _, err := s.ModelPath(selected); err != nil {
			return "", err
		}
		return selected, nil
	}
	installed, err := s.ListInstalled(ctx)
	if err != nil {
		return "", err
	}
	switch len(installed) {
	case 0:
		return "", ErrNotInstalled
	case 1:
		id := installed[0].ID
		s.SetSelectedModelID(id)
		return id, nil
	default:
		return "", ErrNoSelection
	}
}

// verifiedCached reports whether path was previously verified and its size+mtime are unchanged.
func (s *Store) verifiedCached(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.verified[path]
	return ok && rec.size == info.Size() && rec.mtime == info.ModTime()
}

// markVerified records the current stat identity of a digest-verified file.
func (s *Store) markVerified(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.verified[path] = verifiedFile{size: info.Size(), mtime: info.ModTime()}
	s.mu.Unlock()
}

// verifyCached skips SHA-256 re-hashing when the file's size+mtime match the
// last successful verification; otherwise it re-verifies and updates the cache.
func (s *Store) verifyCached(path, wantSHA string) error {
	if s.verifiedCached(path) {
		return nil
	}
	if err := verifyFileSHA256(path, wantSHA); err != nil {
		return err
	}
	s.markVerified(path)
	return nil
}

func (s *Store) ModelPath(id string) (string, error) {
	entry, ok := catalogByID(id)
	if !ok {
		return "", ErrUnknownModel
	}
	m, err := loadManifest(s.dataRoot)
	if err != nil {
		return "", err
	}
	rec, ok := m.Installed[id]
	if !ok {
		return "", ErrNotInstalled
	}
	if err := s.verifyInstalledRecord(rec, entry.SHA256); err != nil {
		return "", err
	}
	return rec.Path, nil
}

func (s *Store) verifyInstalledRecord(rec installedRecord, catalogSHA string) error {
	path := rec.Path
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	if rec.SHA256 != "" {
		if err := s.verifyCached(path, rec.SHA256); err == nil {
			return nil
		}
	}
	return s.verifyCached(path, catalogSHA)
}

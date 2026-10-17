package models

import "context"

// Model describes an installed transcription model file.
type Model struct {
	ID   string
	Name string
}

// EntryStatus combines catalog metadata with installation state.
type EntryStatus struct {
	Entry     CatalogEntry
	Installed bool
	Path      string
}

// Manager lists, downloads, and selects local transcription models.
type Manager interface {
	SelectedModelID() string
	SetSelectedModelID(id string)
	ListInstalled(ctx context.Context) ([]Model, error)
	ListCatalog(ctx context.Context) ([]EntryStatus, error)
	Download(ctx context.Context, id string, progress func(DownloadProgress)) error
	CancelDownload(id string)
	Remove(ctx context.Context, id string) error
	// ModelPath returns the on-disk path for an installed model.
	ModelPath(id string) (string, error)
	// EffectiveModelID returns the model to use for dictation (selected or sole install).
	EffectiveModelID(ctx context.Context) (string, error)
}

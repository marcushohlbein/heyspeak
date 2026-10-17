package models

import "context"

// FakeManager is a test double for Manager.
type FakeManager struct {
	Selected string
	Models   []Model
	Catalog  []EntryStatus
	Err      error
}

func (f *FakeManager) SelectedModelID() string { return f.Selected }

func (f *FakeManager) SetSelectedModelID(id string) { f.Selected = id }

func (f *FakeManager) ListInstalled(context.Context) ([]Model, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Models, nil
}

func (f *FakeManager) ListCatalog(context.Context) ([]EntryStatus, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Catalog, nil
}

func (f *FakeManager) Download(context.Context, string, func(DownloadProgress)) error {
	return f.Err
}

func (f *FakeManager) CancelDownload(string) {}

func (f *FakeManager) Remove(context.Context, string) error {
	return f.Err
}

func (f *FakeManager) EffectiveModelID(ctx context.Context) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}
	if f.Selected != "" {
		_, err := f.ModelPath(f.Selected)
		return f.Selected, err
	}
	if len(f.Models) == 1 {
		return f.Models[0].ID, nil
	}
	if len(f.Models) == 0 {
		return "", ErrNotInstalled
	}
	return "", ErrNoSelection
}

func (f *FakeManager) ModelPath(id string) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}
	for _, m := range f.Models {
		if m.ID == id {
			return "/fake/" + id, nil
		}
	}
	if f.Selected == id && id != "" {
		return "/fake/" + id, nil
	}
	return "", ErrNotInstalled
}

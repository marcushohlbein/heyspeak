package models

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type installedRecord struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type manifestFile struct {
	Installed map[string]installedRecord `json:"installed"`
}

func loadManifest(root string) (manifestFile, error) {
	path, err := manifestPathFor(root)
	if err != nil {
		return manifestFile{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestFile{Installed: map[string]installedRecord{}}, nil
		}
		return manifestFile{}, err
	}
	var m manifestFile
	if err := json.Unmarshal(data, &m); err != nil {
		return manifestFile{}, err
	}
	if m.Installed == nil {
		m.Installed = map[string]installedRecord{}
	}
	return m, nil
}

func saveManifest(root string, m manifestFile) error {
	path, err := manifestPathFor(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

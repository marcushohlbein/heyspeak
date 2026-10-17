package models

import (
	"path/filepath"

	"github.com/speek/speek/internal/config"
)

func installDirFor(root string) (string, error) {
	if root != "" {
		return filepath.Join(root, "models"), nil
	}
	base, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "models"), nil
}

func downloadDirFor(root string) (string, error) {
	base, err := installDirFor(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ".downloads"), nil
}

func manifestPathFor(root string) (string, error) {
	base, err := installDirFor(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "installed.json"), nil
}

func finalModelPathFor(root, id string) (string, error) {
	dir, err := installDirFor(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".bin"), nil
}

func tempDownloadPathFor(root, id string) (string, error) {
	dir, err := downloadDirFor(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".part"), nil
}

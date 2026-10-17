package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadProgress reports byte progress for a model download.
type DownloadProgress struct {
	ModelID  string
	Received int64
	Total    int64
	Complete bool
}

func downloadToTemp(ctx context.Context, entry CatalogEntry, dest string, onProgress func(DownloadProgress)) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	_ = os.Remove(dest)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.URL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("models: download HTTP %s", resp.Status)
	}

	total := entry.SizeBytes
	if cl := resp.ContentLength; cl > 0 {
		total = cl
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	var received int64
	buf := make([]byte, 32*1024)
	for {
		if ctx.Err() != nil {
			_ = os.Remove(dest)
			return ctx.Err()
		}
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				_ = os.Remove(dest)
				return werr
			}
			received += int64(n)
			if onProgress != nil {
				onProgress(DownloadProgress{
					ModelID:  entry.ID,
					Received: received,
					Total:    total,
				})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = os.Remove(dest)
			return readErr
		}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(dest)
		return err
	}
	if onProgress != nil {
		onProgress(DownloadProgress{ModelID: entry.ID, Received: received, Total: total, Complete: true})
	}
	return nil
}

func verifyFileSHA256(path string, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != want {
		return fmt.Errorf("%w: expected %s got %s", ErrChecksumMismatch, want, got)
	}
	return nil
}

func installVerifiedAt(root, tempPath, finalPath string, entry CatalogEntry) error {
	if err := verifyFileSHA256(tempPath, entry.SHA256); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		if err := copyFile(tempPath, finalPath); err != nil {
			return err
		}
		_ = os.Remove(tempPath)
	}
	info, err := os.Stat(finalPath)
	if err != nil {
		return err
	}
	m, err := loadManifest(root)
	if err != nil {
		return err
	}
	m.Installed[entry.ID] = installedRecord{
		Path:   finalPath,
		SHA256: entry.SHA256,
		Size:   info.Size(),
	}
	return saveManifest(root, m)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

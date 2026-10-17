package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/speek/speek/internal/config"
)

func testStore(t *testing.T) *Store {
	dir := t.TempDir()
	cfgStore, err := config.Open(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	return NewStoreAt(cfgStore, dir)
}

func TestDownloadVerifyInstall(t *testing.T) {
	payload := []byte("speek-test-model-bytes")
	sum := sha256.Sum256(payload)
	digest := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	dir := t.TempDir()
	entry := CatalogEntry{
		ID:        "ggml-tiny",
		Name:      "Test",
		URL:       srv.URL,
		SizeBytes: int64(len(payload)),
		SHA256:    digest,
	}

	temp, err := tempDownloadPathFor(dir, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	final, err := finalModelPathFor(dir, entry.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := downloadToTemp(context.Background(), entry, temp, nil); err != nil {
		t.Fatal(err)
	}
	if err := installVerifiedAt(dir, temp, final, entry); err != nil {
		t.Fatal(err)
	}
	if err := verifyFileSHA256(final, digest); err != nil {
		t.Fatal(err)
	}
}

func TestCorruptFileNotInstalled(t *testing.T) {
	store := testStore(t)
	entry := DefaultCatalog()[0]
	final, err := finalModelPathFor(store.dataRoot, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(final, []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(store.dataRoot)
	if err != nil {
		t.Fatal(err)
	}
	m.Installed[entry.ID] = installedRecord{Path: final, SHA256: entry.SHA256, Size: 7}
	if err := saveManifest(store.dataRoot, m); err != nil {
		t.Fatal(err)
	}
	cat, err := store.ListCatalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range cat {
		if st.Entry.ID == entry.ID && st.Installed {
			t.Fatal("corrupt file marked installed")
		}
	}
}

func TestDuplicateDownloadBlocked(t *testing.T) {
	store := testStore(t)
	store.mu.Lock()
	store.downloads["ggml-tiny"] = func() {}
	store.mu.Unlock()
	if err := store.Download(context.Background(), "ggml-tiny", nil); err != ErrDownloadInFlight {
		t.Fatalf("got %v", err)
	}
}

func TestCatalogListsAvailableAndInstalled(t *testing.T) {
	store := testStore(t)
	cat, err := store.ListCatalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cat) != len(DefaultCatalog()) {
		t.Fatalf("catalog len %d", len(cat))
	}
	for _, st := range cat {
		if st.Installed {
			t.Fatal("expected none installed")
		}
	}
}

func TestChecksumMismatchRejected(t *testing.T) {
	dir := t.TempDir()
	entry := DefaultCatalog()[0]
	temp := filepath.Join(dir, "x.part")
	final := filepath.Join(dir, "x.bin")
	if err := os.WriteFile(temp, []byte("wrong"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installVerifiedAt(dir, temp, final, entry); err == nil {
		t.Fatal("expected checksum error")
	}
}

// installFakeModel writes a small payload with a matching digest into the
// store's manifest so the model counts as installed.
func installFakeModel(t *testing.T, store *Store, id string) {
	t.Helper()
	payload := []byte("speek-fake-model-" + id)
	sum := sha256.Sum256(payload)
	entry := CatalogEntry{
		ID:        id,
		URL:       "unused://local",
		SizeBytes: int64(len(payload)),
		SHA256:    hex.EncodeToString(sum[:]),
	}
	temp, err := tempDownloadPathFor(store.dataRoot, id)
	if err != nil {
		t.Fatal(err)
	}
	final, err := finalModelPathFor(store.dataRoot, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(temp), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(temp, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installVerifiedAt(store.dataRoot, temp, final, entry); err != nil {
		t.Fatal(err)
	}
}

// ponytail: Store verifies installs against hardcoded real catalog SHA256
// digests and offers no catalog seam, so auto-select with a fake install is
// not reachable in a unit test; only the manifest-level Remove path is.

func TestRemoveClearsSelection(t *testing.T) {
	store := testStore(t)
	installFakeModel(t, store, "ggml-tiny")
	store.SetSelectedModelID("ggml-tiny")

	if err := store.Remove(context.Background(), "ggml-tiny"); err != nil {
		t.Fatal(err)
	}
	if got := store.SelectedModelID(); got != "" {
		t.Fatalf("selection %q not cleared after remove", got)
	}
	if _, err := store.EffectiveModelID(context.Background()); err == nil {
		t.Fatal("expected error after removing the selected model, got stale id")
	}
}

// stubHTTP points http.DefaultClient at a canned round-trip func so
// Store.Download's hardcoded catalog URLs stay offline in tests.
func stubHTTP(t *testing.T, rt func(*http.Request) (*http.Response, error)) {
	t.Helper()
	orig := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(rt)}
	t.Cleanup(func() { http.DefaultClient = orig })
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type dripReader struct {
	data []byte
	pos  int
}

func (d *dripReader) Read(p []byte) (int, error) {
	if d.pos >= len(d.data) {
		return 0, io.EOF
	}
	time.Sleep(5 * time.Millisecond)
	p[0] = d.data[d.pos]
	d.pos++
	return 1, nil
}

func (d *dripReader) Close() error { return nil }

func TestDownloadHTTPFailure(t *testing.T) {
	store := testStore(t)
	stubHTTP(t, func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Status:     "500 Internal Server Error",
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("boom")),
		}, nil
	})

	if err := store.Download(context.Background(), "ggml-tiny", nil); err == nil {
		t.Fatal("expected download error on HTTP 500")
	}
	part, err := tempDownloadPathFor(store.dataRoot, "ggml-tiny")
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(part); statErr == nil {
		t.Fatal("temp .part file left behind after failure")
	}
	retryErr := store.Download(context.Background(), "ggml-tiny", nil)
	if errors.Is(retryErr, ErrDownloadInFlight) {
		t.Fatal("retry blocked after failed download")
	}
	if retryErr == nil {
		t.Fatal("expected retry to fail again with HTTP 500")
	}
}

func TestDownloadCancelMidStream(t *testing.T) {
	store := testStore(t)
	payload := []byte("slow-drip-model-bytes")
	stubHTTP(t, func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{},
			Body:       &dripReader{data: payload},
		}, nil
	})

	errc := make(chan error, 1)
	go func() {
		errc <- store.Download(context.Background(), "ggml-tiny", nil)
	}()

	part, err := tempDownloadPathFor(store.dataRoot, "ggml-tiny")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, statErr := os.Stat(part); statErr == nil {
			break
		}
		time.Sleep(time.Millisecond)
	}
	store.CancelDownload("ggml-tiny")

	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("download did not return after cancel")
	}
	if _, statErr := os.Stat(part); statErr == nil {
		t.Fatal("temp file not removed after cancel")
	}
	m, err := loadManifest(store.dataRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Installed) != 0 {
		t.Fatal("manifest has an entry after cancelled download")
	}
}

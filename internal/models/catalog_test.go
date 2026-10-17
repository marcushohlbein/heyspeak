package models

import "testing"

func TestCatalogDigestIsContentSHA256(t *testing.T) {
	// Hugging Face also publishes an Xet hash. The installer compares file bytes.
	e, ok := catalogByID("ggml-tiny")
	if !ok {
		t.Fatal("missing ggml-tiny")
	}
	const content = "be07e048e1e599ad46341c8d2a135645097a538221678b7acdd1b1919c6e1b21"
	if e.SHA256 != content {
		t.Fatalf("SHA256 = %s, want content digest %s", e.SHA256, content)
	}
}

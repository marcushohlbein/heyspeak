package models

import "github.com/speek/speek/internal/config"

// Engine identifiers for catalog entries.
const EngineWhisperCpp = "whisper.cpp"

// DefaultModelID is the catalog entry used when no model is configured.
// It aliases config.DefaultModelID; models cannot own it because config
// must not import models (config is the lower layer).
const DefaultModelID = config.DefaultModelID

// CatalogEntry describes a downloadable model artifact.
type CatalogEntry struct {
	ID           string
	Name         string
	Engine       string
	Languages    []string
	URL          string
	SizeBytes    int64
	SHA256       string
	Format       string
	License      string
	Attribution  string
	Requirements string
}

// DefaultCatalog returns built-in models with digests from Hugging Face LFS metadata
// (ggerganov/whisper.cpp, main branch).
func DefaultCatalog() []CatalogEntry {
	return []CatalogEntry{
		{
			ID:           "ggml-tiny",
			Name:         "Whisper Tiny (multilingual)",
			Engine:       EngineWhisperCpp,
			Languages:    []string{"auto", "en", "de"},
			URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
			SizeBytes:    77_691_713,
			SHA256:       "be07e048e1e599ad46341c8d2a135645097a538221678b7acdd1b1919c6e1b21",
			Format:       "ggml",
			License:      "MIT (model weights subject to OpenAI Whisper terms)",
			Attribution:  "OpenAI Whisper; ggml format by whisper.cpp",
			Requirements: "~1 GB RAM; suitable for quick tests",
		},
		{
			ID:           "ggml-base",
			Name:         "Whisper Base (multilingual)",
			Engine:       EngineWhisperCpp,
			Languages:    []string{"auto", "en", "de"},
			URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
			SizeBytes:    147_951_465,
			SHA256:       "60ed5bc3dd14eea856493d334349b405782ddcaf0028d4b5df4088345fba2efe",
			Format:       "ggml",
			License:      "MIT (model weights subject to OpenAI Whisper terms)",
			Attribution:  "OpenAI Whisper; ggml format by whisper.cpp",
			Requirements: "~1.5 GB RAM; better accuracy than Tiny",
		},
	}
}

func catalogByID(id string) (CatalogEntry, bool) {
	for _, e := range DefaultCatalog() {
		if e.ID == id {
			return e, true
		}
	}
	return CatalogEntry{}, false
}

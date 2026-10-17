package models

import "github.com/speek/speek/internal/config"

// DefaultModelID is the catalog entry used when no model is configured.
// It aliases config.DefaultModelID; models cannot own it because config
// must not import models (config is the lower layer).
const DefaultModelID = config.DefaultModelID

const whisperModelBaseURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/"

// CatalogEntry describes a downloadable model artifact.
// Weights are MIT, subject to the OpenAI Whisper license; ggml builds are from whisper.cpp.
type CatalogEntry struct {
	ID        string
	Name      string
	URL       string
	SizeBytes int64
	SHA256    string
}

func whisperEntry(id, name, sha string, size int64) CatalogEntry {
	return CatalogEntry{
		ID:        id,
		Name:      name,
		URL:       whisperModelBaseURL + id + ".bin",
		SizeBytes: size,
		SHA256:    sha,
	}
}

// DefaultCatalog returns built-in models. SHA256 is the file's content digest
// (Hugging Face LFS oid), not the Xet hash.
func DefaultCatalog() []CatalogEntry {
	return []CatalogEntry{
		whisperEntry("ggml-tiny", "Whisper Tiny", "be07e048e1e599ad46341c8d2a135645097a538221678b7acdd1b1919c6e1b21", 77_691_713),
		whisperEntry("ggml-tiny.en", "Whisper Tiny", "921e4cf8686fdd993dcd081a5da5b6c365bfde1162e72b08d75ac75289920b1f", 77_704_715),
		whisperEntry("ggml-base", "Whisper Base", "60ed5bc3dd14eea856493d334349b405782ddcaf0028d4b5df4088345fba2efe", 147_951_465),
		whisperEntry("ggml-base.en", "Whisper Base", "a03779c86df3323075f5e796cb2ce5029f00ec8869eee3fdfb897afe36c6d002", 147_964_211),
		whisperEntry("ggml-small", "Whisper Small", "1be3a9b2063867b937e64e2ec7483364a79917e157fa98c5d94b5c1fffea987b", 487_601_967),
		whisperEntry("ggml-small.en", "Whisper Small", "c6138d6d58ecc8322097e0f987c32f1be8bb0a18532a3f88f734d1bbf9c41e5d", 487_614_201),
		whisperEntry("ggml-medium", "Whisper Medium", "6c14d5adee5f86394037b4e4e8b59f1673b6cee10e3cf0b11bbdbee79c156208", 1_533_763_059),
		whisperEntry("ggml-medium.en", "Whisper Medium", "cc37e93478338ec7700281a7ac30a10128929eb8f427dda2e865faa8f6da4356", 1_533_774_781),
		whisperEntry("ggml-large-v3-turbo", "Whisper Large v3 Turbo", "1fc70f774d38eb169993ac391eea357ef47c88757ef72ee5943879b7e8e2bc69", 1_624_555_275),
		whisperEntry("ggml-large-v3", "Whisper Large v3", "64d182b440b98d5203c4f9bd541544d84c605196c4f7b845dfa11fb23594d1e2", 3_095_033_483),
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

// LookupCatalogEntry returns a built-in catalog entry by model ID.
func LookupCatalogEntry(id string) (CatalogEntry, bool) {
	return catalogByID(id)
}

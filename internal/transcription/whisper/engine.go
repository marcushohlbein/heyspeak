package whisper

import (
	"sync"

	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
)

// Engine transcribes audio in-process with whisper.cpp (linked at build time).
// Models are downloaded and managed by speek; run `make whisper` once to build native libs.
type Engine struct {
	models models.Manager
	cfg    *config.Store

	mu         sync.Mutex
	loadedPath string
	model      modelHandle
}

// modelHandle is implemented by the CGO whisper model; nil when !cgo.
type modelHandle interface {
	Close() error
}

// NewEngine creates an in-process whisper.cpp engine.
func NewEngine(m models.Manager, cfg *config.Store) *Engine {
	ensureRuntimeEnv()
	return &Engine{models: m, cfg: cfg}
}

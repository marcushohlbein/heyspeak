package app

import (
	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
	"github.com/speek/speek/internal/overlay"
	"github.com/speek/speek/internal/paste"
	"github.com/speek/speek/internal/shortcut"
	"github.com/speek/speek/internal/transcription"
	"github.com/speek/speek/internal/transcription/whisper"
)

// Services groups injectable application dependencies.
type Services struct {
	Shortcut      shortcut.Monitor
	Overlay       overlay.Controller
	Audio         audio.Capturer
	Transcription transcription.Engine
	Paste         paste.Service
	Models        models.Manager
}

// DefaultServices wires production implementations for the desktop shell.
func DefaultServices(store *config.Store) Services {
	modelsMgr := models.NewStore(store)
	overlayMgr := overlay.NewManager()
	rec := audio.NewRecorder(store)
	rec.SetOnLevel(overlayMgr.OnMicPeak)
	return Services{
		Shortcut:      shortcut.NewHoldMonitor(),
		Overlay:       overlayMgr,
		Audio:         rec,
		Transcription: whisper.NewEngine(modelsMgr, store),
		Paste:         paste.Default(store),
		Models:        modelsMgr,
	}
}

package transcription

import (
	"context"
	"errors"
)

// Engine transcribes captured PCM16 mono audio.
type Engine interface {
	Transcribe(ctx context.Context, samples []int16, modelID string) (string, error)
}

// ErrUnavailable is returned when the local engine cannot be used (missing build or model).
var ErrUnavailable = errors.New("transcription: engine not configured")

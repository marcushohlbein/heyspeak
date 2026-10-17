//go:build !cgo

package whisper

import (
	"context"
	"fmt"

	"github.com/speek/speek/internal/transcription"
)

// Close is a no-op when speek is built without CGO.
func (e *Engine) Close() error { return nil }

// Transcribe is unavailable unless speek is built with CGO and whisper.cpp libraries.
func (e *Engine) Transcribe(context.Context, []int16, string) (string, error) {
	return "", fmt.Errorf("%w: rebuild with CGO enabled and run `make whisper`", transcription.ErrUnavailable)
}

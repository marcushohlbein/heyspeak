//go:build cgo

package whisper

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	w "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"

	"github.com/speek/speek/internal/transcription"
)

// Close releases the loaded whisper model, if any.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.model != nil {
		err := e.model.Close()
		e.model = nil
		e.loadedPath = ""
		return err
	}
	return nil
}

// Transcribe runs whisper.cpp on PCM samples using the installed model file.
func (e *Engine) Transcribe(ctx context.Context, samples []int16, modelID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(samples) == 0 {
		return "", transcription.ErrEmptyAudio
	}
	if len(samples) < minSpeechSamples {
		return "", transcription.ErrNoSpeech
	}

	modelPath, err := e.models.ModelPath(modelID)
	if err != nil {
		return "", err
	}

	model, err := e.loadModel(modelPath)
	if err != nil {
		return "", fmt.Errorf("%w: run `make whisper` from the speek repo to build whisper.cpp, then rebuild speek", transcription.ErrUnavailable)
	}

	wctx, err := model.NewContext()
	if err != nil {
		return "", err
	}

	lang := strings.TrimSpace(e.cfg.Get().Language)
	if lang == "" {
		lang = "auto"
	}
	if lang != "auto" || model.IsMultilingual() {
		if err := wctx.SetLanguage(lang); err != nil && lang != "auto" {
			return "", err
		}
	}
	wctx.SetTokenTimestamps(false)

	peak := peakAbsPCM(samples)
	boosted := boostQuietPCM(samples, peak)
	data := pcm16ToFloat32(boosted)
	encBegin := func() bool { return ctx.Err() == nil }
	if err := wctx.Process(data, encBegin, nil, nil); err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", err
	}

	var b strings.Builder
	for {
		seg, err := wctx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if s := strings.TrimSpace(seg.Text); s != "" {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(s)
		}
	}
	text := strings.TrimSpace(b.String())
	if err := classifyTranscript(text, peak); err != nil {
		log.Printf("transcription: %s (%.0f ms, peak %d, lang %s)", err, float64(len(samples))/16000*1000, peak, wctx.DetectedLanguage())
		return "", err
	}
	log.Printf("transcription: %d characters, lang %s, peak %d", len(text), wctx.DetectedLanguage(), peak)
	return text, nil
}

func (e *Engine) loadModel(path string) (w.Model, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.model != nil && e.loadedPath == path {
		return e.model.(w.Model), nil
	}
	if e.model != nil {
		_ = e.model.Close()
		e.model = nil
		e.loadedPath = ""
	}
	model, err := w.New(path)
	if err != nil {
		return nil, err
	}
	e.model = model
	e.loadedPath = path
	return model, nil
}

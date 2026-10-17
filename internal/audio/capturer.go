package audio

import "context"

// SampleRate is the capture rate used for transcription adapters (mono PCM16).
const SampleRate = 16000

// Capturer records microphone audio in sessions (start on PTT down, stop on release).
type Capturer interface {
	Start(ctx context.Context) error
	Stop() ([]int16, error)
}

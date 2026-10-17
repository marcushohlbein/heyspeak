package whisper

import (
	"strings"
	"unicode"

	"github.com/speek/speek/internal/transcription"
)

func classifyTranscript(text string, peak int) error {
	text = strings.TrimSpace(text)
	if text == "" {
		if peak >= 500 {
			return transcription.ErrUnintelligible
		}
		return transcription.ErrNoSpeech
	}
	// A real recording can be a short word. Only drop silence hallucinations
	// when the microphone itself was quiet.
	if peak >= 500 {
		return nil
	}
	lower := strings.ToLower(strings.Trim(text, ".,!? "))
	phrases := []string{
		"thank you",
		"thanks for watching",
		"subtitle",
		"you",
		"bye",
		"silence",
		"[blank_audio]",
		"(silence)",
	}
	for _, p := range phrases {
		if lower == p {
			return transcription.ErrNoSpeech
		}
	}
	if len(lower) < 3 && !hasLetter(lower) {
		return transcription.ErrNoSpeech
	}
	return nil
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

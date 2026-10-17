package whisper

import (
	"testing"

	"github.com/speek/speek/internal/transcription"
)

func TestClassifySilenceHallucination(t *testing.T) {
	if err := classifyTranscript("Thank you.", 10); err == nil {
		t.Fatal("expected no speech")
	}
	if err := classifyTranscript("Hello world.", 10); err != nil {
		t.Fatal(err)
	}
	if err := classifyTranscript("you", 4000); err != nil {
		t.Fatal(err)
	}
	if err := classifyTranscript("", 4000); err != transcription.ErrUnintelligible {
		t.Fatal(err)
	}
}

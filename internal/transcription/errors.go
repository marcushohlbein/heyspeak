package transcription

import "errors"

// ErrNoSpeech indicates the engine did not detect spoken content.
var ErrNoSpeech = errors.New("transcription: no speech detected")

// ErrEmptyAudio is returned when there is nothing to transcribe.
var ErrEmptyAudio = errors.New("transcription: empty audio")

// ErrUnintelligible means the microphone had signal but the model returned no words.
var ErrUnintelligible = errors.New("transcription: sound was captured but not recognized")

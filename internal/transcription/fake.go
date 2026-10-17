package transcription

import "context"

// FakeEngine returns fixed text for tests.
type FakeEngine struct {
	Text string
	Err  error
}

func (f FakeEngine) Transcribe(context.Context, []int16, string) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}
	return f.Text, nil
}

package paste

import "context"

// FakeService records pasted text for tests.
type FakeService struct {
	Last     string
	Err      error
	Captures int
}

func (f *FakeService) CapturePasteTarget() {
	f.Captures++
}

func (f *FakeService) Paste(_ context.Context, text string) error {
	if f.Err != nil {
		return f.Err
	}
	f.Last = text
	return nil
}

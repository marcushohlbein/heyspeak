package overlay

// FakeController records overlay calls for tests.
type FakeController struct {
	Visible      bool
	Transcribing bool
}

func (f *FakeController) PrepareSession() {}

func (f *FakeController) ShowRecording() {
	f.Visible = true
	f.Transcribing = false
}

func (f *FakeController) ShowTranscribing() {
	f.Visible = true
	f.Transcribing = true
}

func (f *FakeController) Hide() {
	f.Visible = false
	f.Transcribing = false
}

func (f *FakeController) Close() {
	f.Visible = false
	f.Transcribing = false
}

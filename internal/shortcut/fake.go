package shortcut

import "fmt"

// FakeMonitor records shortcut lifecycle for tests.
type FakeMonitor struct {
	Accelerator string
	Events      Events
	Started     bool
}

func (f *FakeMonitor) Start(accelerator string, events Events) error {
	if f.Started {
		return fmt.Errorf("fake shortcut: already started")
	}
	f.Accelerator = accelerator
	f.Events = events
	f.Started = true
	return nil
}

func (f *FakeMonitor) Stop() error {
	f.Started = false
	return nil
}

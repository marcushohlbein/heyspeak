//go:build !darwin

package holdshortcut

type stubMonitor struct{}

func newMonitor() Monitor { return stubMonitor{} }

func (stubMonitor) Start(string, Listener) error { return ErrUnsupported }

func (stubMonitor) Stop() error { return nil }

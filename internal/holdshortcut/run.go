//go:build !darwin

package holdshortcut

// Run blocks until Stop is called on the monitor's run loop (platform-specific).
func Run(m Monitor) error {
	return ErrUnsupported
}

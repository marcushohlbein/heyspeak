//go:build darwin

package holdshortcut

// Run blocks until Stop is called on the monitor's run loop.
func Run(m Monitor) error {
	return runMonitor(m)
}

package paste

type focusState struct {
	frontmost int
	own       int
	target    int
}

// resolveFocusForPaste decides whether to activate the captured target before posting paste.
// ok is false when a third application is focused (do not paste or activate).
// target == -1 means focus capture failed (frontmostPID unavailable): refuse to
// paste into an unverified target unless speek itself is frontmost.
func resolveFocusForPaste(s focusState) (activateTarget bool, ok bool) {
	if s.target == -1 && s.frontmost != s.own {
		return false, false
	}
	if s.target <= 0 {
		return false, true
	}
	if s.frontmost == s.target {
		return false, true
	}
	if s.frontmost == s.own {
		return true, true
	}
	return false, false
}

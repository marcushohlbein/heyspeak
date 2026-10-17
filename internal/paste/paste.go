package paste

// PostPasteChord posts a Cmd+V (or platform equivalent) so the focused app inserts clipboard text.
func PostPasteChord() error {
	return paste()
}

package paste

import (
	"github.com/egoist/mygo"
)

func readPlainClipboard() string {
	return mygo.Clipboard.ReadText()
}

func writePlainClipboard(text string) {
	mygo.Clipboard.WriteText(text)
}

// restorePlainClipboard puts prior plain text back when the clipboard still holds inserted.
func restorePlainClipboard(inserted, prior string) {
	if mygo.Clipboard.ReadText() != inserted {
		return
	}
	if prior == "" {
		mygo.Clipboard.Clear()
		return
	}
	mygo.Clipboard.WriteText(prior)
}

package overlay

import (
	"github.com/egoist/mygo"
	"github.com/speek/speek/internal/overlayglow"
)

// lockDisplay resolves the display under the pointer at recording start.
func lockDisplay() (mygo.Display, mygo.Rectangle) {
	pt := mygo.Screen.CursorScreenPoint()
	d := mygo.Screen.DisplayNearestPoint(pt)
	return d, overlayglow.DisplayRect(d)
}

func displayByID(id int64) (mygo.Display, bool) {
	for _, d := range mygo.Screen.Displays() {
		if d.ID == id {
			return d, true
		}
	}
	return mygo.Display{}, false
}

package overlayglow

import "github.com/egoist/mygo"

// DisplayRect is the overlay target in MyGo screen coordinates (NSScreen.frame).
func DisplayRect(d mygo.Display) mygo.Rectangle {
	b := d.Bounds
	return mygo.Rectangle{X: b.X, Y: b.Y, Width: b.Width, Height: b.Height}
}

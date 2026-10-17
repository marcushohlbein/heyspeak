package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"runtime"
)

// trayIcon is a simple mic shape: template (black) on macOS, blue elsewhere.
func trayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	ink := color.NRGBA{R: 79, G: 124, B: 255, A: 255}
	if runtime.GOOS == "darwin" {
		ink = color.NRGBA{A: 255}
	}
	// Mic body
	for _, r := range []image.Rectangle{
		image.Rect(12, 8, 20, 20),
		image.Rect(10, 18, 22, 22),
		image.Rect(14, 22, 18, 26),
		image.Rect(8, 26, 24, 28),
	} {
		draw.Draw(img, r, image.NewUniform(ink), image.Point{}, draw.Src)
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}

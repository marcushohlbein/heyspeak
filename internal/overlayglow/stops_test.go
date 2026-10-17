package overlayglow

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSiriGradientSeamMatches(t *testing.T) {
	stops := siriGradientStops()
	if len(stops) < 10 {
		t.Fatalf("expected dense stops, got %d", len(stops))
	}
	// mygo's painter keeps at most scene.MaxStops = 32 stops and silently
	// drops the rest, which would cut the closing seam stop.
	if len(stops) > 32 {
		t.Fatalf("gradient has %d stops; mygo scene.MaxStops caps at 32", len(stops))
	}
	if stops[0].At != 0 || stops[len(stops)-1].At != 1 {
		t.Fatalf("expected seam stops at 0 and 1")
	}
	if !colorsClose(stops[0].Color, stops[len(stops)-1].Color) {
		t.Fatalf("seam colors should match for a smooth loop")
	}
	// Positions must never go backwards. The first interior stop shares At 0
	// with the seam stop (k=1 makes them the same pure ring[0] color), so
	// equal is allowed at that pair only.
	for i := 1; i < len(stops)-1; i++ {
		if stops[i].At < stops[i-1].At {
			t.Fatalf("stops should be non-decreasing at %d", i)
		}
	}
}

func colorsClose(a, b ui.Color) bool {
	const d = 2
	return abs(int(a.R)-int(b.R)) <= d &&
		abs(int(a.G)-int(b.G)) <= d &&
		abs(int(a.B)-int(b.B)) <= d
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

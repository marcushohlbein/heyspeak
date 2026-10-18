package overlayglow

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSiriGradientSeamMatches(t *testing.T) {
	stops := siriGradientStops()
	if len(stops) > 32 {
		t.Fatalf("gradient has %d stops; max is 32", len(stops))
	}
	if stops[0].At != 0 || stops[len(stops)-1].At != 1 {
		t.Fatal("expected seam stops at 0 and 1")
	}
	if !colorsClose(stops[0].Color, stops[len(stops)-1].Color) {
		t.Fatal("seam colors should match")
	}
	for i := 1; i < len(stops)-1; i++ {
		if stops[i].At < stops[i-1].At {
			t.Fatalf("stops out of order at %d", i)
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

package overlay

import (
	"testing"
	"time"
)

func TestAdvanceOpacityReachesTarget(t *testing.T) {
	start := time.Now()
	got := advanceOpacity(start.Add(fadeInDuration), start, 0, 1, fadeInDuration, easeOutShow)
	if got < 0.99 {
		t.Fatalf("expected ~1, got %v", got)
	}
}

func TestAdvanceOpacityFastStart(t *testing.T) {
	start := time.Now()
	got := advanceOpacity(start.Add(fadeInDuration/4), start, 0, 1, fadeInDuration, easeOutShow)
	if got < 0.7 {
		t.Fatalf("expected fast entrance, got %v", got)
	}
}

func TestFadeDuration(t *testing.T) {
	d := 520 * time.Millisecond
	cases := []struct {
		name         string
		hiding       bool
		target, from float64
		reduce       bool
		want         time.Duration
	}{
		{"reduce+hiding", true, 0, 1, true, fadeOutReduced},
		{"reduce+fade in", false, 1, 0, true, fadeInReduced},
		{"reduce+neither", false, 1, 1, true, d},
		{"full motion", true, 0, 1, false, d},
	}
	for _, c := range cases {
		if got := fadeDuration(c.hiding, c.target, c.from, c.reduce, d); got != c.want {
			t.Fatalf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

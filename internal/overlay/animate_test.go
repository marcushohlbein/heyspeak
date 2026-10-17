package overlay

import (
	"testing"
	"time"
)

func TestAdvanceOpacityReachesTarget(t *testing.T) {
	start := time.Now()
	got := advanceOpacity(start.Add(fadeInDuration), start, 0, 1, fadeInDuration)
	if got < 0.99 {
		t.Fatalf("expected ~1, got %v", got)
	}
}

func TestAdvanceOpacityMidpoint(t *testing.T) {
	start := time.Now()
	mid := start.Add(fadeInDuration / 2)
	got := advanceOpacity(mid, start, 0, 1, fadeInDuration)
	if got <= 0.1 || got >= 0.9 {
		t.Fatalf("expected mid-range opacity, got %v", got)
	}
}

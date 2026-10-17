package overlay

import (
	"math"
	"time"
)

const (
	fadeInDuration  = 280 * time.Millisecond
	fadeOutDuration = 220 * time.Millisecond
	// TranscribeVisualDelay is how long processing may run before showing the subtle state.
	TranscribeVisualDelay = 300 * time.Millisecond
)

// advanceOpacity moves current toward target over the given duration.
func advanceOpacity(now, animStart time.Time, from, target float64, duration time.Duration) float64 {
	if duration <= 0 {
		return target
	}
	elapsed := now.Sub(animStart)
	if elapsed >= duration {
		return target
	}
	t := float64(elapsed) / float64(duration)
	t = easeInOut(t)
	return from + (target-from)*t
}

func easeInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

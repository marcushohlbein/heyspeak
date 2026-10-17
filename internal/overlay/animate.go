package overlay

import (
	"math"
	"time"
)

const (
	fadeInDuration  = 420 * time.Millisecond
	fadeOutDuration = 520 * time.Millisecond
	fadeInReduced   = 120 * time.Millisecond
	fadeOutReduced  = 150 * time.Millisecond
	// TranscribeVisualDelay is how long processing may run before showing the subtle state.
	TranscribeVisualDelay = 300 * time.Millisecond
)

// fadeDuration picks the fade length for the current animation, honoring
// reduced motion with shorter fixed durations.
func fadeDuration(hiding bool, target, from float64, reduceMotion bool, d time.Duration) time.Duration {
	if !reduceMotion {
		return d
	}
	if hiding {
		return fadeOutReduced
	}
	if target > from {
		return fadeInReduced
	}
	return d
}

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

package overlay

import (
	"math"
	"time"
)

const (
	fadeInDuration  = 240 * time.Millisecond
	fadeOutDuration = 520 * time.Millisecond
	fadeInReduced   = 120 * time.Millisecond
	fadeOutReduced  = 150 * time.Millisecond
	// TranscribeVisualDelay is how long processing may run before showing the subtle state.
	TranscribeVisualDelay = 300 * time.Millisecond
	transcribeDimDuration = 200 * time.Millisecond
)

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
func advanceOpacity(now, animStart time.Time, from, target float64, duration time.Duration, ease func(float64) float64) float64 {
	if duration <= 0 {
		return target
	}
	elapsed := now.Sub(animStart)
	if elapsed >= duration {
		return target
	}
	t := float64(elapsed) / float64(duration)
	t = ease(t)
	return from + (target-from)*t
}

func easeInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

func easeOutShow(t float64) float64 { return 1 - (1-t)*(1-t)*(1-t)*(1-t)*(1-t) }

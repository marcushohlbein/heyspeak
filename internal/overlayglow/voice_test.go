package overlayglow

import (
	"math"
	"testing"
	"time"
)

func TestVoicePulseQueueDeliversOnFrameClock(t *testing.T) {
	v := VoicePulse{}
	v.Queue(0.45, 0)
	v.Queue(0.45, 100*time.Millisecond)
	v.Advance(1 / 60.0)
	if v.Level() < 0.1 {
		t.Fatalf("expected first queued pulse, got %v", v.Level())
	}
	for i := 0; i < 8; i++ {
		v.Advance(1 / 60.0)
	}
	if v.Level() < 0.12 {
		t.Fatalf("expected second queued pulse, got %v", v.Level())
	}
}

func TestVoicePulseRespondsAndDecays(t *testing.T) {
	v := VoicePulse{}
	for i := 0; i < 30; i++ {
		v.Push(1)
		v.Advance(1 / 60.0)
	}
	if v.Level() < 0.3 {
		t.Fatalf("expected strong level after push, got %v", v.Level())
	}
	for i := 0; i < 120; i++ {
		v.Advance(1 / 60.0)
	}
	if v.Level() > 0.05 {
		t.Fatalf("expected decay toward silence, got %v", v.Level())
	}
}

func TestPeakToPulseLevel(t *testing.T) {
	if PeakToPulseLevel(1200) <= float32(1200)/8000 {
		t.Fatal("pulse mapping should exceed settings meter scale")
	}
	for _, c := range []struct {
		peak int
		want float32
	}{{0, 0}, {400, 0.5}, {1600, 1}} {
		if got := PeakToPulseLevel(c.peak); math.Abs(float64(got-c.want)) > 1e-6 {
			t.Fatalf("PeakToPulseLevel(%d) = %v, want %v", c.peak, got, c.want)
		}
	}
}

func TestBreathForVoiceDimsWhenSpeaking(t *testing.T) {
	for _, now := range []float64{0, 1, 3} {
		b := BreathForVoice(now, 0.6, false)
		if b.OpacityMult > 1 || b.OpacityMult < 1-0.30*0.6 {
			t.Fatalf("opacity mult %v out of range", b.OpacityMult)
		}
	}
}

func TestGradientAngleIntegratesSmoothly(t *testing.T) {
	phase := 0.0
	a := stepGradient(&phase, 1, false)
	b := stepGradient(&phase, 1, false)
	if math.Abs(float64(b-a)-gradientDegPerSec) > 0.01 {
		t.Fatalf("angle delta %v want ~%v", b-a, gradientDegPerSec)
	}
	if stepGradient(&phase, 1, true) != stepGradient(&phase, 1, true) {
		t.Fatal("reduce motion should hold angle steady")
	}
}

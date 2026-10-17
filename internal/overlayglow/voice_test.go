package overlayglow

import (
	"math"
	"testing"
)

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

func TestPeakToPulseLevelMoreSensitiveThanMeter(t *testing.T) {
	const peak = 1200
	meter := float32(peak) / 8000
	pulse := PeakToPulseLevel(peak)
	if pulse <= meter {
		t.Fatalf("pulse %v should exceed meter %v for speech-like peaks", pulse, meter)
	}
}

func TestBreathForVoiceDimsWhenSpeaking(t *testing.T) {
	if idle := BreathForVoice(1, 0, false); idle.OpacityMult != 1 {
		t.Fatalf("idle must not dim, got %v", idle.OpacityMult)
	}
	dipped := false
	for _, now := range []float64{0, 0.3, 1, 2.7, 5.1} {
		b := BreathForVoice(now, 0.6, false)
		// mult dips: never above 1, at most 0.30·wave below it (wave ≤ voice).
		if b.OpacityMult > 1 || b.OpacityMult < 1-0.30*0.6 {
			t.Fatalf("now=%v: opacity mult %v outside [1-0.30·voice, 1]", now, b.OpacityMult)
		}
		if b.OpacityMult < 1 {
			dipped = true
		}
	}
	if !dipped {
		t.Fatalf("expected a visible opacity dip while speaking")
	}
}

func TestVoicePulseAsymmetry(t *testing.T) {
	v := VoicePulse{}
	for i := 0; i < 60; i++ {
		v.Push(1)
		v.Advance(1 / 60.0)
	}
	if v.Level() <= 0.95 {
		t.Fatalf("expected attack to saturate, got %v", v.Level())
	}
	// tau 0.11 -> ~0.40 after 6 frames; a swapped-tau mutant gives ~0.03.
	for i := 0; i < 6; i++ {
		v.Advance(1 / 60.0)
	}
	if v.Level() <= 0.2 {
		t.Fatalf("expected slow decay, got %v", v.Level())
	}

	// tau 0.028 -> ~0.83 after 3 frames.
	fast := VoicePulse{}
	for i := 0; i < 3; i++ {
		fast.Push(1)
		fast.Advance(1 / 60.0)
	}
	if fast.Level() <= 0.6 {
		t.Fatalf("expected fast attack, got %v", fast.Level())
	}
}

func TestBreathForVoiceReduceMotionIsTimeInvariant(t *testing.T) {
	a := BreathForVoice(1, 0.6, true)
	b := BreathForVoice(1000, 0.6, true)
	if a != b {
		t.Fatalf("reduce motion must freeze the ripple: %v vs %v", a, b)
	}
	if a.OpacityMult == BreathForVoice(0, 0.6, false).OpacityMult {
		t.Fatalf("expected motion-driven ripple, got %v", a.OpacityMult)
	}
	if got := BreathForVoice(5, 0.02, false); got != (VoiceBreath{1, 1, 1}) {
		t.Fatalf("expected idle passthrough at or below 0.02, got %v", got)
	}
}

func TestPeakToPulseLevelMapping(t *testing.T) {
	cases := []struct {
		peak int
		want float32
	}{{0, 0}, {-100, 0}, {800, 0.5}, {3200, 1}, {32767, 1}}
	for _, c := range cases {
		if got := PeakToPulseLevel(c.peak); math.Abs(float64(got-c.want)) > 1e-6 {
			t.Fatalf("PeakToPulseLevel(%d) = %v, want %v", c.peak, got, c.want)
		}
	}

	v := VoicePulse{}
	v.Push(2)
	v.Push(-1)
	v.Advance(1 / 60.0)
	if l := v.Level(); math.IsNaN(l) || l < 0 || l > 1 {
		t.Fatalf("level out of range after clamped pushes: %v", l)
	}
}

func TestGradientAngleIntegratesSmoothly(t *testing.T) {
	phase := 0.0
	a := stepGradient(&phase, 1, false)
	b := stepGradient(&phase, 1, false)
	if math.Abs(float64(b-a)-gradientDegPerSec) > 0.01 {
		t.Fatalf("angle delta %v want ~%v", b-a, gradientDegPerSec)
	}
	frozen := stepGradient(&phase, 1, true)
	if frozen != 90 {
		t.Fatalf("reduce motion should freeze angle")
	}
}

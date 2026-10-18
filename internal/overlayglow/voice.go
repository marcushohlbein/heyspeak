package overlayglow

import (
	"math"
	"time"
)

type queuedPulse struct {
	level float32
	wait  float64
}

type VoicePulse struct {
	smooth float64
	target float64
	queue  []queuedPulse
}

func (v *VoicePulse) Queue(level float32, after time.Duration) {
	v.queue = append(v.queue, queuedPulse{clampLevel(level), after.Seconds()})
}

func (v *VoicePulse) Push(level float32) {
	x := float64(clampLevel(level))
	if x > v.target {
		v.target = x
	}
}

func (v *VoicePulse) Advance(dt float64) {
	if dt > 0.1 {
		dt = 0.1
	}
	if dt > 0 {
		for i := range v.queue {
			v.queue[i].wait -= dt
		}
		for len(v.queue) > 0 && v.queue[0].wait <= 0 {
			v.Push(v.queue[0].level)
			v.queue = v.queue[1:]
		}
	}
	if dt <= 0 {
		return
	}
	target := v.target
	v.target = 0

	tau := 0.028
	if target < v.smooth {
		tau = 0.11
	}
	k := 1 - math.Exp(-dt/tau)
	v.smooth += (target - v.smooth) * k
	if v.smooth < 1e-4 {
		v.smooth = 0
	}
}

func (v *VoicePulse) Level() float64 {
	return v.smooth
}

func (v *VoicePulse) Reset() {
	v.smooth = 0
	v.target = 0
	v.queue = nil
}

// PeakToPulseLevel maps PCM16 peaks to 0..1 for the edge glow (more sensitive than the settings meter).
func PeakToPulseLevel(peak int) float32 {
	if peak <= 0 {
		return 0
	}
	x := float64(peak) / 1600
	if x > 1 {
		x = 1
	}
	return float32(math.Sqrt(x))
}

type VoiceBreath struct {
	OpacityMult float64
	WidthMult   float32
	BlurMult    float32
}

func BreathForVoice(now float64, voice float64, reduceMotion bool) VoiceBreath {
	if voice <= 0.02 {
		if reduceMotion {
			return VoiceBreath{1, 1, 1}
		}
		return VoiceBreath{
			OpacityMult: 1,
			WidthMult:   float32(1 + 0.025*math.Sin(now*1.2566)),
			BlurMult:    1,
		}
	}
	beat := 1.0
	if !reduceMotion {
		beat = 0.55 + 0.45*math.Sin(now*15.8)
	}
	wave := voice * (0.65 + 0.35*beat)
	return VoiceBreath{
		OpacityMult: 1 - wave*0.30*(1-beat),
		WidthMult:   float32(1 + wave*0.10),
		BlurMult:    float32(1 + wave*0.18),
	}
}

func dampBreath(b VoiceBreath, opacity float64) VoiceBreath {
	if opacity >= 0.94 {
		return b
	}
	t := opacity / 0.94
	ft := float32(t)
	return VoiceBreath{
		OpacityMult: 1 + (b.OpacityMult-1)*t,
		WidthMult:   1 + (b.WidthMult-1)*ft,
		BlurMult:    1 + (b.BlurMult-1)*ft,
	}
}

func clampLevel(level float32) float32 {
	if level < 0 {
		return 0
	}
	if level > 1 {
		return 1
	}
	return level
}

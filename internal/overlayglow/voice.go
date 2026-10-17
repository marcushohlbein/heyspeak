package overlayglow

import "math"

// VoicePulse smooths microphone peaks into a 0..1 envelope for Siri-like edge pulsing.
type VoicePulse struct {
	smooth float64
	target float64
}

// Push records the loudest sample seen for the current frame (0..1).
func (v *VoicePulse) Push(level float32) {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	x := float64(level)
	if x > v.target {
		v.target = x
	}
}

// Advance moves the envelope toward the frame peak; call once per rendered frame.
func (v *VoicePulse) Advance(dt float64) {
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

// Level returns the smoothed voice activity in 0..1.
func (v *VoicePulse) Level() float64 {
	return v.smooth
}

// Reset clears voice reactivity (session ended).
func (v *VoicePulse) Reset() {
	v.smooth = 0
	v.target = 0
}

// PeakToPulseLevel maps capture peaks to a 0..1 pulse. Lower than the settings
// meter (peak/8000) so normal speech moves the edge.
func PeakToPulseLevel(peak int) float32 {
	if peak <= 0 {
		return 0
	}
	x := float64(peak) / 3200
	if x > 1 {
		x = 1
	}
	return float32(math.Sqrt(x))
}

// VoiceBreath scales the glow while the user speaks (opacity, stroke, blur).
type VoiceBreath struct {
	OpacityMult float64
	WidthMult   float32
	BlurMult    float32
}

// VoiceBreath returns multipliers for a Siri-like inhale/exhale tied to voice level.
func BreathForVoice(now float64, voice float64, reduceMotion bool) VoiceBreath {
	if voice <= 0.02 {
		return VoiceBreath{1, 1, 1}
	}
	beat := 1.0
	if !reduceMotion {
		// ~2.5 Hz ripple on top of the level envelope.
		beat = 0.55 + 0.45*math.Sin(now*15.8)
	}
	wave := voice * (0.65 + 0.35*beat)
	return VoiceBreath{
		// Dim-breathe: brighten-only clamps at full overlay opacity, so the
		// pulse was invisible in the steady recording state. Dip instead.
		OpacityMult: 1 - wave*0.30*(1-beat),
		WidthMult:   float32(1 + wave*0.42),
		BlurMult:    float32(1 + wave*0.65),
	}
}

// dampBreath quiets the pulse while the overlay is still fading in or out.
func dampBreath(b VoiceBreath, opacity float64) VoiceBreath {
	if opacity >= 0.94 {
		return b
	}
	t := opacity / 0.94
	if t < 0 {
		t = 0
	}
	ft := float32(t)
	return VoiceBreath{
		OpacityMult: 1 + (b.OpacityMult-1)*t,
		WidthMult:   1 + (b.WidthMult-1)*ft,
		BlurMult:    1 + (b.BlurMult-1)*ft,
	}
}

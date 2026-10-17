package whisper

import "github.com/speek/speek/internal/audio"

func pcm16ToFloat32(samples []int16) []float32 {
	if len(samples) == 0 {
		return nil
	}
	out := make([]float32, len(samples))
	const scale = 1.0 / 32768.0
	for i, s := range samples {
		out[i] = float32(float64(s) * scale)
	}
	return out
}

const minSpeechSamples = int(0.15 * float64(audio.SampleRate))

func peakAbsPCM(samples []int16) int {
	var peak int
	for _, s := range samples {
		v := int(s)
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	return peak
}

// boostQuietPCM raises a quiet-but-present signal into whisper's comfortable range.
// Near-silence is left alone so it is not amplified into a hallucination.
func boostQuietPCM(samples []int16, peak int) []int16 {
	if peak < 120 || peak >= 6000 || len(samples) == 0 {
		return samples
	}
	gain := 6000.0 / float64(peak)
	if gain > 6 {
		gain = 6
	}
	out := make([]int16, len(samples))
	for i, s := range samples {
		v := float64(s) * gain
		if v > 32767 {
			v = 32767
		}
		if v < -32768 {
			v = -32768
		}
		out[i] = int16(v)
	}
	return out
}

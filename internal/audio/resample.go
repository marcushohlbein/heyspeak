package audio

// ResampleLinear converts PCM16 samples to targetRate using linear interpolation.
func ResampleLinear(samples []int16, sourceRate, targetRate int) []int16 {
	if sourceRate <= 0 || targetRate <= 0 || len(samples) == 0 {
		return samples
	}
	if sourceRate == targetRate {
		return samples
	}
	outLen := len(samples) * targetRate / sourceRate
	if outLen == 0 {
		return nil
	}
	out := make([]int16, outLen)
	ratio := float64(sourceRate) / float64(targetRate)
	for i := 0; i < outLen; i++ {
		src := float64(i) * ratio
		j := int(src)
		if j >= len(samples)-1 {
			out[i] = samples[len(samples)-1]
			continue
		}
		frac := src - float64(j)
		a := float64(samples[j])
		b := float64(samples[j+1])
		out[i] = int16(a + (b-a)*frac)
	}
	return out
}

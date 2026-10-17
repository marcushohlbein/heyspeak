package audio

// appendBounded appends PCM16 frames up to maxSamples. Returns true when the cap is reached.
func appendBounded(dst []int16, frame []int16, maxSamples int) ([]int16, bool) {
	if maxSamples <= 0 {
		return append(dst, frame...), false
	}
	room := maxSamples - len(dst)
	if room <= 0 {
		return dst, true
	}
	if len(frame) > room {
		dst = append(dst, frame[:room]...)
		return dst, true
	}
	dst = append(dst, frame...)
	return dst, len(dst) >= maxSamples
}

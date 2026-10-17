package audio

import "context"

// ProbeResult describes a short microphone check without keeping the samples.
type ProbeResult struct {
	Samples    int
	DurationMS float64
	Peak       int
}

// Probe records until ctx is cancelled and reports level, not the audio itself.
func Probe(ctx context.Context, deviceID string) (ProbeResult, error) {
	samples, err := recordSession(ctx, SessionConfig{
		DeviceID:   deviceID,
		SampleRate: uint32(SampleRate),
		MaxSeconds: 5,
	})
	if err != nil {
		return ProbeResult{}, err
	}
	return ProbeResult{
		Samples:    len(samples),
		DurationMS: float64(len(samples)) / float64(SampleRate) * 1000,
		Peak:       peakAbs(samples),
	}, nil
}

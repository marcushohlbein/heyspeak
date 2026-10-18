package overlay

import (
	"math"
	"sync"
	"testing"
)

func TestOnMicPeakKeepsLoudest(t *testing.T) {
	m := Manager{}
	m.OnMicPeak(16000)
	m.OnMicPeak(4000)
	m.OnMicPeak(800)
	if got := math.Float32frombits(m.micPeak.Load()); got != 1 {
		t.Fatalf("expected clamped peak 1, got %v", got)
	}

	m.micPeak.Store(0)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.OnMicPeak(500 + i*400)
		}(i)
	}
	wg.Wait()
	if got := math.Float32frombits(m.micPeak.Load()); got != 1 {
		t.Fatalf("expected loudest peak kept, got %v", got)
	}
}

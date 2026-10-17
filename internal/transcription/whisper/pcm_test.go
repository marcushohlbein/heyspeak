package whisper

import "testing"

func peakOf(samples []int16) int { return peakAbsPCM(samples) }

func TestBoostQuietPCM(t *testing.T) {
	// Below the floor: unchanged.
	quiet := []int16{50, -100, 80}
	if got := boostQuietPCM(quiet, peakOf(quiet)); &got[0] != &quiet[0] {
		t.Fatalf("peak 100 should be unchanged, got %v", got)
	}

	// Already loud enough: unchanged.
	loud := []int16{7000, -6500}
	if got := boostQuietPCM(loud, peakOf(loud)); &got[0] != &loud[0] {
		t.Fatalf("peak 7000 should be unchanged, got %v", got)
	}

	// Quiet-but-present: raised toward 6000.
	quietPresent := []int16{3000, -1500, 750}
	got := boostQuietPCM(quietPresent, peakOf(quietPresent))
	wantPeak := 6000
	if p := peakOf(got); abs(p-wantPeak) > 50 {
		t.Fatalf("peak 3000: output peak %d want ~%d (samples %v)", p, wantPeak, got)
	}

	// Clamping: gain 2x on 30000 stays in range with no sign flip.
	big := []int16{30000, -30000}
	boosted := boostQuietPCM(big, 3000)
	if boosted[0] != 32767 || boosted[1] != -32768 {
		t.Fatalf("clamped samples %v want [32767 -32768]", boosted)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

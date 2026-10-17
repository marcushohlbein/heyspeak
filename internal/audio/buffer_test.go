package audio

import "testing"

func TestAppendBoundedCaps(t *testing.T) {
	dst := []int16{1, 2}
	dst, full := appendBounded(dst, []int16{3, 4, 5}, 4)
	if !full || len(dst) != 4 {
		t.Fatalf("got len=%d full=%v", len(dst), full)
	}
}

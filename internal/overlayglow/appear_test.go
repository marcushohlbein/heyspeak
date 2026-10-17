package overlayglow

import "testing"

func TestAppearBloomSoftensLowOpacity(t *testing.T) {
	if appearBloom(0.5) >= 0.5 {
		t.Fatalf("bloom should be softer than linear at 0.5")
	}
	if appearBloom(1) != 1 {
		t.Fatalf("full opacity should bloom to 1")
	}
}

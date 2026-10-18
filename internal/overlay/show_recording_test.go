package overlay

import "testing"

func TestShowRecordingOpacityContinuesFromFade(t *testing.T) {
	m := &Manager{opacity: 0.42, hiding: true}
	m.mu.Lock()
	m.hiding = false
	if m.opacity <= 0.001 {
		m.opacity = 0
		m.opacityFrom = 0
	} else {
		m.opacityFrom = m.opacity
	}
	m.opacityTarget = 1
	m.mu.Unlock()

	if m.opacityFrom != 0.42 {
		t.Fatalf("opacityFrom %v want 0.42", m.opacityFrom)
	}
	if m.opacity != 0.42 {
		t.Fatalf("opacity should not snap to zero")
	}
}

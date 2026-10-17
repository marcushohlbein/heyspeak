package paste

import "testing"

func TestResolveFocusForPaste(t *testing.T) {
	const target = 100
	const own = 200

	tests := []struct {
		name     string
		front    int
		target   int
		activate bool
		ok       bool
	}{
		{name: "target focused", front: target, target: target, activate: false, ok: true},
		{name: "speek took focus", front: own, target: target, activate: true, ok: true},
		{name: "other app", front: 300, target: target, activate: false, ok: false},
		{name: "no target captured", front: target, target: 0, activate: false, ok: true},
		{name: "capture failed, other app focused", front: target, target: -1, activate: false, ok: false},
		{name: "capture failed, third app focused", front: 300, target: -1, activate: false, ok: false},
		{name: "capture failed, speek frontmost", front: own, target: -1, activate: false, ok: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			activate, ok := resolveFocusForPaste(focusState{
				frontmost: tc.front,
				own:       own,
				target:    tc.target,
			})
			if activate != tc.activate || ok != tc.ok {
				t.Fatalf("activate=%v ok=%v", activate, ok)
			}
		})
	}
}

package audio

import "testing"

func TestMicStatusLabel(t *testing.T) {
	if !MicAuthorized.Granted() || MicAuthorized.Label() != "Granted" {
		t.Fatal(MicAuthorized)
	}
	if MicDenied.Granted() || MicDenied.Label() != "Denied" {
		t.Fatal(MicDenied)
	}
	if MicNotDetermined.Label() != "Not asked yet" {
		t.Fatal(MicNotDetermined.Label())
	}
}

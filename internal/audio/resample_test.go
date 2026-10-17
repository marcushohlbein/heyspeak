package audio

import (
	"reflect"
	"testing"
)

func TestResampleLinearHalvesRate(t *testing.T) {
	in := []int16{0, 1000, 2000, 3000}
	out := ResampleLinear(in, 16000, 8000)
	if want := []int16{0, 2000}; !reflect.DeepEqual(out, want) {
		t.Fatalf("16k->8k: got %v want %v", out, want)
	}
}

func TestResampleLinearDoublesRate(t *testing.T) {
	in := []int16{0, 1000, 2000, 3000}
	out := ResampleLinear(in, 8000, 16000)
	want := []int16{0, 500, 1000, 1500, 2000, 2500, 3000, 3000}
	if len(out) != 2*len(in) {
		t.Fatalf("len %d want %d", len(out), 2*len(in))
	}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("8k->16k: got %v want %v", out, want)
	}
	// Tail replicates the last input sample rather than interpolating past it.
	if out[len(out)-1] != in[len(in)-1] || out[len(out)-2] != in[len(in)-1] {
		t.Fatalf("tail %v should replicate last sample %d", out[len(out)-2:], in[len(in)-1])
	}
}

func TestResampleLinearPassthrough(t *testing.T) {
	in := []int16{100, -200, 300}
	if got := ResampleLinear(in, 16000, 16000); !reflect.DeepEqual(got, in) {
		t.Fatalf("equal rate: got %v want %v", got, in)
	}
	if got := ResampleLinear(in, 0, 16000); !reflect.DeepEqual(got, in) {
		t.Fatalf("zero source rate: got %v want %v", got, in)
	}
	if got := ResampleLinear(in, 16000, 0); !reflect.DeepEqual(got, in) {
		t.Fatalf("zero target rate: got %v want %v", got, in)
	}
}

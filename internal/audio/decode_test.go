package audio

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/gen2brain/malgo"
)

func TestDecodeStereoS16AveragesChannels(t *testing.T) {
	buf := make([]byte, 8)
	put := func(off int, v int16) {
		binary.LittleEndian.PutUint16(buf[off:], uint16(v))
	}
	put(0, 1000)
	put(2, 3000)
	put(4, -1000)
	put(6, -3000)
	out := decodeCapture(buf, 2, 2, malgo.FormatS16)
	if len(out) != 2 || out[0] != 2000 || out[1] != -2000 {
		t.Fatalf("%v", out)
	}
}

func TestDecodeF32(t *testing.T) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(0.5))
	out := decodeCapture(buf, 1, 1, malgo.FormatF32)
	if len(out) != 1 || out[0] < 16000 || out[0] > 17000 {
		t.Fatalf("%v", out)
	}
}

func TestChooseCaptureRatePrefers48k(t *testing.T) {
	rate := chooseCaptureRate([]malgo.DataFormat{
		{SampleRate: 16000},
		{SampleRate: 48000},
	})
	if rate != 48000 {
		t.Fatalf("%d", rate)
	}
}

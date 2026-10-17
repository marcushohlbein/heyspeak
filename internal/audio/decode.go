package audio

import (
	"encoding/binary"
	"math"

	"github.com/gen2brain/malgo"
)

// decodeCapture converts one callback buffer into mono PCM16.
// frameCount and the opened device's channel count and format must be used.
// The values requested before the device starts can differ from those.
func decodeCapture(input []byte, frameCount int, channels int, format malgo.FormatType) []int16 {
	if frameCount <= 0 || len(input) == 0 {
		return nil
	}
	if channels <= 0 {
		channels = 1
	}
	out := make([]int16, frameCount)
	switch format {
	case malgo.FormatF32:
		for i := 0; i < frameCount; i++ {
			var sum float64
			n := 0
			for ch := 0; ch < channels; ch++ {
				off := (i*channels + ch) * 4
				if off+4 > len(input) {
					break
				}
				sum += float64(math.Float32frombits(binary.LittleEndian.Uint32(input[off : off+4])))
				n++
			}
			if n == 0 {
				break
			}
			out[i] = floatToPCM16(sum / float64(n))
		}
	default:
		// F32 is requested when offered (see chooseCaptureFormat); other formats are read as 16-bit little-endian pairs.
		for i := 0; i < frameCount; i++ {
			var sum int32
			n := 0
			for ch := 0; ch < channels; ch++ {
				off := (i*channels + ch) * 2
				if off+2 > len(input) {
					break
				}
				sum += int32(int16(binary.LittleEndian.Uint16(input[off : off+2])))
				n++
			}
			if n == 0 {
				break
			}
			out[i] = int16(sum / int32(n))
		}
	}
	return out
}

func floatToPCM16(v float64) int16 {
	if v > 1 {
		v = 1
	}
	if v < -1 {
		v = -1
	}
	return int16(v * 32767)
}

func chooseCaptureFormat(formats []malgo.DataFormat) malgo.FormatType {
	for _, f := range formats {
		if f.Format == malgo.FormatF32 {
			return malgo.FormatF32
		}
	}
	for _, f := range formats {
		if f.Format == malgo.FormatS16 {
			return malgo.FormatS16
		}
	}
	return malgo.FormatF32
}

func chooseCaptureRate(formats []malgo.DataFormat) uint32 {
	have := map[uint32]bool{}
	for _, f := range formats {
		if f.SampleRate > 0 {
			have[f.SampleRate] = true
		}
	}
	for _, rate := range []uint32{48000, 44100, 32000, 16000} {
		if have[rate] {
			return rate
		}
	}
	for rate := range have {
		return rate
	}
	return 48000
}

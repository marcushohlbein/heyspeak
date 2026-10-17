package overlayglow

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// Siri / Apple Intelligence edge palette. The warm arc (pink → coral →
// amber) is mirrored 180° so both sides of the screen run warm, with a
// cool gap (violet, blue) between the arcs; motion comes from rotating
// the conic angle in view.go.
var siriRing = []ui.Color{
	ui.Hex("#F5B9EA"),
	ui.Hex("#FF6778"),
	ui.Hex("#FFBA71"),
	ui.Hex("#BC82F3"),
	ui.Hex("#F5B9EA"),
	ui.Hex("#FF6778"),
	ui.Hex("#FFBA71"),
	ui.Hex("#0894FF"),
}

// 8 palette colors × 3 samples − 1 seam-shared + 2 seam stops = 25 ≤ 32;
// mygo's painter caps gradients at scene.MaxStops = 32 and silently drops
// the rest (ui/paint.go stops()), which would cut the closing seam stop.
const siriBlendSteps = 3 // Oklab samples between each palette color

type stop struct {
	l, a, b, at float64
}

// Stops is the fixed ring. Rotation is a separate phase so stop positions never jump.
type Stops struct{}

func NewStops() Stops { return Stops{} }

var ringStops = siriGradientStops()

func (s *Stops) Gradient(float64) []ui.GradientStop { return ringStops }

func siriGradientStops() []ui.GradientStop {
	st := siriRingStops(siriBlendSteps)
	if len(st) == 0 {
		return nil
	}
	first, last := st[0], st[len(st)-1]
	k := (1 - last.at) / max(first.at+1-last.at, 1e-6)
	seam := srgb(
		last.l+(first.l-last.l)*k,
		last.a+(first.a-last.a)*k,
		last.b+(first.b-last.b)*k,
	)
	out := make([]ui.GradientStop, 0, len(st)+2)
	out = append(out, ui.GradientStop{Color: seam, At: 0})
	for _, x := range st {
		out = append(out, ui.GradientStop{Color: srgb(x.l, x.a, x.b), At: float32(x.at)})
	}
	return append(out, ui.GradientStop{Color: seam, At: 1})
}

func siriRingStops(stepsPerSegment int) []stop {
	n := len(siriRing)
	out := make([]stop, 0, n*stepsPerSegment)
	for i := 0; i < n; i++ {
		l0, a0, b0 := oklab(siriRing[i])
		l1, a1, b1 := oklab(siriRing[(i+1)%n])
		for j := 0; j < stepsPerSegment; j++ {
			t := float64(j) / float64(stepsPerSegment)
			out = append(out, stop{
				l:  l0 + (l1-l0)*t,
				a:  a0 + (a1-a0)*t,
				b:  b0 + (b1-b0)*t,
				at: (float64(i) + t) / float64(n),
			})
		}
	}
	return out
}

func srgb(l, a, b float64) ui.Color {
	lc := l + 0.3963377774*a + 0.2158037573*b
	mc := l - 0.1055613458*a - 0.0638541728*b
	sc := l - 0.0894841775*a - 1.2914855480*b
	lc, mc, sc = lc*lc*lc, mc*mc*mc, sc*sc*sc
	enc := func(v float64) uint8 {
		v = min(max(v, 0), 1)
		if v <= 0.0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		return uint8(math.Round(v * 255))
	}
	return ui.RGB(enc(4.0767416621*lc-3.3077115913*mc+0.2309699292*sc),
		enc(-1.2684380046*lc+2.6097574011*mc-0.3413193965*sc),
		enc(-0.0041960863*lc-0.7034186147*mc+1.7076147010*sc))
}

func oklab(c ui.Color) (l, a, b float64) {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	lc := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*bl)
	mc := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*bl)
	sc := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*bl)
	return 0.2104542553*lc + 0.7936177850*mc - 0.0040720468*sc,
		1.9779984951*lc - 2.4285922050*mc + 0.4505937099*sc,
		0.0259040371*lc + 0.7827717662*mc - 0.8086757660*sc
}

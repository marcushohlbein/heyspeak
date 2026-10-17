package overlayglow

import (
	"runtime"
	"time"

	"github.com/egoist/mygo/ui"
)

var glowLayers = []struct{ width, blur float32 }{{6, 0}, {9, 4}, {11, 12}, {15, 15}}

// State drives the edge overlay view.
type State struct {
	Visible bool
	Opacity float64 // 0..1 animated envelope
	Visual  Visual
	Stops   Stops
	GPU     bool
	Start   time.Time
}

// Draw renders a full-window pass-through edge glow (adapted from MyGo chat-glow).
func Draw(s *State, c *ui.Context) {
	c.Root().Background(ui.Transparent)

	if !s.Visible || s.Opacity <= 0.001 {
		return
	}

	now := time.Since(s.Start).Seconds()
	corner := float32(10)
	if runtime.GOOS == "darwin" {
		corner = 16
	}
	g := ui.ConicGradient{Stops: s.Stops.Gradient(now), Angle: 90, Oklab: true}

	ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().
		Draw(func(p *ui.Painter, r ui.Rect) { s.GPU = p.Accelerated() }).
		Children(func() {
			if !s.GPU {
				return
			}
			strength := ease(s.Opacity)
			if s.Visual == VisualTranscribing {
				strength *= 0.45
			}
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().
				Opacity(float32(strength)).Children(func() {
				for _, l := range glowLayers {
					ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().Radius(corner).
						Border(l.width, ui.Color{}).BorderConicGradient(g).Blur(0.95 * l.blur)
				}
			})
		})
	c.AnimationFrame()
}

package overlayglow

import (
	"runtime"
	"time"

	"github.com/egoist/mygo/ui"
)

var glowLayers = []struct{ width, blur float32 }{{5, 0}, {7, 3}, {8, 9}, {11, 11}}

const gradientDegPerSec = 20.0

// State drives the edge overlay view.
type State struct {
	Visible       bool
	Opacity       float64 // 0..1 animated envelope
	TranscribeDim float64
	Stops         Stops
	Voice         VoicePulse
	GPU           bool
	Start         time.Time
	LastFrame     time.Time
	GradientPhase float64
}

// Draw renders a full-window pass-through edge glow (adapted from MyGo chat-glow).
func Draw(s *State, c *ui.Context) {
	c.Root().Background(ui.Transparent)

	if !s.Visible || s.Opacity <= 0.001 {
		return
	}

	reduceMotion := c.Preferences().ReduceMotion

	frameNow := time.Now()
	dt := 0.0
	if !s.LastFrame.IsZero() {
		dt = frameNow.Sub(s.LastFrame).Seconds()
		if dt > 0.1 {
			dt = 0.1
		}
	}
	s.LastFrame = frameNow

	now := time.Since(s.Start).Seconds()
	s.Voice.Advance(dt)
	voice := s.Voice.Level()

	corner := float32(10)
	if runtime.GOOS == "darwin" {
		corner = 16
	}
	angle := stepGradient(&s.GradientPhase, dt, reduceMotion)
	g := ui.ConicGradient{Stops: s.Stops.Gradient(now), Angle: angle, Oklab: true}

	ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().
		Draw(func(p *ui.Painter, r ui.Rect) { s.GPU = p.Accelerated() }).
		Children(func() {
			if !s.GPU {
				return
			}
			strength := s.Opacity * s.TranscribeDim
			bloom := appearBloom(s.Opacity)
			breath := dampBreath(BreathForVoice(now, voice, reduceMotion), s.Opacity)
			strength *= breath.OpacityMult
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().
				Opacity(float32(strength)).Children(func() {
				for _, l := range glowLayers {
					ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().Radius(corner).
						Border(l.width*breath.WidthMult*bloom, ui.Color{}).BorderConicGradient(g).
						Blur(0.95 * l.blur * breath.BlurMult * bloom)
				}
			})
		})
	c.AnimationFrame()
}

func stepGradient(phase *float64, dt float64, reduceMotion bool) float32 {
	if reduceMotion {
		return float32(90 + *phase)
	}
	if dt > 0 {
		*phase += dt * gradientDegPerSec
	}
	return float32(90 + *phase)
}

func appearBloom(opacity float64) float32 {
	if opacity <= 0 {
		return 0
	}
	if opacity >= 1 {
		return 1
	}
	return float32(opacity * opacity)
}

// Spike 1B: display-sized edge glow overlay (non-activating, click-through).
//
//	go run ./cmd/spike-overlay
//
// Toggle overlay: Alt+Shift+G. Quit: tray menu.
package main

import (
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/overlayglow"
)

type overlay struct {
	win        *mygo.Window
	visible    bool
	state      overlayglow.State
	lastBounds mygo.Rectangle
}

func main() {
	o := &overlay{state: overlayglow.State{Stops: overlayglow.NewStops(), Start: time.Now(), Opacity: 1}}
	mygo.App.SetName("speek overlay spike")
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.OnWindowAllClosed(func() {})

	mygo.App.WhenReady(func() {
		mygo.Screen.OnDisplaysChanged(func() {
			if o.visible {
				o.refreshWindow()
			}
		})

		if err := mygo.GlobalShortcut.Register("Alt+Shift+G", o.toggle); err != nil {
			log.Printf("toggle shortcut: %v", err)
		}

		tray, err := mygo.NewTray(mygo.TrayOptions{ToolTip: "speek overlay spike"})
		if err == nil {
			tray.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
				{Label: "Toggle glow (Alt+Shift+G)", Click: func(*mygo.MenuItem, *mygo.Window) { o.toggle() }},
				{Role: mygo.RoleQuit},
			}))
		}
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func (o *overlay) view(c *ui.Context) {
	o.state.Visible = o.visible
	overlayglow.Draw(&o.state, c)
}

func (o *overlay) toggle() {
	if o.visible {
		o.hide()
	} else {
		o.show()
	}
}

func (o *overlay) show() {
	o.visible = true
	o.state.Start = time.Now()
	o.refreshWindow()
	o.present()
}

func (o *overlay) hide() {
	o.visible = false
	if o.win != nil {
		o.win.Hide()
	}
}

func (o *overlay) refreshWindow() {
	pt := mygo.Screen.CursorScreenPoint()
	d := mygo.Screen.DisplayNearestPoint(pt)
	target := overlayglow.DisplayRect(d)

	if o.win != nil && o.lastBounds == target {
		return
	}

	if o.win != nil {
		o.win.Close()
		o.win = nil
	}

	o.lastBounds = target
	// X and Y must not both be 0: MyGo treats that as "center on the visible frame".
	// PlaceOverDisplay then moves the panel to the full screen, above the menu bar and Dock.
	o.win = mygo.NewWindow(mygo.WindowOptions{
		Title:         "speek overlay",
		X:             1,
		Y:             1,
		Width:         target.Width,
		Height:        target.Height,
		Hidden:        true,
		Panel:         true,
		Frameless:     true,
		Transparent:   true,
		DisableShadow: true,
		AlwaysOnTop:   true,
		SkipTaskbar:   true,
		DisableResize: true,
		Content:       ui.View(o.view),
	})
	if o.visible {
		o.present()
	}
}

func (o *overlay) present() {
	if o.win == nil {
		return
	}
	o.win.ShowInactive()
	o.win.SetIgnoreMouseEvents(true)
	overlayglow.PlaceOverDisplay(o.win.NativeHandle())
	o.win.Update(func() {})
}

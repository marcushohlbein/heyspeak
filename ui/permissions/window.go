package permissions

import (
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/paste"
)

const (
	width  = 460
	height = 420

	micSettingsURL    = "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone"
	accessSettingsURL = "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
)

// Window lists the system permissions speek needs and whether each is granted.
type Window struct {
	win      *mygo.Window
	mic      audio.MicStatus
	accessOK bool
}

// New builds a hidden permissions window.
func New() *Window {
	w := &Window{}
	w.refresh()
	w.win = mygo.NewWindow(mygo.WindowOptions{
		Title:         "speek permissions",
		Width:         width,
		Height:        height,
		Hidden:        true,
		DisableResize: true,
		Content:       ui.View(w.view),
	})
	w.win.OnClose(func(e *mygo.CloseEvent) {
		e.PreventDefault()
		w.Hide()
	})
	return w
}

// NeedsAttention reports whether a permission is still missing on this platform.
func NeedsAttention() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	return !audio.MicrophoneStatus().Granted() || !paste.AccessibilityTrusted()
}

// Show brings the permissions window to the front.
func (w *Window) Show() {
	w.refresh()
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyRegular)
	mygo.App.Show()
	mygo.App.Focus()
	d := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
	w.win.SetPosition(d.WorkArea.X+80, d.WorkArea.Y+80)
	w.win.Show()
	w.win.Focus()
	w.win.Update(func() {})
}

// Hide dismisses the window and returns speek to the menu bar.
func (w *Window) Hide() {
	if w.win != nil {
		w.win.Hide()
	}
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
}

func (w *Window) refresh() {
	w.mic = audio.MicrophoneStatus()
	w.accessOK = paste.AccessibilityTrusted()
}

func (w *Window) view(c *ui.Context) {
	t := c.Theme()
	c.Root().Background(t.Background)
	c.OnShortcut(ui.Cmd, ui.KeyW, w.Hide)

	ui.Column(c).Fill().Padding(24).Gap(14).Children(func() {
		ui.Text(c, "Permissions").FontSize(22)
		ui.Text(c, "speek needs these before dictation can hear you and paste the result. In System Settings, allow the app named speek. T3, Cursor, and Terminal do not appear in Microphone unless that app itself asked for access.").
			FontSize(13).TextColor(t.TextMuted)

		w.row(c, "Microphone", "Records what you dictate.", w.mic.Granted(), w.mic.Label())
		if !w.mic.Granted() {
			ui.Row(c).Gap(8).Children(func() {
				if w.mic == audio.MicNotDetermined && ui.Button(c, "Allow microphone").Clicked() {
					w.mic = audio.RequestMicrophone()
					w.win.Update(func() {})
				}
				if ui.Button(c, "Open Microphone settings").Clicked() {
					_ = mygo.Shell.OpenExternal(micSettingsURL)
				}
			})
		}

		accessLabel := "Not granted"
		if w.accessOK {
			accessLabel = "Granted"
		}
		w.row(c, "Accessibility", "Reads the push-to-talk shortcut and pastes into the app you were using.", w.accessOK, accessLabel)
		if !w.accessOK {
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Allow accessibility").Clicked() {
					w.accessOK = paste.PromptAccessibility()
					w.win.Update(func() {})
				}
				if ui.Button(c, "Open Accessibility settings").Clicked() {
					_ = mygo.Shell.OpenExternal(accessSettingsURL)
				}
			})
		}

		ui.Row(c).Gap(8).Children(func() {
			if ui.Button(c, "Refresh").Clicked() {
				w.refresh()
				w.win.Update(func() {})
			}
			label := "Continue"
			if w.mic.Granted() && w.accessOK {
				label = "Done"
			}
			if ui.Button(c, label).Clicked() {
				w.Hide()
			}
		})
	})
}

func (w *Window) row(c *ui.Context, title, detail string, ok bool, state string) {
	t := c.Theme()
	color := t.Warning
	if ok {
		color = t.Success
	}
	ui.Column(c).Gap(2).Children(func() {
		ui.Row(c).Gap(8).Children(func() {
			ui.Text(c, title).FontSize(15)
			ui.Text(c, state).FontSize(13).TextColor(color)
		})
		ui.Text(c, detail).FontSize(12).TextColor(t.TextMuted)
	})
}

package settings

import (
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/paste"
)

const (
	micSettingsURL    = "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone"
	accessSettingsURL = "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
)

func (w *Window) refreshPermissions() {
	w.mic = audio.MicrophoneStatus()
	w.accessOK = paste.AccessibilityTrusted()
}

func (w *Window) permissionsPanel(c *ui.Context) {
	t := c.Theme()
	if runtime.GOOS != "darwin" {
		ui.Text(c, "On this platform, speek does not use the macOS permission prompts below.").
			FontSize(13).TextColor(t.TextMuted)
		return
	}

	ui.Text(c, "In System Settings, allow the app named speek. T3, Cursor, and Terminal do not appear in Microphone unless that app itself asked for access.").
		FontSize(13).TextColor(t.TextMuted)

	card(c, func() {
		w.permissionHeader(c, "Microphone", w.mic.Granted(), w.mic.Label())
		ui.Text(c, "Records what you dictate.").FontSize(12).TextColor(t.TextMuted)
		if !w.mic.Granted() {
			ui.Row(c).Gap(8).Children(func() {
				if w.mic == audio.MicNotDetermined && ui.Button(c, "Allow microphone").Clicked() {
					w.mic = audio.RequestMicrophone()
				}
				if ui.Button(c, "Open Microphone settings").Clicked() {
					_ = mygo.Shell.OpenExternal(micSettingsURL)
				}
			})
		}
	})

	card(c, func() {
		w.permissionHeader(c, "Accessibility", w.accessOK, accessState(w.accessOK))
		ui.Text(c, "Reads the push-to-talk shortcut and pastes into the app you were using.").
			FontSize(12).TextColor(t.TextMuted)
		if !w.accessOK {
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Allow accessibility").Clicked() {
					w.accessOK = paste.PromptAccessibility()
				}
				if ui.Button(c, "Open Accessibility settings").Clicked() {
					_ = mygo.Shell.OpenExternal(accessSettingsURL)
				}
			})
		}
	})

	if ui.Button(c, "Refresh status").AlignSelf(ui.Start).Clicked() {
		w.refreshPermissions()
	}
}

func accessState(ok bool) string {
	if ok {
		return "Granted"
	}
	return "Not granted"
}

// permissionHeader is a permission's title row: name and granted state.
func (w *Window) permissionHeader(c *ui.Context, title string, ok bool, state string) {
	t := c.Theme()
	color := t.Warning
	if ok {
		color = t.Success
	}
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontSize(14).Bold()
		ui.Spacer(c)
		ui.Text(c, state).FontSize(13).TextColor(color)
	})
}

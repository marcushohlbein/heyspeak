package settings

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/holdshortcut"
)

func (w *Window) shortcutPanel(c *ui.Context) {
	t := c.Theme()

	ui.Text(c, "Hold this key combination anywhere to dictate. It is saved as soon as you record it.").
		FontSize(12).TextColor(t.TextMuted)

	if w.recordingShortcut {
		ui.Column(c).Gap(8).Padding(14, 16).Radius(8).
			Border(2, t.Accent).Background(t.Surface).Children(func() {
			ui.Text(c, "Listening…").FontSize(14).Bold().TextColor(t.Accent)
			ui.Text(c, "Press the keys you want to hold while speaking. Press Esc to cancel.").
				FontSize(12).TextColor(t.TextMuted)
			capture := ui.Box(c).Height(1).Focusable().Bind(&w.shortcutCaptureH).AutoFocus()
			capture.HandleInput(func(ev ui.InputEvent) bool {
				if ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyEscape && ev.Mods == 0 {
					w.stopShortcutRecording(true)
					return true
				}
				acc, ok := acceleratorFromKeyDown(ev)
				if !ok {
					return ev.Kind == ui.InputKeyDown || ev.Kind == ui.InputKeyUp
				}
				if err := holdshortcut.ValidateAccelerator(acc); err != nil {
					mygo.Dialog.Error("Invalid shortcut", err.Error())
					w.stopShortcutRecording(true)
					return true
				}
				if err := w.persistShortcut(acc); err != nil {
					mygo.Dialog.Error("Could not save shortcut", err.Error())
					w.stopShortcutRecording(true)
				} else {
					w.shortcutSaved = true
					// persistShortcut already reloaded the global tap.
					w.stopShortcutRecording(false)
				}
				return true
			})
		})
		return
	}

	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, w.shortcut).Font("monospace").FontSize(12).
			Padding(6, 10).Radius(6).Background(t.Surface).Border(1, t.Border).Selectable()
		if ui.Button(c, "Record shortcut…").Clicked() {
			w.startShortcutRecording()
		}
	})
	if w.shortcutSaved {
		ui.Text(c, "Shortcut saved.").FontSize(12).TextColor(t.Success)
	}
}

func (w *Window) startShortcutRecording() {
	w.recordingShortcut = true
	w.shortcutSaved = false
	if w.onShortcutPause != nil {
		w.onShortcutPause(true)
	}
}

func (w *Window) stopShortcutRecording(resume bool) {
	if !w.recordingShortcut {
		return
	}
	w.recordingShortcut = false
	if resume && w.onShortcutPause != nil {
		w.onShortcutPause(false)
	}
}

func (w *Window) persistShortcut(accelerator string) error {
	if err := holdshortcut.ValidateAccelerator(accelerator); err != nil {
		return err
	}
	w.shortcut = accelerator
	if err := w.mutateConfig(func(cfg *config.Config) {
		cfg.PushToTalkShortcut = accelerator
	}); err != nil {
		return err
	}
	if w.onSaved != nil {
		w.onSaved()
	}
	return nil
}

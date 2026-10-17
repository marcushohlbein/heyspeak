package settings

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/diagnostics"
	"github.com/speek/speek/internal/holdshortcut"
	"github.com/speek/speek/internal/models"
)

const width, height = 520, 700

// Window is the compact settings panel opened from the tray.
type Window struct {
	store            *config.Store
	models           models.Manager
	win              *mygo.Window
	shortcut         string
	modelID          string
	language         string
	inputDeviceID    string
	micLabel         string
	micLabels        []string
	micIDs           []string
	restoreClipboard bool
	modelStatus      string
	diagnostics      string
	onSaved          func()
	shortcutH        ui.Handle
}

// New builds a hidden settings window.
func New(store *config.Store, modelMgr models.Manager, onSaved func()) *Window {
	cfg := store.Get()
	w := &Window{
		store:            store,
		models:           modelMgr,
		shortcut:         cfg.PushToTalkShortcut,
		modelID:          cfg.ModelID,
		language:         cfg.Language,
		inputDeviceID:    cfg.InputDeviceID,
		restoreClipboard: cfg.RestoreClipboard,
		onSaved:          onSaved,
	}
	w.refreshModelStatus()
	w.refreshMics()
	w.refreshDiagnostics()
	w.win = mygo.NewWindow(mygo.WindowOptions{
		Title:         "speek settings",
		Width:         width,
		Height:        height,
		Hidden:        true,
		DisableResize: true,
		Content:       ui.View(w.view),
	})
	w.win.OnClose(func(e *mygo.CloseEvent) {
		e.PreventDefault()
		w.win.Hide()
	})
	return w
}

// Show presents the settings window near the pointer display.
func (w *Window) Show() {
	cfg := w.store.Get()
	w.shortcut = cfg.PushToTalkShortcut
	w.modelID = cfg.ModelID
	w.language = cfg.Language
	w.inputDeviceID = cfg.InputDeviceID
	w.restoreClipboard = cfg.RestoreClipboard
	w.refreshModelStatus()
	w.refreshMics()
	w.refreshDiagnostics()
	d := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
	w.win.SetPosition(d.WorkArea.X+40, d.WorkArea.Y+40)
	w.win.Show()
	w.shortcutH.Focus()
	w.win.Update(func() {})
}

func (w *Window) refreshModelStatus() {
	cat, err := w.models.ListCatalog(context.Background())
	if err != nil {
		w.modelStatus = "Could not read model status"
		return
	}
	var lines []string
	for _, e := range cat {
		state := "not installed"
		if e.Installed {
			state = "installed"
		}
		lines = append(lines, fmt.Sprintf("%s (%s)", e.Entry.ID, state))
	}
	w.modelStatus = strings.Join(lines, " · ")
}

func (w *Window) refreshDiagnostics() {
	s := diagnostics.Collect()
	acc := "unknown"
	if s.AccessibilityTrusted {
		acc = "granted"
	} else if s.PasteSupported {
		acc = "not granted (required for paste)"
	}
	w.diagnostics = fmt.Sprintf("%s\nAccessibility: %s\nInput devices: %d", s.Platform, acc, s.InputDevices)
	if s.InputDeviceList != "" {
		w.diagnostics += "\n" + s.InputDeviceList
	}
}

func (w *Window) view(c *ui.Context) {
	t := c.Theme()
	c.Root().Background(t.Background)
	c.OnShortcut(ui.Cmd, ui.KeyW, w.win.Hide)

	ui.Column(c).Fill().Padding(20).Gap(10).Children(func() {
		ui.Text(c, "speek").FontSize(22)
		ui.Text(c, "Local push-to-talk dictation").TextColor(t.TextMuted)

		ui.Text(c, "Push-to-talk shortcut").FontSize(13).TextColor(t.TextMuted)
		ui.TextInputBase(c, &w.shortcut).Bind(&w.shortcutH).Placeholder("Ctrl+Shift+Space").Label("Shortcut")

		ui.Text(c, "Model ID").FontSize(13).TextColor(t.TextMuted)
		ui.TextInputBase(c, &w.modelID).Placeholder("ggml-tiny").Label("Model")
		ui.Text(c, w.modelStatus).FontSize(11).TextColor(t.TextMuted)

		ui.Row(c).Gap(8).Children(func() {
			if ui.Button(c, "Download model").Clicked() {
				w.downloadModel()
			}
			if ui.Button(c, "Remove model").Clicked() {
				w.removeModel()
			}
		})

		ui.Text(c, "Language").FontSize(13).TextColor(t.TextMuted)
		ui.TextInputBase(c, &w.language).Placeholder("auto (or en, de)").Label("Language")

		ui.Text(c, "Microphone").FontSize(13).TextColor(t.TextMuted)
		if len(w.micLabels) > 0 {
			ui.Select(c, &w.micLabel, w.micLabels)
		}
		if ui.Button(c, "Test microphone").Clicked() {
			w.testMic()
		}

		restoreLabel := "Restore clipboard after paste: on"
		if !w.restoreClipboard {
			restoreLabel = "Restore clipboard after paste: off"
		}
		if ui.Button(c, restoreLabel).Clicked() {
			w.restoreClipboard = !w.restoreClipboard
		}

		ui.Text(c, "Diagnostics").FontSize(13).TextColor(t.TextMuted)
		ui.Text(c, w.diagnostics).FontSize(11).TextColor(t.TextMuted)

		ui.Row(c).Gap(8).Children(func() {
			if ui.Button(c, "Save").Clicked() {
				w.save()
			}
			if ui.Button(c, "Close").Clicked() {
				w.win.Hide()
			}
		})

		ui.Text(c, w.store.Path()).FontSize(11).TextColor(t.TextMuted)
	})
}

func (w *Window) downloadModel() {
	id := strings.TrimSpace(w.modelID)
	if id == "" {
		id = models.DefaultModelID
		w.modelID = id
	}
	go func() {
		err := w.models.Download(context.Background(), id, func(p models.DownloadProgress) {
			if p.Complete || p.Received%5_000_000 < 32_000 {
				log.Printf("model %s: %d / %d bytes", p.ModelID, p.Received, p.Total)
			}
		})
		mygo.RunOnMain(func() {
			w.refreshModelStatus()
			w.win.Update(func() {})
			if err != nil {
				mygo.Dialog.Error("Download failed", err.Error())
				return
			}
			log.Printf("model %s downloaded and verified", id)
		})
	}()
}

func (w *Window) removeModel() {
	id := strings.TrimSpace(w.modelID)
	if id == "" {
		mygo.Dialog.Error("Remove model", "Enter a model ID to remove.")
		return
	}
	go func() {
		err := w.models.Remove(context.Background(), id)
		mygo.RunOnMain(func() {
			w.refreshModelStatus()
			w.win.Update(func() {})
			if err != nil {
				mygo.Dialog.Error("Remove failed", err.Error())
				return
			}
		})
	}()
}

func (w *Window) refreshMics() {
	devs, err := audio.ListInputDevices()
	w.micLabels = []string{"System default"}
	w.micIDs = []string{""}
	if err != nil {
		w.micLabel = w.micLabels[0]
		return
	}
	used := map[string]int{}
	for _, d := range devs {
		label := d.Name
		if label == "" {
			label = "Microphone"
		}
		if d.IsDefault {
			label += " (system default)"
		}
		used[label]++
		if used[label] > 1 {
			label = fmt.Sprintf("%s %d", label, used[label])
		}
		w.micLabels = append(w.micLabels, label)
		w.micIDs = append(w.micIDs, d.ID)
	}
	w.micLabel = w.micLabels[0]
	for i, id := range w.micIDs {
		if id == w.inputDeviceID {
			w.micLabel = w.micLabels[i]
			break
		}
	}
}

func (w *Window) selectedMicID() string {
	for i, label := range w.micLabels {
		if label == w.micLabel && i < len(w.micIDs) {
			return w.micIDs[i]
		}
	}
	return ""
}

func (w *Window) testMic() {
	id := w.selectedMicID()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		defer cancel()
		result, err := audio.Probe(ctx, id)
		mygo.RunOnMain(func() {
			if err != nil && ctx.Err() == nil {
				mygo.Dialog.Error("Microphone test", err.Error())
				return
			}
			msg := fmt.Sprintf("Recorded %.0f ms, peak %d.", result.DurationMS, result.Peak)
			switch {
			case result.Samples == 0 || result.Peak < 120:
				msg += "\n\nAlmost no signal. Allow Microphone access for speek, or choose another device."
			case result.Peak < 800:
				msg += "\n\nThe signal is very quiet. Dictation may miss words."
			default:
				msg += "\n\nThis microphone is receiving sound. Save settings, then try dictation."
			}
			_, _ = mygo.Dialog.Message(mygo.MessageOptions{
				Type:    mygo.MessageInfo,
				Message: "Microphone test",
				Detail:  msg,
			})
		})
	}()
}

func (w *Window) save() {
	shortcut := strings.TrimSpace(w.shortcut)
	if err := holdshortcut.ValidateAccelerator(shortcut); err != nil {
		mygo.Dialog.Error("Invalid shortcut", err.Error())
		return
	}
	cfg := w.store.Get()
	// Pick up external edits made while the window was open, then apply the
	// window's fields on top. Keys the settings UI does not show survive.
	if err := w.store.Reload(); err != nil {
		log.Printf("config reload: %v", err)
	} else {
		cfg = w.store.Get()
	}
	cfg.PushToTalkShortcut = shortcut
	cfg.ModelID = strings.TrimSpace(w.modelID)
	cfg.Language = strings.TrimSpace(w.language)
	cfg.InputDeviceID = w.selectedMicID()
	cfg.RestoreClipboard = w.restoreClipboard
	if err := w.store.Set(cfg); err != nil {
		mygo.Dialog.Error("Could not save settings", err.Error())
		return
	}
	w.refreshModelStatus()
	w.refreshDiagnostics()
	if w.onSaved != nil {
		w.onSaved()
	}
}

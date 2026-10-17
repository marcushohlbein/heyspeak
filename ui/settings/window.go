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
	"github.com/speek/speek/ui/permissions"
)

const (
	width, height   = 780, 640
	titleBarHeight  = 52
)

const (
	sectionGeneral     = "general"
	sectionAudio       = "audio"
	sectionModels      = "models"
	sectionPermissions = "permissions"
	sectionAdvanced    = "advanced"
)

// Window is the settings panel opened from the tray.
type Window struct {
	store            *config.Store
	models           models.Manager
	win              *mygo.Window
	section          string
	shortcut         string
	modelID          string
	language         string
	inputDeviceID    string
	micLabel         string
	micLabels        []string
	micIDs           []string
	micLevel         float32
	testingMic       bool
	micTested        bool
	micCancel        context.CancelFunc
	restoreClipboard bool
	modelCatalog     []models.EntryStatus
	modelCatalogErr  string
	modelFamilyLang  map[string]string
	modelDownloadID  string
	modelDownload    models.DownloadProgress
	diagnostics      string
	mic              audio.MicStatus
	accessOK         bool
	onSaved          func()
	onShortcutPause  func(paused bool)
	recordingShortcut bool
	shortcutSaved     bool
	shortcutCaptureH  ui.Handle
}

// New builds a hidden settings window.
func New(store *config.Store, modelMgr models.Manager, onSaved func()) *Window {
	cfg := store.Get()
	w := &Window{
		store:            store,
		models:           modelMgr,
		section:          sectionGeneral,
		shortcut:         cfg.PushToTalkShortcut,
		modelID:          cfg.ModelID,
		language:         cfg.Language,
		inputDeviceID:    cfg.InputDeviceID,
		restoreClipboard: cfg.RestoreClipboard,
		onSaved:          onSaved,
	}
	w.refreshModelCatalog()
	w.refreshMics()
	w.refreshDiagnostics()
	w.refreshPermissions()
	w.win = mygo.NewWindow(mygo.WindowOptions{
		Title:          "speek settings",
		Width:          width,
		Height:         height,
		Hidden:         true,
		TitleBarStyle:  mygo.TitleBarHiddenInset,
		TitleBarHeight: titleBarHeight,
		Vibrancy:       mygo.VibrancySidebar,
		Content:        ui.View(w.view),
	})
	w.win.OnClose(func(e *mygo.CloseEvent) {
		e.PreventDefault()
		w.hide()
	})
	return w
}

// SetShortcutPause is called with true while a new chord is being recorded,
// so the global hold-to-talk tap does not swallow that keypress.
func (w *Window) SetShortcutPause(fn func(paused bool)) {
	w.onShortcutPause = fn
}

// Show presents the settings window near the pointer display.
func (w *Window) Show() {
	w.ShowSection(sectionGeneral)
}

// ShowSection opens settings with a sidebar section selected.
func (w *Window) ShowSection(section string) {
	if section != "" {
		w.section = section
	}
	w.reloadFromStore()
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyRegular)
	mygo.App.Show()
	mygo.App.Focus()
	d := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
	x, y := centerInWorkArea(d.WorkArea, width, height)
	w.win.SetPosition(x, y)
	w.win.Show()
	w.win.Focus()
	w.win.Update(func() {})
}

func centerInWorkArea(area mygo.Rectangle, winW, winH int) (int, int) {
	return area.X + (area.Width-winW)/2, area.Y + (area.Height-winH)/2
}

func (w *Window) hide() {
	w.stopShortcutRecording(true)
	if w.testingMic && w.micCancel != nil {
		// Hide cancels the capture. Finalize sees testingMic false and
		// does not show an error dialog for the cancel.
		w.testingMic = false
		w.micTested = true
		w.micCancel()
		w.micCancel = nil
	}
	w.win.Hide()
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
}

func (w *Window) reloadFromStore() {
	cfg := w.store.Get()
	w.shortcut = cfg.PushToTalkShortcut
	w.modelID = cfg.ModelID
	w.language = cfg.Language
	w.inputDeviceID = cfg.InputDeviceID
	w.restoreClipboard = cfg.RestoreClipboard
	w.refreshModelCatalog()
	w.refreshMics()
	w.refreshDiagnostics()
	w.refreshPermissions()
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
	vibrant := c.Vibrancy()
	if vibrant {
		c.Root().Background(ui.Transparent)
	} else {
		c.Root().Background(t.Background)
	}
	c.OnShortcut(ui.Cmd, ui.KeyW, w.hide)

	bar := c.TitleBar()
	// The window controls are centered in bar.Height (native strip on
	// macOS); fall back to a standard-height strip elsewhere.
	barH := max(bar.Height, 28)

	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		sidebar := ui.Column(c).Width(220).Shrink(0).BorderEndWidth(1).BorderColor(t.Border)
		if !vibrant {
			sidebar.Background(t.Surface)
		}
		sidebar.Children(func() {
			ui.Box(c).Height(barH).Shrink(0).DragWindow()
			list := ui.Sidebar(c, &w.section, func() {
				ui.SidebarItem(c, sectionGeneral, iconSliders, "General").Margin(3, 0)
				ui.SidebarItem(c, sectionModels, iconPackage, "Models").Margin(3, 0)
				ui.SidebarItem(c, sectionAudio, iconMic, "Audio").Margin(3, 0)
				perms := ui.SidebarItem(c, sectionPermissions, iconShield, "Permissions").Margin(3, 0)
				if permissions.NeedsAttention() {
					perms.Children(func() { ui.Badge(c, "!") })
				}
				ui.SidebarItem(c, sectionAdvanced, iconWrench, "Advanced").Margin(3, 0)
			}).Grow(1).Label("Settings sections")
			if vibrant {
				list.Background(ui.Transparent)
			}
			if list.Changed() {
				if w.recordingShortcut && w.section != sectionGeneral {
					w.stopShortcutRecording(true)
				}
				switch w.section {
				case sectionPermissions:
					w.refreshPermissions()
				case sectionModels:
					w.refreshModelCatalog()
				}
			}
		})

		ui.Column(c).Grow(1).MinWidth(0).Background(t.Background).Children(func() {
			ui.Row(c).Shrink(0).Height(barH).AlignItems(ui.Center).
				Padding(0, bar.Right+16, 0, 20).Children(func() {
				ui.Text(c, w.sectionTitle()).FontSize(15).Bold().SingleLine()
				ui.Box(c).Grow(1).FillHeight().DragWindow()
				closeBtn := ui.ButtonBase(c).Padding(6).Radius(t.Radius).Background(t.Surface).Label("Close settings")
				closeBtn.Children(func() { ui.Icon(c, iconX) })
				if closeBtn.Clicked() {
					w.hide()
				}
			})
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Padding(4, 28, 24).Gap(16).MaxWidth(560).Children(func() {
					if sub := w.sectionSubtitle(); sub != "" {
						ui.Text(c, sub).FontSize(12).TextColor(t.TextMuted)
					}
					switch w.section {
					case sectionGeneral:
						w.generalPanel(c)
					case sectionAudio:
						w.audioPanel(c)
					case sectionModels:
						w.modelsPanel(c)
					case sectionPermissions:
						w.permissionsPanel(c)
					case sectionAdvanced:
						w.advancedPanel(c)
					default:
						w.generalPanel(c)
					}
				})
			})
			ui.Divider(c)
			ui.Row(c).Gap(8).Padding(12, 20).Children(func() {
				ui.Spacer(c)
				if ui.PrimaryButton(c, "Save").Clicked() {
					w.save()
				}
			})
		})
	})
}

// card groups a setting cluster in a rounded panel.
func card(c *ui.Context, fn func()) {
	t := c.Theme()
	ui.Column(c).Gap(10).Padding(14, 16).Radius(t.Radius).
		Background(t.Surface).Border(1, t.Border).Children(fn)
}

func (w *Window) sectionTitle() string {
	switch w.section {
	case sectionGeneral:
		return "General"
	case sectionAudio:
		return "Audio"
	case sectionModels:
		return "Models"
	case sectionPermissions:
		return "Permissions"
	case sectionAdvanced:
		return "Advanced"
	default:
		return "Settings"
	}
}

func (w *Window) sectionSubtitle() string {
	switch w.section {
	case sectionGeneral:
		return "Shortcut, language, and clipboard behavior."
	case sectionAudio:
		return "Choose the microphone used for dictation."
	case sectionModels:
		return "Download, remove, and choose the Whisper model used for dictation."
	case sectionPermissions:
		return "System access speek needs for dictation and paste."
	case sectionAdvanced:
		return "Runtime diagnostics and configuration path."
	default:
		return ""
	}
}

func (w *Window) generalPanel(c *ui.Context) {
	card(c, func() {
		ui.Text(c, "Dictation shortcut").FontSize(13).Bold()
		w.shortcutPanel(c)
	})
	card(c, func() {
		ui.Text(c, "Transcription").FontSize(13).Bold()
		ui.Field(c, "Language", func() {
			ui.TextInput(c, &w.language).Placeholder("auto (or en, de)")
		})
		ui.Checkbox(c, &w.restoreClipboard, "Restore clipboard after paste")
	})
}

func (w *Window) audioPanel(c *ui.Context) {
	t := c.Theme()
	card(c, func() {
		ui.Text(c, "Input device").FontSize(13).Bold()
		ui.Text(c, "The microphone dictation records from.").FontSize(12).TextColor(t.TextMuted)
		if len(w.micLabels) > 0 {
			ui.Select(c, &w.micLabel, w.micLabels)
		}
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			if ui.Button(c, w.testButtonLabel()).AlignSelf(ui.Start).Clicked() {
				w.toggleMicTest()
			}
			if w.testingMic || w.micTested {
				ui.Spacer(c)
				w.levelBar(c, w.micLevel)
			}
		})
	})
}

func (w *Window) testButtonLabel() string {
	if w.testingMic {
		return "Stop test"
	}
	return "Test microphone"
}

// levelBar is the small input-level meter next to the test button. Level
// maps the int16 peak, full scale ~8000: normal speech lands mid-bar.
func (w *Window) levelBar(c *ui.Context, level float32) {
	t := c.Theme()
	const barWidth = 120
	color := t.Success
	switch {
	case level < 0.02:
		color = t.Danger
	case level < 0.1:
		color = t.Warning
	}
	ui.Box(c).Width(barWidth).Height(8).Radius(4).
		Background(t.Text.Alpha(0.12)).Children(func() {
		if level <= 0 {
			return
		}
		fill := int(barWidth * level)
		if fill < 4 {
			fill = 4
		}
		ui.Box(c).Margin(1).Width(float32(fill)).FillHeight().Radius(3).Background(color)
	})
}

func (w *Window) advancedPanel(c *ui.Context) {
	t := c.Theme()
	card(c, func() {
		ui.Text(c, "Diagnostics").FontSize(13).Bold()
		ui.Text(c, w.diagnostics).FontSize(12).TextColor(t.TextMuted).Selectable()
	})
	card(c, func() {
		ui.Text(c, "Config file").FontSize(13).Bold()
		ui.Text(c, w.store.Path()).FontSize(12).TextColor(t.TextMuted).Selectable()
	})
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

// toggleMicTest starts a mic test, or stops the running one. The capture
// runs until the user stops it or the window hides; the level meter fills
// while it runs and the summary dialog shows the verdict when it ends.
func (w *Window) toggleMicTest() {
	if w.testingMic {
		w.micCancel()
		return
	}
	id := w.selectedMicID()
	ctx, cancel := context.WithCancel(context.Background())
	w.micCancel = cancel
	w.testingMic = true
	w.micLevel = 0
	w.win.Update(func() {})
	go func() {
		// Throttle on the capture thread: repaint at most every 40 ms.
		var last time.Time
		_, err := audio.Probe(ctx, id, func(peak int) {
			now := time.Now()
			if now.Sub(last) < 40*time.Millisecond {
				return
			}
			last = now
			level := float32(peak) / 8000
			if level > 1 {
				level = 1
			}
			mygo.RunOnMain(func() {
				w.micLevel = level
				w.win.Update(func() {})
			})
		})
		mygo.RunOnMain(func() {
			if !w.testingMic {
				// The window hid and already wound the test down.
				return
			}
			w.testingMic = false
			w.micTested = true
			w.micCancel = nil
			w.win.Update(func() {})
			if err != nil {
				mygo.Dialog.Error("Microphone test", err.Error())
			}
			// Success needs no popup: the level bar next to the button
			// already shows the verdict.
		})
	}()
}

func (w *Window) save() {
	shortcut := strings.TrimSpace(w.shortcut)
	if err := holdshortcut.ValidateAccelerator(shortcut); err != nil {
		mygo.Dialog.Error("Invalid shortcut", err.Error())
		return
	}
	w.shortcutSaved = false
	if err := w.mutateConfig(func(cfg *config.Config) {
		cfg.PushToTalkShortcut = shortcut
		cfg.ModelID = strings.TrimSpace(w.modelID)
		cfg.Language = strings.TrimSpace(w.language)
		cfg.InputDeviceID = w.selectedMicID()
		cfg.RestoreClipboard = w.restoreClipboard
	}); err != nil {
		mygo.Dialog.Error("Could not save settings", err.Error())
		return
	}
	w.refreshModelCatalog()
	w.refreshDiagnostics()
	w.refreshPermissions()
	if w.onSaved != nil {
		w.onSaved()
	}
}

// mutateConfig reloads disk, applies fn, and writes the result.
func (w *Window) mutateConfig(fn func(*config.Config)) error {
	cfg := w.store.Get()
	if err := w.store.Reload(); err != nil {
		log.Printf("config reload: %v", err)
	} else {
		cfg = w.store.Get()
	}
	fn(&cfg)
	return w.store.Set(cfg)
}

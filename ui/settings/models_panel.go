package settings

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
)

func (w *Window) refreshModelCatalog() {
	cat, err := w.models.ListCatalog(context.Background())
	if err != nil {
		w.modelCatalog = nil
		w.modelCatalogErr = "Could not read installed models"
		return
	}
	w.modelCatalog = cat
	w.modelCatalogErr = ""
	w.syncModelFamilyPicks()
}

func (w *Window) syncModelFamilyPicks() {
	if w.modelFamilyLang == nil {
		w.modelFamilyLang = map[string]string{}
	}
	if w.modelID != "" {
		if fam, variant, ok := models.FamilyForModelID(w.modelID); ok {
			w.modelFamilyLang[fam.Key] = variant.Label
		}
	}
	for _, fam := range models.DefaultFamilies() {
		if w.modelFamilyLang[fam.Key] != "" {
			continue
		}
		w.modelFamilyLang[fam.Key] = fam.Variants[0].Label
	}
}

func (w *Window) modelsPanel(c *ui.Context) {
	t := c.Theme()

	card(c, func() {
		ui.Text(c, "Active model").FontSize(13).Bold()
		if w.modelID != "" {
			activeName := w.modelID
			if fam, variant, ok := models.FamilyForModelID(w.modelID); ok {
				activeName = fmt.Sprintf("%s (%s)", fam.Name, variant.Label)
			}
			ui.Text(c, fmt.Sprintf("%s · %s", activeName, w.modelID)).FontSize(12).TextColor(t.TextMuted).Selectable()
		} else {
			ui.Text(c, "Choose an installed model below, or download one first.").
				FontSize(12).TextColor(t.TextMuted)
		}
		if w.modelCatalogErr != "" {
			ui.Text(c, w.modelCatalogErr).FontSize(12).TextColor(t.Warning)
		}
	})

	ui.Text(c, "Available models").FontSize(13).Bold()
	ui.Text(c, "Pick a size tier and language variant, then download or activate it.").
		FontSize(12).TextColor(t.TextMuted)

	for _, fam := range models.DefaultFamilies() {
		fam := fam
		card(c, func() {
			w.modelFamilyBody(c, fam)
		})
	}
}

func (w *Window) modelFamilyBody(c *ui.Context, fam models.ModelFamily) {
	t := c.Theme()
	labels := w.familyVariantLabels(fam)
	lang := w.familyLangValue(fam)
	familyActive := w.familyContainsModelID(fam, w.modelID)

	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Gap(2).Children(func() {
			ui.Row(c).Gap(8).Children(func() {
				ui.Text(c, fam.Name).FontSize(14).Bold()
				if familyActive {
					ui.Badge(c, "Active")
				}
			})
		})
		if len(labels) > 1 {
			ui.Select(c, &lang, labels).Label("Language")
		} else if len(labels) == 1 {
			ui.Text(c, labels[0]).FontSize(12).TextColor(t.TextMuted)
			lang = labels[0]
		}
	})
	w.modelFamilyLang[fam.Key] = lang

	id := w.variantIDForFamily(fam, lang)
	st := w.entryStatus(id)
	active := w.modelID == id
	downloading := w.modelDownloadID == id
	installed := st.Installed

	ui.Text(c, fmt.Sprintf("%s · %s", formatModelSize(st.Entry.SizeBytes), fam.Requirements)).
		FontSize(11).TextColor(t.TextMuted)

	ui.Text(c, id).FontSize(11).TextColor(t.TextMuted).Selectable()

	status := "Not installed"
	statusColor := t.TextMuted
	switch {
	case downloading:
		status = "Downloading…"
		statusColor = t.Accent
	case installed:
		status = "Installed"
		statusColor = t.Success
	}
	ui.Text(c, status).FontSize(12).TextColor(statusColor)

	if other := w.otherInstalledVariantLabel(fam, id); other != "" {
		ui.Text(c, other).FontSize(11).TextColor(t.TextMuted)
	}

	if downloading && w.modelDownload.Total > 0 {
		frac := float64(w.modelDownload.Received) / float64(w.modelDownload.Total)
		ui.Progress(c, frac).Grow(1)
		ui.Text(c, fmt.Sprintf("%s / %s", formatModelSize(w.modelDownload.Received), formatModelSize(w.modelDownload.Total))).
			FontSize(11).TextColor(t.TextMuted)
	}

	ui.Row(c).Gap(8).Wrap().Children(func() {
		switch {
		case downloading:
			if ui.Button(c, "Cancel download").Clicked() {
				w.models.CancelDownload(id)
			}
		case installed:
			if !active {
				if ui.PrimaryButton(c, "Use for dictation").Clicked() {
					w.selectModel(id)
				}
			}
			rm := ui.ButtonBase(c).Padding(t.Space(1.5), t.Space(3.5)).Radius(t.Radius).
				Background(t.Surface).Border(1, t.Border).TextColor(t.Danger).Label("Remove model")
			rm.Children(func() { ui.Text(c, "Remove") })
			if rm.Clicked() {
				w.removeModel(id)
			}
		default:
			if ui.PrimaryButton(c, "Download").Clicked() {
				w.downloadModel(id)
			}
		}
	})
}

func (w *Window) familyVariantLabels(fam models.ModelFamily) []string {
	var labels []string
	for _, v := range fam.Variants {
		labels = append(labels, v.Label)
	}
	return labels
}

func (w *Window) familyLangValue(fam models.ModelFamily) string {
	if w.modelFamilyLang == nil {
		w.modelFamilyLang = map[string]string{}
	}
	if w.modelFamilyLang[fam.Key] == "" {
		w.modelFamilyLang[fam.Key] = fam.Variants[0].Label
	}
	return w.modelFamilyLang[fam.Key]
}

func (w *Window) variantIDForFamily(fam models.ModelFamily, label string) string {
	for _, v := range fam.Variants {
		if v.Label == label {
			return v.ID
		}
	}
	return fam.Variants[0].ID
}

func (w *Window) familyContainsModelID(fam models.ModelFamily, id string) bool {
	for _, v := range fam.Variants {
		if v.ID == id {
			return true
		}
	}
	return false
}

func (w *Window) catalogStatus(id string) (models.EntryStatus, bool) {
	for _, st := range w.modelCatalog {
		if st.Entry.ID == id {
			return st, true
		}
	}
	return models.EntryStatus{}, false
}

func (w *Window) entryStatus(id string) models.EntryStatus {
	st, ok := w.catalogStatus(id)
	if ok {
		return st
	}
	entry, found := models.LookupCatalogEntry(id)
	if !found {
		entry.ID = id
	}
	return models.EntryStatus{Entry: entry}
}

func (w *Window) otherInstalledVariantLabel(fam models.ModelFamily, selectedID string) string {
	var installed []string
	for _, v := range fam.Variants {
		if v.ID == selectedID {
			continue
		}
		st, ok := w.catalogStatus(v.ID)
		if ok && st.Installed {
			installed = append(installed, v.Label)
		}
	}
	if len(installed) == 0 {
		return ""
	}
	return "Also installed: " + strings.Join(installed, ", ")
}

func formatModelSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := float64(bytes) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(suffix) {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f %s", val, suffix[exp])
}

func (w *Window) selectModel(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	w.modelID = id
	w.syncModelFamilyPicks()
	if err := w.mutateConfig(func(cfg *config.Config) {
		cfg.ModelID = id
	}); err != nil {
		mygo.Dialog.Error("Could not save model", err.Error())
		return
	}
	w.win.Update(func() {})
}

func (w *Window) downloadModel(id string) {
	if w.modelDownloadID != "" {
		mygo.Dialog.Error("Download in progress", "Wait for the current download to finish or cancel it first.")
		return
	}
	w.modelDownloadID = id
	go func() {
		err := w.models.Download(context.Background(), id, func(p models.DownloadProgress) {
			mygo.RunOnMain(func() {
				w.modelDownload = p
				w.win.Update(func() {})
			})
		})
		mygo.RunOnMain(func() {
			if w.modelDownloadID == id {
				w.modelDownloadID = ""
			}
			w.refreshModelCatalog()
			w.win.Update(func() {})
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				mygo.Dialog.Error("Download failed", err.Error())
				return
			}
			log.Printf("model %s downloaded and verified", id)
			if w.modelID == "" {
				w.selectModel(id)
			}
		})
	}()
}

func (w *Window) removeModel(id string) {
	go func() {
		err := w.models.Remove(context.Background(), id)
		mygo.RunOnMain(func() {
			w.refreshModelCatalog()
			cfg := w.store.Get()
			w.modelID = cfg.ModelID
			w.syncModelFamilyPicks()
			w.win.Update(func() {})
			if err != nil {
				mygo.Dialog.Error("Remove failed", err.Error())
			}
		})
	}()
}

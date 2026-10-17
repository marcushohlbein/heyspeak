package settings

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
)

// TestRenderSections renders every section headlessly: it fails if a view
// panics and leaves PNGs in the output dir for eyeballing.
func TestRenderSections(t *testing.T) {
	out := os.Getenv("SETTINGS_RENDER_DIR")
	store, err := config.Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	w := &Window{
		store:            store,
		models:           &models.FakeManager{},
		section:          sectionGeneral,
		shortcut:         "⌥ space",
		modelID:          "ggml-small.en",
		language:         "auto",
		micLabel:         "System default",
		micLabels:        []string{"System default"},
		micIDs:           []string{""},
		micLevel:         0.7,
		micTested:        true,
		restoreClipboard: true,
		diagnostics:      "darwin/arm64\nAccessibility: granted",
	}
	w.syncModelFamilyPicks()
	for _, section := range []string{sectionGeneral, sectionModels, sectionAudio, sectionPermissions, sectionAdvanced} {
		w.section = section
		tr := ui.NewTester(w.view, width, height)
		tr.SetDark(true)
		// Match the real window: native-size title bar with inset controls.
		tr.SetTitleBar(ui.TitleBar{Height: 28, Left: 78})
		tr.SetVibrancy(true)
		tr.Frame()
		img := tr.Image()
		if img == nil {
			t.Fatalf("%s: no image rendered", section)
		}
		if out == "" {
			continue
		}
		f, err := os.Create(filepath.Join(out, section+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}

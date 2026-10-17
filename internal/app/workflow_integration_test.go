package app

import (
	"strings"
	"testing"

	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/models"
	"github.com/speek/speek/internal/overlay"
	"github.com/speek/speek/internal/paste"
)

// End-to-end coordinator flow with fakes (Phase 9 integration).
func TestWorkflowIntegration(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	ov := &overlay.FakeController{}
	pasteFake := svc.Paste.(*paste.FakeService)
	svc.Overlay = ov

	var statuses []string
	c, sc := testCoordinator(svc, store, Options{
		OnError: func(msg, _ string) { statuses = append(statuses, msg) },
	})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateRecording })
	sc.Events.OnRelease()
	waitFor(t, func() bool { return c.State() == StateIdle })

	if pasteFake.Last != "hello" {
		t.Fatalf("paste %q", pasteFake.Last)
	}
	if pasteFake.Captures != 1 {
		t.Fatalf("captures %d", pasteFake.Captures)
	}
}

func TestWorkflowBlocksWithoutModel(t *testing.T) {
	store, _ := config.Open(t.TempDir() + "/cfg.json")
	svc := defaultTestServices()
	svc.Models = &models.FakeManager{Selected: ""}
	ov := &overlay.FakeController{}
	svc.Overlay = ov
	c, sc := testCoordinator(svc, store, Options{})

	sc.Events.OnPress()
	waitFor(t, func() bool { return c.State() == StateIdle })
	if !strings.Contains(c.TrayTooltip(), "not installed") {
		t.Fatalf("tooltip %q", c.TrayTooltip())
	}
	sc.Events.OnRelease()
	if c.State() != StateIdle {
		t.Fatal("unexpected recording")
	}
}

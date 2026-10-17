package tray

import (
	"github.com/egoist/mygo"
	"github.com/speek/speek/internal/app"
)

// UI owns the tray icon and keeps menu state in sync with the coordinator.
type UI struct {
	tray   *mygo.Tray
	cancel *mygo.MenuItem
	coord  *app.Coordinator
}

// Install creates the background tray icon and menu.
func Install(coord *app.Coordinator, onPermissions, onSettings func(), onQuit func()) (*UI, error) {
	ui := &UI{coord: coord}
	tray, err := mygo.NewTray(mygo.TrayOptions{
		Icon:           trayIcon(),
		IconIsTemplate: true,
		Title:          "speek",
		ToolTip:        "speek — click for Settings",
	})
	if err != nil {
		return nil, err
	}
	ui.tray = tray
	ui.cancel = &mygo.MenuItem{
		Label:    "Cancel dictation",
		Disabled: true,
		Click: func(*mygo.MenuItem, *mygo.Window) {
			coord.CancelDictation()
		},
	}
	tray.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		ui.cancel,
		{Label: "Permissions…", Click: func(*mygo.MenuItem, *mygo.Window) { onPermissions() }},
		{Label: "Settings…", Click: func(*mygo.MenuItem, *mygo.Window) { onSettings() }},
		{Role: mygo.RoleQuit, Click: func(*mygo.MenuItem, *mygo.Window) { onQuit() }},
	}))
	ui.sync()
	return ui, nil
}

// OnStateChange refreshes the tray from the coordinator.
func (u *UI) OnStateChange(app.State) {
	u.sync()
}

func (u *UI) sync() {
	if u.cancel != nil {
		u.cancel.SetEnabled(u.coord.CanCancel())
	}
	tip := u.coord.TrayTooltip()
	if tip == "" {
		tip = "speek"
	}
	u.tray.SetToolTip(tip)
}

// Destroy removes the tray icon.
func (u *UI) Destroy() {
	if u.tray != nil {
		u.tray.Destroy()
	}
}

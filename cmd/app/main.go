// speek — local push-to-talk dictation.
//
//	go run ./cmd/app
package main

import (
	"flag"
	"log"

	"github.com/egoist/mygo"
	"github.com/speek/speek/internal/app"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/holdshortcut"
	"github.com/speek/speek/internal/overlay"
	"github.com/speek/speek/internal/singleinstance"
	"github.com/speek/speek/ui/permissions"
	"github.com/speek/speek/ui/settings"
	"github.com/speek/speek/ui/tray"
)

func main() {
	openSettings := flag.Bool("settings", false, "open the settings window")
	flag.Parse()

	if runCLI(flag.Args()) {
		return
	}

	release, err := singleinstance.Acquire()
	if err != nil {
		log.Fatal(err)
	}
	defer release()

	store, err := config.OpenDefault()
	if err != nil {
		log.Fatal(err)
	}

	services := app.DefaultServices(store)
	var trayUI *tray.UI
	coordinator := app.NewCoordinator(services, store, app.Options{
		OnStateChange: func(s app.State) {
			if trayUI != nil {
				trayUI.OnStateChange(s)
			}
		},
		OnError: func(message, recognized string) {
			detail := message
			if recognized != "" {
				detail = message + "\n\n" + recognized
			}
			mygo.Dialog.Error("Dictation failed", detail)
		},
	})

	mygo.App.SetName("speek")
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.OnWindowAllClosed(func() {})

	var settingsWin *settings.Window
	var permissionsWin *permissions.Window

	mygo.App.WhenReady(func() {
		if mgr, ok := services.Overlay.(*overlay.Manager); ok {
			mgr.RegisterDisplayChanges()
		}

		settingsWin = settings.New(store, services.Models, func() {
			if err := coordinator.ReloadShortcut(); err != nil {
				log.Printf("shortcut reload: %v", err)
			}
		})

		permissionsWin = permissions.New()

		trayUI, err = tray.Install(coordinator, permissionsWin.Show, settingsWin.Show, mygo.App.Quit)
		if err != nil {
			log.Printf("tray: %v", err)
			mygo.Dialog.Error("speek tray unavailable",
				"Could not create the menu bar icon.\n\nQuit speek first (menu bar → Quit), then run:\n\n  speek -settings\n\nOr from the repo:\n  ./bin/speek -settings")
		}

		if *openSettings {
			settingsWin.Show()
		}
		if permissions.NeedsAttention() {
			permissionsWin.Show()
		}

		if err := coordinator.StartShortcut(); err != nil {
			if err == holdshortcut.ErrUnsupported {
				log.Printf("push-to-talk: %v (overlay-only on this platform)", err)
			} else {
				log.Printf("push-to-talk: %v (check Accessibility on macOS)", err)
			}
		}
	})

	mygo.App.OnWillQuit(func(*mygo.QuitEvent) {
		coordinator.Shutdown()
		if closer, ok := services.Transcription.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
		_ = coordinator.StopShortcut()
		services.Overlay.Close()
		if trayUI != nil {
			trayUI.Destroy()
		}
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

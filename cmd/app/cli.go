package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/diagnostics"
	"github.com/speek/speek/internal/models"
	"github.com/speek/speek/internal/paste"
)

// runCLI handles headless subcommands and reports whether the process is
// finished (exit without starting the GUI). Commands run before the
// single-instance lock so they work while the tray app is running.
func runCLI(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "status":
		cliStatus()
		return true
	case "model":
		if err := cliModel(args[1:]); err != nil {
			cliUsage()
			log.Fatal(err)
		}
		return true
	case "help":
		cliUsage()
		return true
	default:
		log.Fatalf("speek: unknown command %q", args[0])
		return true
	}
}

func cliUsage() {
	fmt.Print(`usage: speek [-settings] [command]

commands:
  status                 print permissions, devices, and model state
  model list             list catalog models and install state
  model download [id]    download and verify a model (default: ggml-tiny)
  model remove <id>      remove an installed model
`)
}

func cliStatus() {
	s := diagnostics.Collect()
	fmt.Printf("platform:      %s\n", s.Platform)
	fmt.Printf("microphone:    %s\n", audio.MicrophoneStatus().Label())
	fmt.Printf("accessibility: %t\n", paste.AccessibilityTrusted())
	if s.InputDeviceList != "" {
		fmt.Printf("input devices:\n%s\n", s.InputDeviceList)
	} else {
		fmt.Println("input devices: none listed")
	}

	store, err := config.OpenDefault()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	mgr := models.NewStore(store)
	cat, err := mgr.ListCatalog(context.Background())
	if err != nil {
		log.Fatalf("models: %v", err)
	}
	selected := store.Get().ModelID
	if selected == "" {
		if id, err := mgr.EffectiveModelID(context.Background()); err == nil {
			selected = id + " (auto)"
		}
	}
	fmt.Printf("selected model: %s\n", selected)
	for _, e := range cat {
		state := "not installed"
		if e.Installed {
			state = "installed"
		}
		fmt.Printf("  %-16s %s\n", e.Entry.ID, state)
	}
	fmt.Printf("config:        %s\n", store.Path())
}

func openModels() *models.Store {
	store, err := config.OpenDefault()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	return models.NewStore(store)
}

func cliModel(args []string) error {
	if len(args) == 0 {
		return errors.New("speek model: expected list, download, or remove")
	}
	switch args[0] {
	case "list":
		cat, err := openModels().ListCatalog(context.Background())
		if err != nil {
			log.Fatalf("models: %v", err)
		}
		for _, e := range cat {
			state := "not installed"
			if e.Installed {
				state = "installed"
			}
			fmt.Printf("%-16s %s\n", e.Entry.ID, state)
		}
		return nil
	case "download":
		id := models.DefaultModelID
		if len(args) > 1 && args[1] != "" {
			id = args[1]
		}
		fmt.Printf("downloading %s…\n", id)
		err := openModels().Download(context.Background(), id, func(p models.DownloadProgress) {
			if p.Complete {
				fmt.Printf("\r%s: %d bytes, verified  \n", p.ModelID, p.Received)
				return
			}
			fmt.Printf("\r%s: %d / %d bytes", p.ModelID, p.Received, p.Total)
		})
		if err != nil {
			log.Fatalf("download: %v", err)
		}
		return nil
	case "remove":
		if len(args) < 2 || args[1] == "" {
			log.Fatal("speek model remove <id>")
		}
		if err := openModels().Remove(context.Background(), args[1]); err != nil {
			log.Fatalf("remove: %v", err)
		}
		fmt.Printf("removed %s\n", args[1])
		return nil
	default:
		return fmt.Errorf("speek model: unknown command %q", args[0])
	}
}

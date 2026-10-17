// Spike 1C: record while PTT held, then clipboard + synthetic paste.
//
//	go run ./cmd/spike-capture-paste
//
// Hold Ctrl+Shift+Space to record; release to stop, paste "speek spike test".
// Requires Accessibility (PTT + paste). Microphone permission on first capture.
package main

import (
	"context"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/config"
	"github.com/speek/speek/internal/holdshortcut"
	"github.com/speek/speek/internal/paste"
)

func main() {
	if runtime.GOOS != "darwin" {
		log.Fatal("spike-capture-paste: macOS only in Phase 1")
	}

	acc := holdshortcut.DefaultAccelerator
	if v := os.Getenv("SPEEK_PTT_SHORTCUT"); v != "" {
		acc = v
	}

	var recording atomic.Bool
	var cancel context.CancelFunc
	// Paste on by default; set SPEEK_SPIKE_PASTE=0 when testing from this terminal only.
	pasteEnabled := os.Getenv("SPEEK_SPIKE_PASTE") != "0"

	// The spike uses the production Recorder so no spike-only API remains in
	// the audio package. A throwaway config store supplies device defaults.
	spikeCfg, err := config.OpenDefault()
	if err != nil {
		log.Fatal(err)
	}
	rec := audio.NewRecorder(spikeCfg)
	capture := func(ctx context.Context) ([]int16, error) {
		if err := rec.Start(ctx); err != nil {
			return nil, err
		}
		<-ctx.Done()
		return rec.Stop()
	}

	mygo.App.SetName("speek capture spike")
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.OnWindowAllClosed(func() {})

	mygo.App.WhenReady(func() {
		tray, _ := mygo.NewTray(mygo.TrayOptions{ToolTip: "speek capture spike"})
		if tray != nil {
			tray.SetMenu(mygo.NewMenu([]*mygo.MenuItem{{Role: mygo.RoleQuit}}))
		}
	})

	listener := &recListener{
		onPress: func() {
			if recording.Swap(true) {
				return
			}
			ctx, c := context.WithCancel(context.Background())
			cancel = c
			go func() {
				samples, err := capture(ctx)
				recording.Store(false)
				if err != nil {
					log.Printf("capture: %v", err)
					return
				}
				ms := float64(len(samples)) / 16000 * 1000
				log.Printf("captured %d samples (~%.0f ms)", len(samples), ms)
				const text = "speek spike test"
				mygo.Clipboard.WriteText(text)
				log.Printf("clipboard set (%q)", text)
				if !pasteEnabled {
					log.Printf("paste skipped (SPEEK_SPIKE_PASTE=0)")
					return
				}
				if err := paste.PostPasteChord(); err != nil {
					log.Printf("paste: %v (check Accessibility for Cursor/Terminal)", err)
					return
				}
				log.Printf("Cmd+V posted at %s", time.Now().Format("15:04:05.000"))
			}()
		},
		onRelease: func() {
			if cancel != nil {
				cancel()
			}
		},
	}

	m := holdshortcut.New()
	log.Printf("speek spike-capture-paste: hold %s to record, release to stop.", acc)
	log.Printf("On release: clipboard + Cmd+V to the focused app. Click a text field elsewhere before you release.")
	if !pasteEnabled {
		log.Printf("Paste disabled (SPEEK_SPIKE_PASTE=0).")
	}

	go func() {
		runtime.LockOSThread()
		if err := m.Start(acc, listener); err != nil {
			log.Fatal(err)
		}
		if err := holdshortcut.Run(m); err != nil {
			log.Fatal(err)
		}
	}()

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

type recListener struct {
	onPress   func()
	onRelease func()
}

func (r *recListener) OnPress()   { r.onPress() }
func (r *recListener) OnRelease() { r.onRelease() }

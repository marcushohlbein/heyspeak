// Spike 1A: global push-to-talk press and release detection.
//
//	go run ./cmd/spike-ptt
//
// Hold Ctrl+Shift+Space (override with SPEEK_PTT_SHORTCUT). Requires Accessibility.
// Focus another app and watch stdout for press/release lines.
package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/speek/speek/internal/holdshortcut"
)

func main() {
	if runtime.GOOS != "darwin" {
		log.Fatal("spike-ptt: macOS only in Phase 1")
	}

	acc := holdshortcut.DefaultAccelerator
	if v := os.Getenv("SPEEK_PTT_SHORTCUT"); v != "" {
		acc = v
	}

	var holding atomic.Bool
	m := holdshortcut.New()
	listener := holdshortcutListener{holding: &holding}
	if err := m.Start(acc, listener); err != nil {
		log.Fatal(err)
	}

	log.Printf("speek spike-ptt: listening for %s", acc)
	log.Printf("PID %d — enable Accessibility for this app if press and release do not log.", os.Getpid())
	log.Printf("Hold the chord in any app. Quit with Ctrl+C.")

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = m.Stop()
		os.Exit(0)
	}()

	runtime.LockOSThread()
	if err := holdshortcut.Run(m); err != nil {
		log.Fatal(err)
	}
}

type holdshortcutListener struct {
	holding *atomic.Bool
}

func (l holdshortcutListener) OnPress() {
	if l.holding.Swap(true) {
		return
	}
	log.Printf("[%s] PRESS", time.Now().Format("15:04:05.000"))
}

func (l holdshortcutListener) OnRelease() {
	if !l.holding.Swap(false) {
		return
	}
	log.Printf("[%s] RELEASE", time.Now().Format("15:04:05.000"))
}

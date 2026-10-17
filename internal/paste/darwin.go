//go:build darwin

package paste

import (
	"context"
	"time"

	"github.com/speek/speek/internal/config"
)

type darwinService struct {
	cfg       *config.Store
	targetPID int
}

func newPlatformService(cfg *config.Store) Service {
	return &darwinService{cfg: cfg}
}

func (s *darwinService) CapturePasteTarget() {
	pid := frontmostPID()
	if pid <= 0 {
		// -1 marks a failed capture so Paste refuses an unverified target;
		// 0 (zero value) means no capture was attempted at all.
		pid = -1
	}
	s.targetPID = pid
}

// undoTranscriptClipboard puts the pre-paste clipboard back when it still holds text.
// A non-empty snapshot wins so images and files are not replaced by the transcript.
func undoTranscriptClipboard(text, prior string, priorItems int) {
	if readPlainClipboard() != text {
		return
	}
	if priorItems > 0 {
		restorePasteboardSnapshot()
		return
	}
	restorePlainClipboard(text, prior)
}

func (s *darwinService) Paste(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !accessibilityTrusted() {
		writePlainClipboard(text)
		return ErrPermissionDenied
	}

	priorItems := snapshotPasteboard()
	prior := readPlainClipboard()
	writePlainClipboard(text)

	if err := s.deliverPaste(); err != nil {
		// The target never received the transcript. The dialog shows it.
		undoTranscriptClipboard(text, prior, priorItems)
		return err
	}
	if s.cfg == nil || s.cfg.Get().RestoreClipboard {
		// Give the target app time to read the pasteboard before restoring
		// the prior clipboard; async readers (e.g. Electron) can otherwise
		// paste the restored content instead of the transcript.
		time.Sleep(250 * time.Millisecond)
		undoTranscriptClipboard(text, prior, priorItems)
	}
	return nil
}

func (s *darwinService) deliverPaste() error {
	front := frontmostPID()
	own := ownPID()
	activate, ok := resolveFocusForPaste(focusState{
		frontmost: front,
		own:       own,
		target:    s.targetPID,
	})
	if !ok {
		return ErrFocusChanged
	}
	if activate {
		if !activatePID(s.targetPID) {
			return ErrTargetLost
		}
		time.Sleep(50 * time.Millisecond)
		if frontmostPID() != s.targetPID {
			return ErrFocusChanged
		}
	}
	return PostPasteChord()
}

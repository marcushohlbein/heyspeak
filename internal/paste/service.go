package paste

import (
	"context"

	"github.com/speek/speek/internal/config"
)

// Service inserts transcript text into the application that was focused at push-to-talk.
type Service interface {
	// CapturePasteTarget records the frontmost application when dictation starts.
	CapturePasteTarget()
	Paste(ctx context.Context, text string) error
}

// Default returns the platform paste service.
func Default(cfg *config.Store) Service {
	return newPlatformService(cfg)
}

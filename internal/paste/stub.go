//go:build !darwin

package paste

import (
	"context"

	"github.com/speek/speek/internal/config"
)

type stubService struct{}

func newPlatformService(_ *config.Store) Service {
	return stubService{}
}

func (stubService) CapturePasteTarget() {}

func (stubService) Paste(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	writePlainClipboard(text)
	return ErrUnsupported
}

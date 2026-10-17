package audio

import (
	"encoding/hex"
	"fmt"

	"github.com/gen2brain/malgo"
)

// InputDevice is a capture endpoint.
type InputDevice struct {
	ID        string
	Name      string
	IsDefault bool
}

// ListInputDevices returns installed capture devices.
func ListInputDevices() ([]InputDevice, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, mapMalgoError(err)
	}
	defer ctx.Uninit()

	infos, err := ctx.Context.Devices(malgo.Capture)
	if err != nil {
		return nil, mapMalgoError(err)
	}
	out := make([]InputDevice, 0, len(infos))
	for _, d := range infos {
		out = append(out, InputDevice{
			ID:        d.ID.String(),
			Name:      d.Name(),
			IsDefault: d.IsDefault != 0,
		})
	}
	return out, nil
}

func resolveCaptureDeviceID(want string) (*malgo.DeviceID, error) {
	if want == "" {
		return nil, nil
	}
	id, err := parseDeviceID(want)
	if err != nil {
		return nil, err
	}
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, mapMalgoError(err)
	}
	defer ctx.Uninit()
	infos, err := ctx.Context.Devices(malgo.Capture)
	if err != nil {
		return nil, mapMalgoError(err)
	}
	for _, d := range infos {
		if d.ID.String() == id.String() {
			return &id, nil
		}
	}
	return nil, ErrDeviceNotFound
}

func parseDeviceID(s string) (malgo.DeviceID, error) {
	raw, err := hex.DecodeString(s)
	if err != nil {
		return malgo.DeviceID{}, fmt.Errorf("audio: invalid device id: %w", err)
	}
	var id malgo.DeviceID
	if len(raw) > len(id) {
		return malgo.DeviceID{}, fmt.Errorf("audio: device id too long")
	}
	copy(id[:], raw)
	return id, nil
}

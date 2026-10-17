package audio

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gen2brain/malgo"
)

const channels = 1

// SessionConfig configures one capture session.
type SessionConfig struct {
	DeviceID   string
	MaxSeconds int
	SampleRate uint32
}

// recordSession captures until ctx is cancelled or the sample cap is reached.
func recordSession(parent context.Context, cfg SessionConfig) ([]int16, error) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	if cfg.SampleRate == 0 {
		cfg.SampleRate = uint32(SampleRate)
	}

	ctxConfig, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, mapMalgoError(err)
	}
	defer ctxConfig.Uninit()

	devices, err := ctxConfig.Context.Devices(malgo.Capture)
	if err != nil {
		return nil, mapMalgoError(err)
	}
	if len(devices) == 0 {
		return nil, ErrNoInputDevice
	}

	deviceID, err := resolveCaptureDeviceID(cfg.DeviceID)
	if err != nil {
		return nil, err
	}

	openRate, openFormat := captureSetup(ctxConfig.Context, deviceID)
	maxSamples := 0
	if cfg.MaxSeconds > 0 {
		maxSamples = int(openRate) * cfg.MaxSeconds
	}
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.SampleRate = openRate
	// Match the hardware format. Converting the MacBook mic's float samples
	// to 16-bit in miniaudio delivered an all-zero buffer.
	deviceConfig.Capture.Format = openFormat
	deviceConfig.Capture.Channels = channels
	if deviceID != nil {
		deviceConfig.Capture.DeviceID = deviceID.Pointer()
	}

	var mu sync.Mutex
	var samples []int16
	var opened malgo.FormatType
	var openedChannels int
	onRecv := func(_, inputBuffer []byte, frameCount uint32) {
		if len(inputBuffer) == 0 || frameCount == 0 {
			return
		}
		frame := decodeCapture(inputBuffer, int(frameCount), openedChannels, opened)
		mu.Lock()
		var capped bool
		samples, capped = appendBounded(samples, frame, maxSamples)
		mu.Unlock()
		if capped {
			// Cap reached: stop the device and finish the take.
			cancel()
		}
	}

	device, err := malgo.InitDevice(ctxConfig.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: onRecv,
	})
	if err != nil && openRate != cfg.SampleRate {
		deviceConfig.SampleRate = cfg.SampleRate
		device, err = malgo.InitDevice(ctxConfig.Context, deviceConfig, malgo.DeviceCallbacks{
			Data: onRecv,
		})
	}
	if err != nil {
		return nil, mapMalgoError(fmt.Errorf("malgo device: %w", err))
	}
	defer device.Uninit()

	opened = device.CaptureFormat()
	openedChannels = int(device.CaptureChannels())
	if openedChannels <= 0 {
		openedChannels = 1
	}
	log.Printf("audio open: %d Hz (hardware %d Hz), %d ch (hardware %d), format %d (hardware %d)",
		device.SampleRate(), device.CaptureInternalSampleRate(),
		device.CaptureChannels(), device.CaptureInternalChannels(),
		device.CaptureFormat(), device.CaptureInternalFormat())

	if err := device.Start(); err != nil {
		return nil, mapMalgoError(err)
	}
	defer device.Stop()

	<-ctx.Done()

	mu.Lock()
	out := append([]int16(nil), samples...)
	mu.Unlock()

	delivered := device.SampleRate()
	if delivered == 0 {
		delivered = openRate
	}
	if delivered != cfg.SampleRate {
		out = ResampleLinear(out, int(delivered), int(cfg.SampleRate))
	}
	logCapture(out, delivered, cfg.SampleRate)
	if len(out) > int(cfg.SampleRate)/2 && peakAbs(out) == 0 {
		return nil, ErrSilentInput
	}
	return out, nil
}

func captureSetup(ctx malgo.Context, deviceID *malgo.DeviceID) (uint32, malgo.FormatType) {
	infos, err := ctx.Devices(malgo.Capture)
	if err != nil || len(infos) == 0 {
		return 48000, malgo.FormatF32
	}
	var chosen *malgo.DeviceInfo
	for i := range infos {
		if deviceID != nil && infos[i].ID == *deviceID {
			chosen = &infos[i]
			break
		}
		if deviceID == nil && infos[i].IsDefault != 0 {
			chosen = &infos[i]
		}
	}
	if chosen == nil {
		chosen = &infos[0]
	}
	log.Printf("audio device: %s default=%v", chosen.Name(), chosen.IsDefault != 0)
	return chooseCaptureRate(chosen.Formats), chooseCaptureFormat(chosen.Formats)
}

func logCapture(samples []int16, delivered, target uint32) {
	peak := peakAbs(samples)
	ms := 0.0
	if target > 0 {
		ms = float64(len(samples)) / float64(target) * 1000
	}
	log.Printf("audio: %.0f ms, %d samples, peak %d, %d Hz delivered as %d Hz", ms, len(samples), peak, delivered, target)
	if len(samples) > 0 && peak < 400 {
		log.Printf("audio: microphone level is very low; check the input device and permission")
	}
}

func peakAbs(samples []int16) int {
	var peak int
	for _, s := range samples {
		v := int(s)
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	return peak
}

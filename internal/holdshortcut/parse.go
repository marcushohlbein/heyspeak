package holdshortcut

import (
	"fmt"
	"runtime"
	"strings"
)

// chord describes a global shortcut as a main key plus required modifiers.
type chord struct {
	keyCode uint16
	ctrl    bool
	shift   bool
	alt     bool
	super   bool
}

func parseChord(accelerator string) (chord, error) {
	parts := strings.Split(accelerator, "+")
	var c chord
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if i == len(parts)-1 {
			code, ok := keyCodeFor(p)
			if !ok {
				return chord{}, fmt.Errorf("unknown key %q", p)
			}
			c.keyCode = code
			continue
		}
		switch strings.ToLower(p) {
		case "ctrl", "control":
			c.ctrl = true
		case "shift":
			c.shift = true
		case "alt", "option":
			c.alt = true
		case "cmd", "command", "super":
			c.super = true
		case "cmdorctrl":
			if runtime.GOOS == "darwin" {
				c.super = true
			} else {
				c.ctrl = true
			}
		default:
			return chord{}, fmt.Errorf("unknown modifier %q", p)
		}
	}
	if !c.ctrl && !c.shift && !c.alt && !c.super {
		return chord{}, fmt.Errorf("shortcut needs at least one modifier")
	}
	return c, nil
}

func keyCodeFor(key string) (uint16, bool) {
	if key == "+" {
		key = "="
	}
	if code, ok := macKeyCodes[key]; ok {
		return code, ok
	}
	if code, ok := macKeyCodes[strings.ToLower(key)]; ok {
		return code, ok
	}
	if key == "" {
		return 0, false
	}
	titled := strings.ToUpper(key[:1]) + strings.ToLower(key[1:])
	code, ok := macKeyCodes[titled]
	return code, ok
}

// mac virtual key codes (ANSI), aligned with MyGo's darwin hotkey table.
var macKeyCodes = map[string]uint16{
	"a": 0x00, "s": 0x01, "d": 0x02, "f": 0x03, "h": 0x04, "g": 0x05, "z": 0x06, "x": 0x07,
	"c": 0x08, "v": 0x09, "b": 0x0B, "q": 0x0C, "w": 0x0D, "e": 0x0E, "r": 0x0F, "y": 0x10,
	"t": 0x11, "1": 0x12, "2": 0x13, "3": 0x14, "4": 0x15, "6": 0x16, "5": 0x17, "=": 0x18,
	"9": 0x19, "7": 0x1A, "-": 0x1B, "8": 0x1C, "0": 0x1D, "]": 0x1E, "o": 0x1F, "u": 0x20,
	"[": 0x21, "i": 0x22, "p": 0x23, "l": 0x25, "j": 0x26, "'": 0x27, "k": 0x28, ";": 0x29,
	"\\": 0x2A, ",": 0x2B, "/": 0x2C, "n": 0x2D, "m": 0x2E, ".": 0x2F, "`": 0x32,
	"Enter": 0x24, "Tab": 0x30, "Space": 0x31, "Backspace": 0x33, "Escape": 0x35,
	"Delete": 0x75, "Home": 0x73, "End": 0x77, "PageUp": 0x74, "PageDown": 0x79,
	"Left": 0x7B, "Right": 0x7C, "Down": 0x7D, "Up": 0x7E,
	"F1": 0x7A, "F2": 0x78, "F3": 0x63, "F4": 0x76, "F5": 0x60, "F6": 0x61, "F7": 0x62, "F8": 0x64,
	"F9": 0x65, "F10": 0x6D, "F11": 0x67, "F12": 0x6F, "F13": 0x69, "F14": 0x6B, "F15": 0x71,
	"F16": 0x6A, "F17": 0x40, "F18": 0x4F, "F19": 0x50, "F20": 0x5A,
}

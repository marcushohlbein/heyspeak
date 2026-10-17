package settings

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/speek/speek/internal/holdshortcut"
)

func acceleratorFromKeyDown(ev ui.InputEvent) (string, bool) {
	if ev.Kind != ui.InputKeyDown || ev.Repeat {
		return "", false
	}
	key := ev.Key
	if key == ui.KeyUnknown {
		key = ev.PhysicalKey
	}
	if key == ui.KeyEscape && ev.Mods == 0 {
		return "", false
	}
	name := uiKeyName(key)
	if name == "" {
		return "", false
	}
	ctrl := ev.Mods&ui.Ctrl != 0
	shift := ev.Mods&ui.Shift != 0
	alt := ev.Mods&ui.Alt != 0
	meta := ev.Mods&ui.Super != 0
	if !ctrl && !shift && !alt && !meta {
		return "", false
	}
	acc := holdshortcut.FormatAccelerator(ctrl, shift, alt, meta, name)
	if acc == "" {
		return "", false
	}
	return acc, true
}

func uiKeyName(k ui.Key) string {
	switch {
	case k >= ui.KeyA && k <= ui.KeyZ:
		return string(rune('a' + k - ui.KeyA))
	case k >= ui.Key0 && k <= ui.Key9:
		return string(rune('0' + k - ui.Key0))
	case k >= ui.KeyF1 && k <= ui.KeyF12:
		return "F" + strconv.Itoa(int(k-ui.KeyF1)+1)
	}
	switch k {
	case ui.KeyEnter:
		return "Enter"
	case ui.KeyEscape:
		return "Escape"
	case ui.KeyBackspace:
		return "Backspace"
	case ui.KeyTab:
		return "Tab"
	case ui.KeySpace:
		return "Space"
	case ui.KeyDelete:
		return "Delete"
	case ui.KeyInsert:
		return "Insert"
	case ui.KeyHome:
		return "Home"
	case ui.KeyEnd:
		return "End"
	case ui.KeyPageUp:
		return "PageUp"
	case ui.KeyPageDown:
		return "PageDown"
	case ui.KeyLeft:
		return "Left"
	case ui.KeyRight:
		return "Right"
	case ui.KeyUp:
		return "Up"
	case ui.KeyDown:
		return "Down"
	case ui.KeyMinus:
		return "-"
	case ui.KeyEqual:
		return "="
	case ui.KeyComma:
		return ","
	case ui.KeyPeriod:
		return "."
	case ui.KeySlash:
		return "/"
	case ui.KeySemicolon:
		return ";"
	case ui.KeyQuote:
		return "'"
	case ui.KeyBracketLeft:
		return "["
	case ui.KeyBracketRight:
		return "]"
	case ui.KeyBackslash:
		return "\\"
	case ui.KeyBackquote:
		return "`"
	default:
		return ""
	}
}

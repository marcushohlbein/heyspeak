package settings

import "github.com/egoist/mygo/ui"

// Lucide icons (https://lucide.dev, ISC license), stroked so Icon tints
// them like text.
func iconSVG(body string) []byte {
	return []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` +
		body + `</svg>`)
}

var (
	iconSliders = ui.MustParseSVG(iconSVG(
		`<line x1="21" x2="14" y1="4" y2="4"/><line x1="10" x2="3" y1="4" y2="4"/>` +
			`<line x1="21" x2="12" y1="12" y2="12"/><line x1="8" x2="3" y1="12" y2="12"/>` +
			`<line x1="21" x2="16" y1="20" y2="20"/><line x1="12" x2="3" y1="20" y2="20"/>` +
			`<line x1="14" x2="14" y1="2" y2="6"/><line x1="8" x2="8" y1="10" y2="14"/>` +
			`<line x1="16" x2="16" y1="18" y2="22"/>`))
	iconMic = ui.MustParseSVG(iconSVG(
		`<path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/>` +
			`<path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" x2="12" y1="19" y2="22"/>`))
	iconPackage = ui.MustParseSVG(iconSVG(
		`<path d="m7.5 4.27 9 5.15"/>` +
			`<path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/>` +
			`<path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/>`))
	iconShield = ui.MustParseSVG(iconSVG(
		`<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/>`))
	iconWrench = ui.MustParseSVG(iconSVG(
		`<path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>`))
	iconX = ui.MustParseSVG(iconSVG(
		`<path d="M18 6 6 18"/><path d="m6 6 12 12"/>`))
)

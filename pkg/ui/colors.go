package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Neon synthwave color palette — inspired by ColorHunt neon palettes.
// Uses hex colors for precise, vibrant neon tones.
var (
	// Primary accent: active borders, selected items, playing status
	// Bright neon purple — the synthwave signature
	colorAccent = compat.AdaptiveColor{Light: lipgloss.Color("#7700FF"), Dark: lipgloss.Color("#B44AFF")}

	// Secondary accent: app title, panel titles, station name
	// Hot neon pink
	colorTitle = compat.AdaptiveColor{Light: lipgloss.Color("#FF0080"), Dark: lipgloss.Color("#FF3399")}

	// Muted text: help text, placeholders, inactive labels
	// Plain gray — stays out of the way
	colorMuted = compat.AdaptiveColor{Light: lipgloss.Color("#666666"), Dark: lipgloss.Color("#999999")}

	// Inactive borders
	// Quiet, just structure
	colorBorder = compat.AdaptiveColor{Light: lipgloss.Color("#BBBBBB"), Dark: lipgloss.Color("#555555")}

	// Unselected list items (autocomplete suggestions)
	colorDim = compat.AdaptiveColor{Light: lipgloss.Color("#555555"), Dark: lipgloss.Color("#BBBBBB")}

	// Warning/status: paused indicator, sleep timer
	// Neon orange — pops against purple and pink
	colorWarning = compat.AdaptiveColor{Light: lipgloss.Color("#EE4400"), Dark: lipgloss.Color("#FF6622")}

	// Labels: filter names, field labels — readable body text
	colorLabel = compat.AdaptiveColor{Light: lipgloss.Color("#222222"), Dark: lipgloss.Color("#DDDDDD")}
)

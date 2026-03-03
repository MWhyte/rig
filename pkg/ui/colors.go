package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Clean minimal palette — one blue accent, otherwise white and grays.
var (
	// Primary accent: active borders, selected items, playing status
	colorAccent = compat.AdaptiveColor{Light: lipgloss.Color("#2E6EB0"), Dark: lipgloss.Color("#5E9EDB")}

	// Titles: app title, panel titles, station name
	colorTitle = compat.AdaptiveColor{Light: lipgloss.Color("#1A1A1A"), Dark: lipgloss.Color("#E8E8E8")}

	// Muted text: help text, placeholders, inactive labels
	colorMuted = compat.AdaptiveColor{Light: lipgloss.Color("#767676"), Dark: lipgloss.Color("#909090")}

	// Inactive borders
	colorBorder = compat.AdaptiveColor{Light: lipgloss.Color("#AAAAAA"), Dark: lipgloss.Color("#444444")}

	// Dim text: unselected list items, autocomplete suggestions
	colorDim = compat.AdaptiveColor{Light: lipgloss.Color("#999999"), Dark: lipgloss.Color("#666666")}

	// Warning/status: paused indicator, sleep timer
	colorWarning = compat.AdaptiveColor{Light: lipgloss.Color("#C07020"), Dark: lipgloss.Color("#E8A838")}

	// Labels: filter names, field labels
	colorLabel = compat.AdaptiveColor{Light: lipgloss.Color("#333333"), Dark: lipgloss.Color("#C8C8C8")}
)

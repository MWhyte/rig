package ui

import "github.com/charmbracelet/lipgloss"

// Neon synthwave color palette — inspired by ColorHunt neon palettes.
// Uses hex colors for precise, vibrant neon tones.
var (
	// Primary accent: active borders, selected items, playing status
	// Bright neon purple — the synthwave signature
	colorAccent = lipgloss.AdaptiveColor{Light: "#7700FF", Dark: "#B44AFF"}

	// Secondary accent: app title, panel titles, station name
	// Hot neon pink
	colorTitle = lipgloss.AdaptiveColor{Light: "#FF0080", Dark: "#FF3399"}

	// Muted text: help text, placeholders, inactive labels
	// Plain gray — stays out of the way
	colorMuted = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}

	// Inactive borders
	// Quiet, just structure
	colorBorder = lipgloss.AdaptiveColor{Light: "#BBBBBB", Dark: "#555555"}

	// Unselected list items (autocomplete suggestions)
	colorDim = lipgloss.AdaptiveColor{Light: "#555555", Dark: "#BBBBBB"}

	// Warning/status: paused indicator, sleep timer
	// Neon orange — pops against purple and pink
	colorWarning = lipgloss.AdaptiveColor{Light: "#EE4400", Dark: "#FF6622"}

	// Labels: filter names, field labels — readable body text
	colorLabel = lipgloss.AdaptiveColor{Light: "#222222", Dark: "#DDDDDD"}
)

package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Theme defines the full color palette for the UI.
type Theme struct {
	Name    string
	Accent  compat.AdaptiveColor // active borders, selected items, playing status
	Title   compat.AdaptiveColor // titles, station name
	Muted   compat.AdaptiveColor // secondary text, placeholders
	Border  compat.AdaptiveColor // inactive borders
	Dim     compat.AdaptiveColor // very faded text, tech info
	Warning compat.AdaptiveColor // paused indicator, warnings
	Label   compat.AdaptiveColor // filter labels
}

func adaptive(light, dark string) compat.AdaptiveColor {
	return compat.AdaptiveColor{Light: lipgloss.Color(light), Dark: lipgloss.Color(dark)}
}

var themes = []Theme{
	{
		Name:    "Classic",
		Accent:  adaptive("#2E6EB0", "#5E9EDB"),
		Title:   adaptive("#1A1A1A", "#E8E8E8"),
		Muted:   adaptive("#767676", "#909090"),
		Border:  adaptive("#AAAAAA", "#444444"),
		Dim:     adaptive("#999999", "#666666"),
		Warning: adaptive("#C07020", "#E8A838"),
		Label:   adaptive("#333333", "#C8C8C8"),
	},
	{
		Name:    "Nord",
		Accent:  adaptive("#5E81AC", "#88C0D0"),
		Title:   adaptive("#2E3440", "#ECEFF4"),
		Muted:   adaptive("#616E88", "#7B88A1"),
		Border:  adaptive("#D8DEE9", "#3B4252"),
		Dim:     adaptive("#ADB7C9", "#4C566A"),
		Warning: adaptive("#BF616A", "#EBCB8B"),
		Label:   adaptive("#3B4252", "#D8DEE9"),
	},
	{
		Name:    "Catppuccin",
		Accent:  adaptive("#1E66F5", "#89B4FA"),
		Title:   adaptive("#4C4F69", "#CDD6F4"),
		Muted:   adaptive("#6C6F85", "#9399B2"),
		Border:  adaptive("#ACB0BE", "#313244"),
		Dim:     adaptive("#8C8FA1", "#45475A"),
		Warning: adaptive("#FE640B", "#FAB387"),
		Label:   adaptive("#5C5F77", "#BAC2DE"),
	},
	{
		Name:    "Gruvbox",
		Accent:  adaptive("#458588", "#83A598"),
		Title:   adaptive("#282828", "#EBDBB2"),
		Muted:   adaptive("#928374", "#A89984"),
		Border:  adaptive("#D5C4A1", "#504945"),
		Dim:     adaptive("#BDAE93", "#665C54"),
		Warning: adaptive("#D65D0E", "#FE8019"),
		Label:   adaptive("#3C3836", "#D5C4A1"),
	},
	{
		Name:    "Rose Pine",
		Accent:  adaptive("#286983", "#9CCFD8"),
		Title:   adaptive("#191724", "#E0DEF4"),
		Muted:   adaptive("#6E6A86", "#908CAA"),
		Border:  adaptive("#D9D7E0", "#26233A"),
		Dim:     adaptive("#C4C1CE", "#393552"),
		Warning: adaptive("#B4637A", "#EBBCBA"),
		Label:   adaptive("#26233A", "#E0DEF4"),
	},
	{
		Name:    "Neon",
		Accent:  adaptive("#C026D3", "#FF2D9F"),
		Title:   adaptive("#18181B", "#FFFFFF"),
		Muted:   adaptive("#7C3AED", "#BD93F9"),
		Border:  adaptive("#6D28D9", "#3D1A78"),
		Dim:     adaptive("#8B5CF6", "#555580"),
		Warning: adaptive("#F59E0B", "#00E5FF"),
		Label:   adaptive("#4C1D95", "#E0AAFF"),
	},
}

var themeIndex = 0

// Active color vars — reassigned on theme switch.
var (
	colorAccent  = themes[0].Accent
	colorTitle   = themes[0].Title
	colorMuted   = themes[0].Muted
	colorBorder  = themes[0].Border
	colorDim     = themes[0].Dim
	colorWarning = themes[0].Warning
	colorLabel   = themes[0].Label
)

// setTheme applies the theme at the given index.
func setTheme(index int) {
	themeIndex = index
	t := themes[index]
	colorAccent = t.Accent
	colorTitle = t.Title
	colorMuted = t.Muted
	colorBorder = t.Border
	colorDim = t.Dim
	colorWarning = t.Warning
	colorLabel = t.Label
}

// themeIndexByName returns the index of a theme by name, or 0 if not found.
func themeIndexByName(name string) int {
	for i, t := range themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

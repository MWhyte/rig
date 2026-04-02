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
		Name:    "Monokai",
		Accent:  adaptive("#E05C1A", "#FD971F"),
		Title:   adaptive("#272822", "#F8F8F2"),
		Muted:   adaptive("#75715E", "#90876A"),
		Border:  adaptive("#ADA695", "#49483E"),
		Dim:     adaptive("#8F8777", "#3E3D32"),
		Warning: adaptive("#AA9900", "#E6DB74"),
		Label:   adaptive("#49483E", "#CFCFC2"),
	},
	{
		Name:    "Crimson",
		Accent:  adaptive("#A50000", "#FF3333"),
		Title:   adaptive("#1A0000", "#FFF0F0"),
		Muted:   adaptive("#7A3030", "#CC8888"),
		Border:  adaptive("#CC8080", "#5A1A1A"),
		Dim:     adaptive("#B08080", "#3A1010"),
		Warning: adaptive("#C05A00", "#FF8C00"),
		Label:   adaptive("#3A0000", "#FFCCCC"),
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
		Name:    "Hacker",
		Accent:  adaptive("#1A6B1A", "#00FF41"),
		Title:   adaptive("#001A00", "#CCFFCC"),
		Muted:   adaptive("#2D5A1B", "#44AA44"),
		Border:  adaptive("#1A4A1A", "#003300"),
		Dim:     adaptive("#0D2B0D", "#1A4A1A"),
		Warning: adaptive("#6B6B00", "#CCCC00"),
		Label:   adaptive("#004400", "#66CC66"),
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
	for i := range themes {
		if themes[i].Name == name {
			return i
		}
	}
	return 0
}

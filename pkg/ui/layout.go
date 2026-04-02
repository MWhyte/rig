package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func activeBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent)
}

func inactiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorBorder)
}

func panelTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(colorTitle).Padding(0, 1)
}

func activePanelTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Padding(0, 1)
}

// renderMultiPanelLayout renders the main multi-panel layout.
func (m *Model) renderMultiPanelLayout() string {
	// Calculate dimensions
	// Left column (70% width): Filters + Station List
	// Right column (30% width): Player + Sponsors
	leftWidth := int(float64(m.width) * 0.70)
	rightWidth := m.width - leftWidth

	// Reserve space for header and footer
	headerHeight := lipgloss.Height(m.renderHeader())
	footerHeight := lipgloss.Height(m.renderFooter())
	chrome := headerHeight + footerHeight

	// topHeight must be at least 11: border(2) + title(1) + blank(1) + 5 filters(5) + blank(1) + help(1)
	// to prevent the filters panel from overflowing and misaligning the station list.
	topHeight := max(11, int(float64(m.height)*0.30))
	bottomHeight := m.height - topHeight - chrome

	// Build left column: Filters on top, Station List below
	filtersPanel := m.renderFiltersPanel(leftWidth-3, topHeight-2)
	stationListPanel := m.renderStationListPanel(leftWidth-3, bottomHeight)
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, filtersPanel, stationListPanel)

	// Build right column: Sponsors on top, Player below
	sponsorsPanel := m.renderSponsorsPanel(rightWidth-3, topHeight-2)
	playerPanel := m.renderPlayerPanel(rightWidth-3, bottomHeight)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, sponsorsPanel, playerPanel)

	// Combine columns horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Add header and footer
	header := m.renderHeader()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		footer,
	)
}

// renderHeader renders the app header.
func (m *Model) renderHeader() string {
	title := titleStyle.Render(" rig.fm - Terminal Radio")
	return title
}

// renderFooter renders the help footer.
func (m *Model) renderFooter() string {
	// If editing a filter, show different help
	if m.editingFilter != FilterNone {
		help := "Type to edit filter • enter: apply • esc: cancel"
		return "\n" + helpStyle.Render(help)
	}

	var shortcuts string

	switch m.focusedSection {
	case SectionStationList:
		shortcuts = "↑↓/jk: navigate • enter/space: play • f: toggle fav"
	case SectionFilters:
		shortcuts = "↑↓/jk: select • enter: edit • c: clear"
	}

	help := fmt.Sprintf("tab: switch sections [%s] • %s • space: pause • +/-: volume • t: sleep timer • ctrl+t: theme • ctrl+c: quit",
		m.focusedSection.String(),
		shortcuts,
	)

	return "\n" + helpStyle.Render(help)
}

// renderStationListPanel renders the station list panel.
func (m *Model) renderStationListPanel(width, height int) string {
	// width-2 for border side chars, height-3 for border top/bottom + title line
	m.stationList.SetSize(width-2, height-3)

	// Get border style based on focus
	borderStyle := inactiveBorderStyle()
	titleStyle := panelTitleStyle()
	if m.focusedSection == SectionStationList {
		borderStyle = activeBorderStyle()
		titleStyle = activePanelTitleStyle()
	}

	// Build content
	title := titleStyle.Render(fmt.Sprintf("Stations (%d)", len(m.stations)))
	content := m.stationList.View()

	panel := lipgloss.JoinVertical(lipgloss.Left, title, content)

	// Truncate to inner height (height minus border) so the panel never overflows
	panel = truncateLines(panel, height-2)

	return borderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

// waveFrames is a looping ASCII waveform animation (8 chars wide).
var waveFrames = []string{
	"▁▂▃▂▁▁▂▃",
	"▂▃▂▁▁▂▃▂",
	"▃▂▁▁▂▃▂▁",
	"▂▁▁▂▃▂▁▂",
	"▁▁▂▃▂▁▂▃",
	"▁▂▃▂▁▂▃▂",
	"▂▃▂▁▂▃▂▁",
	"▃▂▁▂▃▂▁▁",
}

// truncateLines limits s to at most n lines.
func truncateLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

// truncate cuts s to maxLen, appending "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// renderPlayerPanel renders the player panel.
func (m *Model) renderPlayerPanel(width, height int) string {
	borderStyle := inactiveBorderStyle()
	titleStyle := panelTitleStyle()
	title := titleStyle.Render("Player")

	maxLen := width - 4
	if maxLen < 10 {
		maxLen = 10
	}

	if m.playing == nil {
		content := "\n " +
			lipgloss.NewStyle().Foreground(colorMuted).Render("No station playing") +
			"\n\n " +
			lipgloss.NewStyle().Foreground(colorDim).Render("Select a station and press Enter")

		panel := lipgloss.JoinVertical(lipgloss.Left, title, content)
		return borderStyle.Width(width).Height(height).Render(panel)
	}

	var info strings.Builder

	// Station name
	info.WriteString("\n ")
	info.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTitle).
		Render(truncate(m.playing.Name, maxLen)))
	info.WriteString("\n ")

	// Current song
	if m.currentSong != "" {
		info.WriteString(lipgloss.NewStyle().Foreground(colorAccent).
			Render(truncate(m.currentSong, maxLen)))
	} else {
		info.WriteString(lipgloss.NewStyle().Foreground(colorMuted).
			Render("No song info"))
	}
	info.WriteString("\n\n ")

	// Country · Genre
	location := m.playing.Country
	if m.currentGenre != "" {
		location += " · " + m.currentGenre
	}
	info.WriteString(lipgloss.NewStyle().Foreground(colorMuted).
		Render(truncate(location, maxLen)))
	info.WriteString("\n\n ")

	// Playback status + wave animation (with optional sleep timer to the right)
	if m.isPlaying {
		wave := waveFrames[m.waveFrame%len(waveFrames)]
		info.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render("▶") +
			"  " +
			lipgloss.NewStyle().Foreground(colorMuted).Render(wave))
	} else {
		info.WriteString(lipgloss.NewStyle().Foreground(colorWarning).
			Render("⏸  Paused"))
	}
	if m.sleepTimerActive && m.sleepTimerRemaining > 0 {
		minutes := int(m.sleepTimerRemaining.Minutes())
		seconds := int(m.sleepTimerRemaining.Seconds()) % 60
		timerText := fmt.Sprintf("  ⏱ %d:%02d", minutes, seconds)
		if minutes >= 60 {
			hours := minutes / 60
			minutes %= 60
			timerText = fmt.Sprintf("  ⏱ %d:%02d:%02d", hours, minutes, seconds)
		}
		info.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render(timerText))
	}
	info.WriteString("\n\n ")

	// Animated volume bar
	vol, _ := m.player.GetVolume()
	label := lipgloss.NewStyle().Foreground(colorDim).Render("vol ")
	pct := lipgloss.NewStyle().Foreground(colorMuted).Render(fmt.Sprintf(" %d%%", vol))
	barWidth := maxLen - 4 - 5 // "vol " (4) + " 75%" (4-5)
	if barWidth < 4 {
		barWidth = 4
	}
	m.volumeBar.SetWidth(barWidth)
	info.WriteString(label + m.volumeBar.View() + pct)
	info.WriteString("\n ")

	// Separator
	info.WriteString(lipgloss.NewStyle().Foreground(colorBorder).
		Render(strings.Repeat("─", maxLen)))
	info.WriteString("\n ")

	// Tech info — muted, minimal
	techInfo := m.playing.Codec
	if m.playing.Bitrate > 0 {
		techInfo += fmt.Sprintf(" · %d kbps", m.playing.Bitrate)
	}
	if m.actualKbps > 0 && int(m.actualKbps) != m.playing.Bitrate {
		techInfo += fmt.Sprintf(" (actual: %d)", int(m.actualKbps))
	}
	info.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render(techInfo))

	panel := lipgloss.JoinVertical(lipgloss.Left, title, info.String())
	return borderStyle.Width(width).Height(height).Render(panel)
}

// renderFiltersPanel renders the filters panel.
func (m *Model) renderFiltersPanel(width, height int) string {
	borderStyle := inactiveBorderStyle()
	titleStyle := panelTitleStyle()
	if m.focusedSection == SectionFilters {
		borderStyle = activeBorderStyle()
		titleStyle = activePanelTitleStyle()
	}

	title := titleStyle.Render("Filters")

	var content string

	// MODE 1: Editing - show autocomplete interface
	if m.editingFilter != FilterNone {
		content = m.autocomplete.View(width-4, height-4)
	} else {
		// MODE 2: Normal - show 4 filter options
		content = m.renderFilterList()
	}

	panel := lipgloss.JoinVertical(lipgloss.Left, title, content)

	return borderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

// renderFilterList renders the normal filter list view.
func (m *Model) renderFilterList() string {
	var content strings.Builder

	content.WriteString("\n")

	// Define filter items
	filters := []struct {
		index   int
		label   string
		value   string
		isEmpty bool
	}{
		{0, "Country", m.filters.Country, m.filters.Country == ""},
		{1, "Genre", m.filters.Genre, m.filters.Genre == ""},
		{2, "Language", m.filters.Language, m.filters.Language == ""},
		{3, "Station", m.filters.StationName, m.filters.StationName == ""},
		{4, "Favorites", "", false}, // Special case for favorites
	}

	for _, filter := range filters {
		// Determine if this filter is selected
		isSelected := (m.selectedFilterIndex == filter.index)

		// Build prefix (arrow for selected item)
		prefix := "  "
		if isSelected {
			prefix = "→ "
		}

		// Build value text and style
		var valueText string
		var valueStyle lipgloss.Style

		if filter.index == 4 {
			// Special handling for favorites
			if m.filters.FavoritesOnly {
				valueText = "★ Only"
				valueStyle = lipgloss.NewStyle().Foreground(colorAccent)
			} else {
				valueText = "All Stations"
				valueStyle = lipgloss.NewStyle().Foreground(colorMuted)
			}
		} else {
			// Regular filters
			if filter.isEmpty {
				valueText = "All " + filter.label + "s"
				valueStyle = lipgloss.NewStyle().Foreground(colorMuted)
			} else {
				valueText = filter.value
				valueStyle = lipgloss.NewStyle().Foreground(colorAccent)
			}
		}

		// Apply selection highlighting
		if isSelected {
			valueStyle = valueStyle.Bold(true).Foreground(colorAccent)
		}

		// Format line
		labelStyle := lipgloss.NewStyle().Foreground(colorLabel)
		fmt.Fprintf(&content, "%s%s %s\n",
			prefix,
			labelStyle.Render(fmt.Sprintf("%d. %s:", filter.index+1, filter.label)),
			valueStyle.Render(valueText))
	}

	content.WriteString("\n")

	// Help text
	if m.focusedSection == SectionFilters {
		content.WriteString(lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("  ↑↓/jk: select • enter: edit • 1-5: direct • c: clear"))
	} else {
		content.WriteString(lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("  Press Tab to focus"))
	}

	return content.String()
}

// renderThemeModal renders the theme picker modal with live preview.
func (m *Model) renderThemeModal() string {
	var content strings.Builder
	content.WriteString("\n")

	for i := range themes {
		if i == m.themeModalIndex {
			content.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Bold(true).
				Render(fmt.Sprintf("  → %s", themes[i].Name)))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(colorMuted).
				Render(fmt.Sprintf("    %s", themes[i].Name)))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Foreground(colorDim).
		Render("  ↑↓/jk: navigate • enter: select • esc: cancel"))

	title := lipgloss.NewStyle().Bold(true).Foreground(colorTitle).Padding(0, 1).Render("Theme")
	panel := lipgloss.JoinVertical(lipgloss.Left, title, content.String())

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Width(30).
		Render(panel)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

// renderTimerModal renders the sleep timer configuration modal.
func (m *Model) renderTimerModal() string {
	// Modal dimensions
	modalWidth := 50
	modalHeight := 10

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorTitle).
		Padding(0, 1).
		Render("⏱ Sleep Timer")

	var content strings.Builder
	content.WriteString("\n")

	// Show current timer status
	if m.sleepTimerActive {
		minutes := int(m.sleepTimerRemaining.Minutes())
		seconds := int(m.sleepTimerRemaining.Seconds()) % 60
		statusText := fmt.Sprintf("Current timer: %d:%02d remaining", minutes, seconds)
		content.WriteString("  ")
		content.WriteString(lipgloss.NewStyle().
			Foreground(colorAccent).
			Render(statusText))
		content.WriteString("\n\n")
	} else {
		content.WriteString("  ")
		content.WriteString(lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("No timer active"))
		content.WriteString("\n\n")
	}

	// Input field
	content.WriteString("  Set timer (minutes): ")
	content.WriteString(m.timerInput.View())
	content.WriteString("\n\n")

	// Help text
	helpText := "enter: start timer"
	if m.sleepTimerActive {
		helpText += " • x: cancel timer"
	}
	helpText += " • esc: close"

	content.WriteString("  ")
	content.WriteString(lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(helpText))

	// Create modal panel
	panel := lipgloss.JoinVertical(lipgloss.Left, title, content.String())

	// Style the modal with border
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight).
		Render(panel)

	// Center the modal
	centered := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)

	return centered
}

// renderSponsorsPanel renders the sponsors panel with a vertical scrolling list.
func (m *Model) renderSponsorsPanel(width, height int) string {
	title := panelTitleStyle().Render("♥ Sponsors")

	if len(m.liveSponsors) == 0 {
		fallback := lipgloss.NewStyle().
			Foreground(colorDim).
			Render("\n  Sponsor rig.fm: github.com/sponsors/MWhyte")
		panel := lipgloss.JoinVertical(lipgloss.Left, title, fallback)
		return inactiveBorderStyle().Width(width).Height(height).Render(panel)
	}

	// Build virtual item list: [name, dot, name, dot, ...]
	virtual := make([]string, len(m.liveSponsors)*2)
	for i, s := range m.liveSponsors {
		display := s.Name
		if display == "" {
			display = s.Login
		}
		virtual[i*2] = display
		virtual[i*2+1] = "·"
	}

	// Window height = inner panel height minus border (2) and title (1)
	windowHeight := height - 3
	if windowHeight < 1 {
		windowHeight = 1
	}

	nameStyle := lipgloss.NewStyle().Foreground(colorMuted)
	dotStyle := lipgloss.NewStyle().Foreground(colorDim)

	var lines []string
	n := len(virtual)
	for i := 0; i < windowHeight; i++ {
		idx := (m.sponsorScrollOffset + i) % n
		item := virtual[idx]
		if idx%2 == 0 {
			lines = append(lines, "  "+nameStyle.Render(item))
		} else {
			lines = append(lines, "  "+dotStyle.Render(item))
		}
	}

	content := strings.Join(lines, "\n")
	panel := lipgloss.JoinVertical(lipgloss.Left, title, content)
	return inactiveBorderStyle().Width(width).Height(height).Render(panel)
}

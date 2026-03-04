package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	// Border styles
	activeBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent)

	inactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder)

	// Panel styles
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle).
			Padding(0, 1)

	activePanelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				Padding(0, 1)
)

// renderMultiPanelLayout renders the main multi-panel layout
func (m *Model) renderMultiPanelLayout() string {
	// Calculate dimensions
	// Left column (70% width): Filters + Station List
	// Right column (30% width): Player + Sponsors
	leftWidth := int(float64(m.width) * 0.70)
	rightWidth := m.width - leftWidth

	// topHeight must be at least 11: border(2) + title(1) + blank(1) + 5 filters(5) + blank(1) + help(1)
	// to prevent the filters panel from overflowing and misaligning the station list.
	topHeight := max(11, int(float64(m.height)*0.30))
	bottomHeight := m.height - topHeight - 8

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

// renderHeader renders the app header
func (m *Model) renderHeader() string {
	title := titleStyle.Render(" rig.fm - Terminal Radio")
	return title
}

// renderFooter renders the help footer
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

	help := fmt.Sprintf("tab: switch sections [%s] • %s • space: pause • +/-: volume • t: sleep timer • ctrl+c: quit",
		m.focusedSection.String(),
		shortcuts,
	)

	return "\n" + helpStyle.Render(help)
}

// renderStationListPanel renders the station list panel
func (m *Model) renderStationListPanel(width, height int) string {
	// Set list size
	m.stationList.SetSize(width, height-2)

	// Get border style based on focus
	borderStyle := inactiveBorderStyle
	titleStyle := panelTitleStyle
	if m.focusedSection == SectionStationList {
		borderStyle = activeBorderStyle
		titleStyle = activePanelTitleStyle
	}

	// Build content
	title := titleStyle.Render(fmt.Sprintf("Stations (%d)", len(m.stations)))
	content := m.stationList.View()

	panel := lipgloss.JoinVertical(lipgloss.Left, title, content)

	return borderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

// waveFrames is a looping ASCII waveform animation (8 chars wide)
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

// truncate cuts s to maxLen, appending "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// renderPlayerPanel renders the player panel
func (m *Model) renderPlayerPanel(width, height int) string {
	borderStyle := inactiveBorderStyle
	titleStyle := panelTitleStyle
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
			minutes = minutes % 60
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

// renderFiltersPanel renders the filters panel
func (m *Model) renderFiltersPanel(width, height int) string {
	borderStyle := inactiveBorderStyle
	titleStyle := panelTitleStyle
	if m.focusedSection == SectionFilters {
		borderStyle = activeBorderStyle
		titleStyle = activePanelTitleStyle
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

// renderFilterList renders the normal filter list view
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
		content.WriteString(fmt.Sprintf("%s%s %s\n",
			prefix,
			labelStyle.Render(fmt.Sprintf("%d. %s:", filter.index+1, filter.label)),
			valueStyle.Render(valueText)))
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

// renderTimerModal renders the sleep timer configuration modal
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

// renderSponsorsPanel renders the sponsors/ads panel with wipe animation
func (m *Model) renderSponsorsPanel(width, height int) string {
	title := panelTitleStyle.Render("Sponsors")

	if len(m.sponsorAds) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("\n  Your ad here")
		panel := lipgloss.JoinVertical(lipgloss.Left, title, placeholder)
		return inactiveBorderStyle.
			Width(width).
			Height(height).
			Render(panel)
	}

	currentLines := strings.Split(m.sponsorAds[m.sponsorIndex], "\n")

	var displayLines []string

	switch m.sponsorWipePhase {
	case wipeOut:
		prevLines := strings.Split(m.sponsorAds[m.sponsorPrevIndex], "\n")
		for i, line := range prevLines {
			if i < m.sponsorFrame {
				displayLines = append(displayLines, "")
			} else {
				displayLines = append(displayLines, line)
			}
		}
	case wipePause:
		// Show blank
		displayLines = []string{""}
	case wipeIn:
		for i, line := range currentLines {
			if i < m.sponsorFrame {
				displayLines = append(displayLines, line)
			} else {
				displayLines = append(displayLines, "")
			}
		}
	default:
		displayLines = currentLines
	}

	// Clamp to available inner height: border(2) + title(1) = 3 reserved rows
	maxLines := height - 3
	if maxLines < 1 {
		maxLines = 1
	}
	if len(displayLines) > maxLines {
		displayLines = displayLines[:maxLines]
	}

	content := strings.Join(displayLines, "\n")

	adStyle := lipgloss.NewStyle().Foreground(colorDim)
	panel := lipgloss.JoinVertical(lipgloss.Left, title, adStyle.Render(content))

	return inactiveBorderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Border styles
	activeBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86"))

	inactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))

	// Panel styles
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	activePanelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Padding(0, 1)
)

// renderMultiPanelLayout renders the main multi-panel layout
func (m *Model) renderMultiPanelLayout() string {
	// Calculate dimensions
	leftWidth := int(float64(m.width) * 0.6)
	rightWidth := m.width - leftWidth - 4

	topRightHeight := 12
	bottomRightHeight := m.height - topRightHeight - 8

	stationListHeight := m.height - 8

	// Build panels
	stationListPanel := m.renderStationListPanel(leftWidth-2, stationListHeight)
	nowPlayingPanel := m.renderNowPlayingPanel(rightWidth-2, topRightHeight-2)
	filtersPanel := m.renderFiltersPanel(rightWidth-2, bottomRightHeight-2)

	// Combine right panels vertically
	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		nowPlayingPanel,
		filtersPanel,
	)

	// Combine left and right columns horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stationListPanel,
		rightColumn,
	)

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
	title := titleStyle.Render("  rig.fm - Terminal Radio")
	return title + "\n"
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
		shortcuts = "↑↓/jk: navigate • enter: play • /: filter"
	case SectionFilters:
		shortcuts = "1-3: edit filter • c: clear"
	}

	help := fmt.Sprintf("tab: switch sections [%s] • %s • space: pause • +/-: volume • ?: help • q: quit",
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
	title := titleStyle.Render(fmt.Sprintf(" Stations (%d) ", len(m.stations)))
	content := m.stationList.View()

	panel := lipgloss.JoinVertical(lipgloss.Left, title, content)

	return borderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

// renderNowPlayingPanel renders the now playing panel
func (m *Model) renderNowPlayingPanel(width, height int) string {
	borderStyle := inactiveBorderStyle
	titleStyle := panelTitleStyle

	title := titleStyle.Render(" Now Playing ")

	if m.nowPlaying == nil {
		var content strings.Builder
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("  No station playing"))
		content.WriteString("\n\n  Select a station and press Enter to play")

		panel := lipgloss.JoinVertical(lipgloss.Left, title, content.String())

		return borderStyle.
			Width(width).
			Height(height).
			Render(panel)
	}

	// Build station info (right side)
	var info strings.Builder

	status := "⏸ Paused"
	statusColor := lipgloss.Color("208")
	if m.isPlaying {
		status = "▶ Playing"
		statusColor = lipgloss.Color("86")
	}

	vol, _ := m.player.GetVolume()

	info.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render(m.nowPlaying.Name))
	info.WriteString("\n\n")

	info.WriteString(fmt.Sprintf("%s • %s",
		m.nowPlaying.Country,
		m.nowPlaying.Codec))
	info.WriteString("\n")

	info.WriteString(fmt.Sprintf("Bitrate: %d kbps",
		m.nowPlaying.Bitrate))
	info.WriteString("\n\n")

	info.WriteString(lipgloss.NewStyle().
		Foreground(statusColor).
		Render(status))
	info.WriteString("\n")

	info.WriteString(fmt.Sprintf("Volume: %d%%", vol))

	// Get icon (or show loading/placeholder)
	iconDisplay := m.stationIcon
	if iconDisplay == "" {
		iconDisplay = "  Loading\n  icon..."
	}

	// Combine icon and info horizontally
	contentLayout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		iconDisplay,
		"  ", // spacing
		info.String(),
	)

	panel := lipgloss.JoinVertical(lipgloss.Left, title, "\n"+contentLayout)

	return borderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

// renderFiltersPanel renders the filters panel
func (m *Model) renderFiltersPanel(width, height int) string {
	borderStyle := inactiveBorderStyle
	titleStyle := panelTitleStyle
	if m.focusedSection == SectionFilters {
		borderStyle = activeBorderStyle
		titleStyle = activePanelTitleStyle
	}

	title := titleStyle.Render(" Filters ")

	var content strings.Builder

	content.WriteString("\n")

	// If editing a filter, show text input
	if m.editingFilter != FilterNone {
		var fieldName string
		switch m.editingFilter {
		case FilterCountry:
			fieldName = "Country"
		case FilterGenre:
			fieldName = "Genre"
		case FilterLanguage:
			fieldName = "Language"
		}

		content.WriteString(fmt.Sprintf("  %s:\n", fieldName))
		content.WriteString("  " + m.filterInput.View())
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("  Enter: apply • Esc: cancel"))
	} else {
		// Country filter
		countryText := "All Countries"
		countryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		if m.filters.Country != "" {
			countryText = m.filters.Country
			countryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
		}
		content.WriteString(fmt.Sprintf("  1. Country: %s", countryStyle.Render(countryText)))
		content.WriteString("\n")

		// Genre filter
		genreText := "All Genres"
		genreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		if m.filters.Genre != "" {
			genreText = m.filters.Genre
			genreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
		}
		content.WriteString(fmt.Sprintf("  2. Genre: %s", genreStyle.Render(genreText)))
		content.WriteString("\n")

		// Language filter
		langText := "All Languages"
		langStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		if m.filters.Language != "" {
			langText = m.filters.Language
			langStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
		}
		content.WriteString(fmt.Sprintf("  3. Language: %s", langStyle.Render(langText)))
		content.WriteString("\n\n")

		// Help text
		if m.focusedSection == SectionFilters {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render("  1-3: edit filter • c: clear"))
		} else {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render("  Press Tab to focus"))
		}
	}

	panel := lipgloss.JoinVertical(lipgloss.Left, title, content.String())

	return borderStyle.
		Width(width).
		Height(height).
		Render(panel)
}

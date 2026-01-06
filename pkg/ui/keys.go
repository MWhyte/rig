package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If editing a filter, handle input differently
	if m.editingFilter != FilterNone {
		return m.handleFilterInput(msg)
	}

	// Global shortcuts (work in any section)
	switch msg.String() {
	case "ctrl+c", "q":
		m.stopPlayback()
		return m, tea.Quit

	case "tab":
		// Move to next section
		m.focusedSection = m.focusedSection.next()
		return m, nil

	case "shift+tab":
		// Move to previous section
		m.focusedSection = m.focusedSection.prev()
		return m, nil

	case " ":
		// Toggle play/pause
		if m.isPlaying && m.nowPlaying != nil {
			if err := m.player.Pause(); err == nil {
				m.isPlaying = false
			}
		} else if !m.isPlaying && m.nowPlaying != nil {
			if err := m.player.Resume(); err == nil {
				m.isPlaying = true
			}
		}
		return m, nil

	case "s":
		// Stop playback
		return m, func() tea.Msg {
			return stopPlaybackMsg{}
		}

	case "+", "=":
		// Increase volume
		vol, _ := m.player.GetVolume()
		newVol := vol + 10
		if newVol > 100 {
			newVol = 100
		}
		_ = m.player.SetVolume(newVol)
		return m, nil

	case "-", "_":
		// Decrease volume
		vol, _ := m.player.GetVolume()
		newVol := vol - 10
		if newVol < 0 {
			newVol = 0
		}
		_ = m.player.SetVolume(newVol)
		return m, nil

	case "?":
		// Toggle help
		if m.view == ViewHelp {
			m.view = ViewStationList
		} else {
			m.view = ViewHelp
		}
		return m, nil

	case "r":
		// Refresh stations
		m.view = ViewLoading
		if m.hasActiveFilters() {
			return m, m.fetchFilteredStations()
		}
		return m, m.fetchPopularStations()
	}

	// Section-specific shortcuts
	switch m.focusedSection {
	case SectionStationList:
		return m.handleStationListKeys(msg)

	case SectionFilters:
		return m.handleFiltersKeys(msg)
	}

	return m, nil
}

// handleStationListKeys handles keys when station list is focused
func (m *Model) handleStationListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Play selected station
		if m.view == ViewStationList && len(m.stations) > 0 {
			selected := m.stationList.Index()
			if selected >= 0 && selected < len(m.stations) {
				return m, func() tea.Msg {
					return playStationMsg{&m.stations[selected]}
				}
			}
		}
	}

	// Pass other keys to the list
	var cmd tea.Cmd
	m.stationList, cmd = m.stationList.Update(msg)
	return m, cmd
}

// handleFiltersKeys handles keys when filters section is focused
func (m *Model) handleFiltersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		// Clear all filters
		m.filters = Filters{}
		m.editingFilter = FilterNone
		return m, func() tea.Msg {
			return applyFiltersMsg{}
		}

	case "1":
		// Edit country filter
		m.editingFilter = FilterCountry
		m.filterInput.SetValue(m.filters.Country)
		m.filterInput.Focus()
		m.filterInput.Placeholder = "Enter country name..."
		return m, textinput.Blink

	case "2":
		// Edit genre filter
		m.editingFilter = FilterGenre
		m.filterInput.SetValue(m.filters.Genre)
		m.filterInput.Focus()
		m.filterInput.Placeholder = "Enter genre/tag..."
		return m, textinput.Blink

	case "3":
		// Edit language filter
		m.editingFilter = FilterLanguage
		m.filterInput.SetValue(m.filters.Language)
		m.filterInput.Focus()
		m.filterInput.Placeholder = "Enter language..."
		return m, textinput.Blink
	}

	return m, nil
}

// handleFilterInput handles keyboard input when editing a filter
func (m *Model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Allow quitting even during input
		m.stopPlayback()
		return m, tea.Quit

	case "esc":
		// Cancel editing
		m.editingFilter = FilterNone
		m.filterInput.Blur()
		return m, nil

	case "enter":
		// Apply the filter value
		value := m.filterInput.Value()

		switch m.editingFilter {
		case FilterCountry:
			m.filters.Country = value
			// Try to find matching country code
			for _, country := range m.countries {
				if country.Name == value {
					m.filters.CountryCode = country.ISO3166_1
					break
				}
			}
		case FilterGenre:
			m.filters.Genre = value
		case FilterLanguage:
			m.filters.Language = value
		}

		m.editingFilter = FilterNone
		m.filterInput.Blur()

		return m, func() tea.Msg {
			return applyFiltersMsg{}
		}
	}

	// Pass to text input
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	return m, cmd
}

// hasActiveFilters returns true if any filters are active
func (m *Model) hasActiveFilters() bool {
	return m.filters.CountryCode != "" ||
		m.filters.Genre != "" ||
		m.filters.Language != ""
}

package ui

import (
	"strings"
	"time"

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
		// Toggle play/pause, or play selected station
		if m.isPlaying && m.nowPlaying != nil {
			// Currently playing - pause it
			if err := m.player.Pause(); err == nil {
				m.isPlaying = false
			}
			return m, nil
		} else if !m.isPlaying && m.nowPlaying != nil {
			// Paused - resume it
			if err := m.player.Resume(); err == nil {
				m.isPlaying = true
			}
			return m, nil
		} else if m.focusedSection == SectionStationList && len(m.stations) > 0 {
			// No station playing and in station list - play selected station
			selected := m.stationList.Index()
			if selected >= 0 && selected < len(m.stations) {
				return m, func() tea.Msg {
					return playStationMsg{&m.stations[selected]}
				}
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
		m.autocomplete.SetFieldName("Country")
		m.autocomplete.SetSuggestions(m.autocompleteData[FilterCountry])
		m.autocomplete.SetValue(m.filters.Country)
		m.autocomplete.Filter(m.filters.Country)
		return m, m.autocomplete.Focus()

	case "2":
		// Edit genre filter
		m.editingFilter = FilterGenre
		m.filterInput.SetValue(m.filters.Genre)
		m.filterInput.Focus()
		m.filterInput.Placeholder = "Enter genre/tag..."
		m.autocomplete.SetFieldName("Genre")
		m.autocomplete.SetSuggestions(m.autocompleteData[FilterGenre])
		m.autocomplete.SetValue(m.filters.Genre)
		m.autocomplete.Filter(m.filters.Genre)
		return m, m.autocomplete.Focus()

	case "3":
		// Edit language filter
		m.editingFilter = FilterLanguage
		m.filterInput.SetValue(m.filters.Language)
		m.filterInput.Focus()
		m.filterInput.Placeholder = "Enter language..."
		m.autocomplete.SetFieldName("Language")
		m.autocomplete.SetSuggestions(m.autocompleteData[FilterLanguage])
		m.autocomplete.SetValue(m.filters.Language)
		m.autocomplete.Filter(m.filters.Language)
		return m, m.autocomplete.Focus()

	case "4":
		// Edit station name filter
		m.editingFilter = FilterStationName
		m.filterInput.SetValue(m.filters.StationName)
		m.filterInput.Focus()
		m.filterInput.Placeholder = "Type station name..."
		m.autocomplete.SetFieldName("Station")
		m.autocomplete.SetSuggestions([]string{}) // Empty until typing
		m.autocomplete.SetValue(m.filters.StationName)
		return m, m.autocomplete.Focus()
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
		m.autocomplete.Blur()
		return m, nil

	case "enter":
		// Use selected suggestion or typed value
		selected := m.autocomplete.GetSelected()
		value := ""

		if selected != "" {
			// Extract value from "Name (count)" format
			value = extractFilterValue(selected)
		} else {
			// Use typed value directly
			value = m.autocomplete.Value()
		}

		// Apply the filter value
		m.applyFilterValue(value)

		m.editingFilter = FilterNone
		m.filterInput.Blur()
		m.autocomplete.Blur()

		return m, func() tea.Msg {
			return applyFiltersMsg{}
		}

	case "up", "k", "down", "j":
		// Navigate autocomplete suggestions
		var cmd tea.Cmd
		m.autocomplete, cmd = m.autocomplete.Update(msg)
		return m, cmd
	}

	// Update text input and filter autocomplete
	var cmd tea.Cmd
	m.autocomplete, cmd = m.autocomplete.UpdateTextInput(msg)

	// Update autocomplete suggestions based on new input
	query := m.autocomplete.Value()
	searchCmd := m.updateAutocompleteSuggestions(query)

	// Return both commands
	if searchCmd != nil {
		return m, tea.Batch(cmd, searchCmd)
	}

	return m, cmd
}

// extractFilterValue extracts the actual value from "Name (count)" format
func extractFilterValue(suggestion string) string {
	// Find the last opening parenthesis
	lastParen := -1
	for i := len(suggestion) - 1; i >= 0; i-- {
		if suggestion[i] == '(' {
			lastParen = i
			break
		}
	}

	if lastParen > 0 {
		// Trim space before parenthesis
		return strings.TrimSpace(suggestion[:lastParen])
	}

	return suggestion
}

// applyFilterValue applies a filter value to the appropriate filter field
func (m *Model) applyFilterValue(value string) {
	switch m.editingFilter {
	case FilterCountry:
		m.filters.Country = value
		// Try to find matching country code
		m.filters.CountryCode = ""
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
	case FilterStationName:
		m.filters.StationName = value
	}
}

// updateAutocompleteSuggestions updates autocomplete suggestions based on query
func (m *Model) updateAutocompleteSuggestions(query string) tea.Cmd {
	switch m.editingFilter {
	case FilterCountry, FilterGenre, FilterLanguage:
		// Use precomputed metadata with fuzzy filtering
		m.autocomplete.Filter(query)
		return nil

	case FilterStationName:
		// For station name, we need live API search
		if len(query) < 2 {
			m.autocomplete.SetSuggestions([]string{})
			return nil
		}

		// Check cache first
		if cached, ok := m.stationNameCache[query]; ok {
			m.autocomplete.SetSuggestions(cached)
			m.autocomplete.Filter("")
			return nil
		}

		// Trigger debounced search
		// Return a command that waits 300ms then fetches suggestions
		return tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
			return m.fetchStationNameSuggestions(query)()
		})
	}

	return nil
}

// hasActiveFilters returns true if any filters are active
func (m *Model) hasActiveFilters() bool {
	return m.filters.CountryCode != "" ||
		m.filters.Genre != "" ||
		m.filters.Language != "" ||
		m.filters.StationName != ""
}

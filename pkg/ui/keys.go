package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mrwhyte/rig/pkg/config"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.showThemeModal {
		return m.handleThemeModalInput(msg)
	}

	// If timer modal is open, handle input differently
	if m.showTimerModal {
		return m.handleTimerModalInput(msg)
	}

	// If editing a filter, handle input differently
	if m.editingFilter != FilterNone {
		return m.handleFilterInput(msg)
	}

	// If station list is filtering, pass keys to it first (except ctrl+c)
	if m.ready && m.focusedSection == SectionStationList && m.stationList.FilterState() == list.Filtering {
		if msg.String() != "ctrl+c" {
			var cmd tea.Cmd
			m.stationList, cmd = m.stationList.Update(msg)
			return m, cmd
		}
	}

	// Global shortcuts (work in any section)
	switch msg.String() {
	case "ctrl+c":
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

	case "space":
		// Toggle play/pause, or play selected station
		if m.isPlaying && m.playing != nil {
			// Currently playing - pause it
			if err := m.player.Pause(); err == nil {
				m.isPlaying = false
				if m.sleepTimerActive {
					m.sleepTimerPaused = true
					elapsed := time.Since(m.sleepTimerStart)
					m.sleepTimerRemaining = m.sleepTimerDuration - elapsed
				}
			}
			return m, nil
		} else if !m.isPlaying && m.playing != nil {
			// Paused - resume it
			if err := m.player.Resume(); err == nil {
				m.isPlaying = true
				if m.sleepTimerActive && m.sleepTimerPaused {
					m.sleepTimerPaused = false
					m.sleepTimerDuration = m.sleepTimerRemaining
					m.sleepTimerStart = time.Now()
					return m, tea.Batch(m.waveTick(), m.sleepTimerTick())
				}
				return m, m.waveTick()
			}
			return m, nil
		} else if m.focusedSection == SectionStationList && len(m.stations) > 0 {
			// No station playing and in station list - play selected station
			// Get the actual selected item from the filtered list
			if item := m.stationList.SelectedItem(); item != nil {
				if stationItem, ok := item.(StationItem); ok {
					return m, func() tea.Msg {
						return playStationMsg{&stationItem.station}
					}
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
		newVol := min(vol+5, 100)
		_ = m.player.SetVolume(newVol)
		return m, m.volumeBar.SetPercent(float64(newVol) / 100.0)

	case "-", "_":
		// Decrease volume
		vol, _ := m.player.GetVolume()
		newVol := max(vol-5, 0)
		_ = m.player.SetVolume(newVol)
		return m, m.volumeBar.SetPercent(float64(newVol) / 100.0)

	case "ctrl+t":
		// Open theme picker modal
		m.originalThemeIndex = themeIndex
		m.themeModalIndex = themeIndex
		m.showThemeModal = true
		return m, nil

	case "t":
		// Open timer modal
		m.showTimerModal = true
		// Pre-fill with current timer duration if active
		if m.sleepTimerActive {
			minutes := int(m.sleepTimerRemaining.Minutes())
			m.timerInput.SetValue(fmt.Sprintf("%d", minutes))
			m.timerInput.CursorEnd()
		} else {
			m.timerInput.SetValue("")
		}
		return m, m.timerInput.Focus()

	case "f":
		// Toggle favorite for selected station
		if m.focusedSection == SectionStationList && len(m.stations) > 0 {
			// Get the actual selected item from the filtered list
			if item := m.stationList.SelectedItem(); item != nil {
				if stationItem, ok := item.(StationItem); ok {
					if m.favManager != nil {
						if err := m.favManager.Toggle(
							stationItem.station.StationUUID,
							stationItem.station.Name,
							stationItem.station.URLResolved,
						); err != nil {
							m.err = fmt.Errorf("failed to save favorite: %w", err)
						}
						// Refresh list to update ★ indicator
						m.initList()
					}
				}
			}
		}
		return m, nil
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
func (m *Model) handleStationListKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Play selected station
		if m.view == ViewStationList && len(m.stations) > 0 {
			// Get the actual selected item from the filtered list
			if item := m.stationList.SelectedItem(); item != nil {
				if stationItem, ok := item.(StationItem); ok {
					return m, func() tea.Msg {
						return playStationMsg{&stationItem.station}
					}
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
func (m *Model) handleFiltersKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Navigate up through filters
		if m.selectedFilterIndex > 0 {
			m.selectedFilterIndex--
		}
		return m, nil

	case "down", "j":
		// Navigate down through filters
		if m.selectedFilterIndex < 4 { // 0-4 = 5 filter options
			m.selectedFilterIndex++
		}
		return m, nil

	case "enter":
		// Activate the selected filter
		switch m.selectedFilterIndex {
		case 0:
			// Edit country filter
			m.editingFilter = FilterCountry
			m.autocomplete.Reset(m.filters.Country) // Create fresh textinput with current value
			m.autocomplete.SetFieldName("Country")
			m.autocomplete.SetSuggestions(m.autocompleteData[FilterCountry])
			m.autocomplete.Filter(m.filters.Country)
			return m, m.autocomplete.Focus()

		case 1:
			// Edit genre filter
			m.editingFilter = FilterGenre
			m.autocomplete.Reset(m.filters.Genre) // Create fresh textinput with current value
			m.autocomplete.SetFieldName("Genre")
			m.autocomplete.SetSuggestions(m.autocompleteData[FilterGenre])
			m.autocomplete.Filter(m.filters.Genre)
			return m, m.autocomplete.Focus()

		case 2:
			// Edit language filter
			m.editingFilter = FilterLanguage
			m.autocomplete.Reset(m.filters.Language) // Create fresh textinput with current value
			m.autocomplete.SetFieldName("Language")
			m.autocomplete.SetSuggestions(m.autocompleteData[FilterLanguage])
			m.autocomplete.Filter(m.filters.Language)
			return m, m.autocomplete.Focus()

		case 3:
			// Edit station name filter
			m.editingFilter = FilterStationName
			m.autocomplete.Reset(m.filters.StationName) // Create fresh textinput with current value
			m.autocomplete.SetFieldName("Station")
			m.autocomplete.SetSuggestions([]string{})
			return m, m.autocomplete.Focus()

		case 4:
			// Toggle favorites filter
			m.filters.FavoritesOnly = !m.filters.FavoritesOnly
			m.focusedSection = SectionStationList

			// If turning off favorites and no other filters are active, return to popular stations
			if !m.filters.FavoritesOnly &&
				m.filters.CountryCode == "" &&
				m.filters.Genre == "" &&
				m.filters.Language == "" &&
				m.filters.StationName == "" {
				return m, m.fetchPopularStations()
			}

			return m, func() tea.Msg {
				return applyFiltersMsg{}
			}
		}
		return m, nil

	case "c":
		// Clear all filters and return to popular stations
		m.filters = Filters{}
		m.editingFilter = FilterNone
		return m, m.fetchPopularStations()

	case "1":
		// Edit country filter
		m.editingFilter = FilterCountry
		m.autocomplete.Reset(m.filters.Country) // Create fresh textinput with current value
		m.autocomplete.SetFieldName("Country")
		m.autocomplete.SetSuggestions(m.autocompleteData[FilterCountry])
		m.autocomplete.Filter(m.filters.Country)
		return m, m.autocomplete.Focus()

	case "2":
		// Edit genre filter
		m.editingFilter = FilterGenre
		m.autocomplete.Reset(m.filters.Genre) // Create fresh textinput with current value
		m.autocomplete.SetFieldName("Genre")
		m.autocomplete.SetSuggestions(m.autocompleteData[FilterGenre])
		m.autocomplete.Filter(m.filters.Genre)
		return m, m.autocomplete.Focus()

	case "3":
		// Edit language filter
		m.editingFilter = FilterLanguage
		m.autocomplete.Reset(m.filters.Language) // Create fresh textinput with current value
		m.autocomplete.SetFieldName("Language")
		m.autocomplete.SetSuggestions(m.autocompleteData[FilterLanguage])
		m.autocomplete.Filter(m.filters.Language)
		return m, m.autocomplete.Focus()

	case "4":
		// Edit station name filter
		m.editingFilter = FilterStationName
		m.autocomplete.Reset(m.filters.StationName) // Create fresh textinput with current value
		m.autocomplete.SetFieldName("Station")
		m.autocomplete.SetSuggestions([]string{}) // Empty until typing
		return m, m.autocomplete.Focus()

	case "5":
		// Toggle favorites filter
		m.filters.FavoritesOnly = !m.filters.FavoritesOnly
		// Switch focus to station list to see results
		m.focusedSection = SectionStationList

		// If turning off favorites and no other filters are active, return to popular stations
		if !m.filters.FavoritesOnly &&
			m.filters.CountryCode == "" &&
			m.filters.Genre == "" &&
			m.filters.Language == "" &&
			m.filters.StationName == "" {
			return m, m.fetchPopularStations()
		}

		return m, func() tea.Msg {
			return applyFiltersMsg{}
		}
	}

	return m, nil
}

// handleFilterInput handles keyboard input when editing a filter
func (m *Model) handleFilterInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Allow quitting even during input
		m.stopPlayback()
		return m, tea.Quit

	case "esc":
		// Cancel editing
		m.editingFilter = FilterNone
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
		m.autocomplete.Blur()

		// Stay in filters section so user can add more filters
		// (User can Tab to station list when ready)

		return m, func() tea.Msg {
			return applyFiltersMsg{}
		}

	case "up", "down":
		// Navigate autocomplete suggestions (only arrow keys, so j/k can be typed)
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
		m.filters.StationName != "" ||
		m.filters.FavoritesOnly
}

// handleTimerModalInput handles keyboard input in the timer modal
func (m *Model) handleTimerModalInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Check for Enter key first (before textinput consumes it)
	if msg.Code == tea.KeyEnter {
		// Start/update timer with entered value
		value := strings.TrimSpace(m.timerInput.Value())
		if value == "" {
			// Empty input, just close modal
			m.showTimerModal = false
			m.timerInput.Blur()
			return m, nil
		}

		// Parse minutes using strconv.Atoi
		mins, err := strconv.Atoi(value)
		if err != nil || mins <= 0 {
			// Invalid input, just close modal
			m.showTimerModal = false
			m.timerInput.Blur()
			return m, nil
		}

		// Close modal and start timer
		m.showTimerModal = false
		m.timerInput.Blur()

		return m, func() tea.Msg {
			return sleepTimerSetMsg{duration: time.Duration(mins) * time.Minute}
		}
	}

	switch msg.String() {
	case "ctrl+c":
		// Allow quitting even during timer modal
		m.stopPlayback()
		return m, tea.Quit

	case "esc":
		// Close modal without changes
		m.showTimerModal = false
		m.timerInput.Blur()
		return m, nil

	case "x":
		// Cancel active timer
		if m.sleepTimerActive {
			m.showTimerModal = false
			m.timerInput.Blur()
			return m, func() tea.Msg {
				return sleepTimerCancelledMsg{}
			}
		}
		return m, nil
	}

	// Update text input
	var cmd tea.Cmd
	m.timerInput, cmd = m.timerInput.Update(msg)
	return m, cmd
}

// handleThemeModalInput handles keyboard input in the theme picker modal.
func (m *Model) handleThemeModalInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.stopPlayback()
		return m, tea.Quit

	case "up", "k":
		m.themeModalIndex = max(m.themeModalIndex-1, 0)
		m.applyTheme(m.themeModalIndex)
		return m, nil

	case "down", "j":
		m.themeModalIndex = min(m.themeModalIndex+1, len(themes)-1)
		m.applyTheme(m.themeModalIndex)
		return m, nil

	case "enter":
		// Confirm — save to config
		m.showThemeModal = false
		cfg := &config.Config{Theme: themes[themeIndex].Name}
		_ = config.Save(cfg)
		return m, nil

	case "esc":
		// Revert to original theme
		m.applyTheme(m.originalThemeIndex)
		m.showThemeModal = false
		return m, nil
	}

	return m, nil
}

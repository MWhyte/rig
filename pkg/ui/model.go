package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrwhyte/rig/pkg/player"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewLoading ViewMode = iota
	ViewStationList
	ViewSearch
	ViewHelp
)

// FilterField represents which filter field is being edited
type FilterField int

const (
	FilterNone FilterField = iota
	FilterCountry
	FilterGenre
	FilterLanguage
	FilterStationName
)

// Filters represents the current filter state
type Filters struct {
	Country     string
	CountryCode string
	Genre       string
	Language    string
	StationName string
}

// Model is the main application model
type Model struct {
	// UI state
	view   ViewMode
	width  int
	height int
	ready  bool
	err    error

	// Focus management
	focusedSection Section

	// Components
	stationList list.Model
	stations    []radiobrowser.Station

	// Playback
	player     player.Player
	nowPlaying *radiobrowser.Station
	isPlaying  bool

	// API
	apiClient *radiobrowser.Client

	// Filters
	filters       Filters
	countries     []radiobrowser.Country
	tags          []radiobrowser.Tag
	languages     []radiobrowser.Language
	editingFilter FilterField
	filterInput   textinput.Model

	// Autocomplete
	autocomplete     AutocompleteModel
	autocompleteData map[FilterField][]string
	stationNameCache map[string][]string

	// Search
	searchQuery string
	searching   bool
}

// NewModel creates a new application model
func NewModel() (*Model, error) {
	// Create API client
	apiClient, err := radiobrowser.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Create player
	p, err := player.NewMPVPlayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	// Create filter text input
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50

	m := &Model{
		view:             ViewLoading,
		apiClient:        apiClient,
		player:           p,
		focusedSection:   SectionStationList,
		filters:          Filters{},
		editingFilter:    FilterNone,
		filterInput:      ti,
		autocomplete:     NewAutocompleteModel(),
		autocompleteData: make(map[FilterField][]string),
		stationNameCache: make(map[string][]string),
	}

	return m, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchPopularStations(),
		m.fetchMetadata(),
	)
}

// fetchPopularStations fetches popular stations from the API
func (m *Model) fetchPopularStations() tea.Cmd {
	return func() tea.Msg {
		stations, err := m.apiClient.GetPopularStations(100)
		if err != nil {
			return errMsg{err}
		}
		return stationsLoadedMsg{stations}
	}
}

// fetchMetadata fetches countries, tags, and languages
func (m *Model) fetchMetadata() tea.Cmd {
	return func() tea.Msg {
		countries, err := m.apiClient.GetCountries()
		if err != nil {
			return errMsg{err}
		}

		tags, err := m.apiClient.GetTags()
		if err != nil {
			return errMsg{err}
		}

		languages, err := m.apiClient.GetLanguages()
		if err != nil {
			return errMsg{err}
		}

		return metadataLoadedMsg{countries, tags, languages}
	}
}

// buildAutocompleteData builds autocomplete suggestions from metadata
func (m *Model) buildAutocompleteData() {
	// Country suggestions
	countrySugs := make([]string, len(m.countries))
	for i, c := range m.countries {
		countrySugs[i] = fmt.Sprintf("%s (%d stations)", c.Name, c.StationCount)
	}
	m.autocompleteData[FilterCountry] = countrySugs

	// Genre/Tag suggestions
	tagSugs := make([]string, len(m.tags))
	for i, t := range m.tags {
		tagSugs[i] = fmt.Sprintf("%s (%d stations)", t.Name, t.StationCount)
	}
	m.autocompleteData[FilterGenre] = tagSugs

	// Language suggestions
	langSugs := make([]string, len(m.languages))
	for i, l := range m.languages {
		langSugs[i] = fmt.Sprintf("%s (%d stations)", l.Name, l.StationCount)
	}
	m.autocompleteData[FilterLanguage] = langSugs

	// Station name suggestions are populated on-demand (not precomputed)
}

// fetchStationNameSuggestions fetches station name suggestions from the API
func (m *Model) fetchStationNameSuggestions(query string) tea.Cmd {
	return func() tea.Msg {
		// Search by name
		stations, err := m.apiClient.SearchByName(query)
		if err != nil {
			return errMsg{err}
		}

		// Limit to 10 suggestions
		limit := 10
		if len(stations) > limit {
			stations = stations[:limit]
		}

		// Format suggestions as "Station Name (Country)"
		suggestions := make([]string, len(stations))
		for i, s := range stations {
			suggestions[i] = fmt.Sprintf("%s (%s)", s.Name, s.Country)
		}

		return stationNameSuggestionsMsg{query, suggestions}
	}
}

// fetchFilteredStations fetches stations based on current filters
func (m *Model) fetchFilteredStations() tea.Cmd {
	return func() tea.Msg {
		params := radiobrowser.SearchParams{
			Name:        m.filters.StationName,
			CountryCode: m.filters.CountryCode,
			Tag:         m.filters.Genre,
			Language:    m.filters.Language,
			Order:       "clickcount",
			Reverse:     true,
			Limit:       100,
			HideBroken:  true,
		}

		stations, err := m.apiClient.SearchStations(params)
		if err != nil {
			return errMsg{err}
		}

		return stationsLoadedMsg{stations}
	}
}

// Messages
type errMsg struct{ err error }
type stationsLoadedMsg struct{ stations []radiobrowser.Station }
type metadataLoadedMsg struct {
	countries []radiobrowser.Country
	tags      []radiobrowser.Tag
	languages []radiobrowser.Language
}
type playStationMsg struct{ station *radiobrowser.Station }
type stopPlaybackMsg struct{}
type applyFiltersMsg struct{}
type stationNameSuggestionsMsg struct {
	query       string
	suggestions []string
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.ready = true
			m.initList()
		} else {
			m.stationList.SetSize(msg.Width, msg.Height-10)
		}

		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case stationsLoadedMsg:
		m.stations = msg.stations
		m.view = ViewStationList
		m.initList()
		return m, nil

	case metadataLoadedMsg:
		m.countries = msg.countries
		m.tags = msg.tags
		m.languages = msg.languages
		m.buildAutocompleteData()
		return m, nil

	case applyFiltersMsg:
		return m, m.fetchFilteredStations()

	case playStationMsg:
		return m.playStation(msg.station)

	case stopPlaybackMsg:
		m.stopPlayback()
		return m, nil

	case stationNameSuggestionsMsg:
		// Cache the results
		m.stationNameCache[msg.query] = msg.suggestions

		// Update autocomplete if still editing station name and query matches
		if m.editingFilter == FilterStationName && m.autocomplete.Value() == msg.query {
			m.autocomplete.SetSuggestions(msg.suggestions)
			m.autocomplete.Filter("") // Reset filter to show all suggestions
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	// Update the list only if station list is focused
	if m.focusedSection == SectionStationList {
		var cmd tea.Cmd
		m.stationList, cmd = m.stationList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	if !m.ready {
		return "Initializing rig.fm...\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}

	switch m.view {
	case ViewLoading:
		return "Loading stations...\n"

	case ViewStationList:
		return m.renderStationList()

	case ViewHelp:
		return m.renderHelp()

	default:
		return "Unknown view\n"
	}
}

// Close cleans up resources
func (m *Model) Close() error {
	if m.player != nil {
		return m.player.Close()
	}
	return nil
}

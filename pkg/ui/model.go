package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrwhyte/rig/pkg/favorites"
	"github.com/mrwhyte/rig/pkg/player"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewLoading ViewMode = iota
	ViewStationList
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
	Country       string
	CountryCode   string
	Genre         string
	Language      string
	StationName   string
	FavoritesOnly bool
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

	// Favorites
	favManager *favorites.Manager

	// Metadata
	currentSong string
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

	// Create favorites manager
	favManager, err := favorites.NewManager()
	if err != nil {
		// Log error but don't fail - favorites not critical
		fmt.Fprintf(os.Stderr, "Warning: Could not load favorites: %v\n", err)
	}

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
		favManager:       favManager,
	}

	return m, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchPopularStations(),
		m.fetchMetadata(),
		m.tick(),
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
		// Use SearchStations with a limit for efficiency (only fetch what we need)
		params := radiobrowser.SearchParams{
			Name:       query,
			Order:      "clickcount",
			Reverse:    true,
			Limit:      10, // Only fetch 10 stations for autocomplete
			HideBroken: true,
		}
		stations, err := m.apiClient.SearchStations(params)
		if err != nil {
			return errMsg{err}
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
		// If favorites-only mode, get favorites and filter by other criteria
		if m.filters.FavoritesOnly {
			return m.fetchFavoritesFiltered()()
		}

		// Normal API search with filters
		params := radiobrowser.SearchParams{
			Name:        m.filters.StationName,
			CountryCode: m.filters.CountryCode,
			Tag:         m.filters.Genre,
			Language:    m.filters.Language,
			Order:       "clickcount",
			Reverse:     true,
			Limit:       0, // No limit - show all matching stations
			HideBroken:  true,
		}

		stations, err := m.apiClient.SearchStations(params)
		if err != nil {
			return errMsg{err}
		}

		return stationsLoadedMsg{stations}
	}
}

// fetchFavoritesFiltered fetches favorites and applies other filters
func (m *Model) fetchFavoritesFiltered() tea.Cmd {
	return func() tea.Msg {
		// Get all favorite UUIDs
		if m.favManager == nil {
			return stationsLoadedMsg{stations: []radiobrowser.Station{}}
		}

		favs := m.favManager.GetAll()
		if len(favs) == 0 {
			return stationsLoadedMsg{stations: []radiobrowser.Station{}}
		}

		// Search by UUIDs to get fresh metadata from API
		uuids := make([]string, len(favs))
		for i, fav := range favs {
			uuids[i] = fav.StationUUID
		}

		stations, err := m.apiClient.SearchByUUIDs(uuids)
		if err != nil {
			return errMsg{err}
		}

		// Apply other filters client-side
		filtered := make([]radiobrowser.Station, 0, len(stations))
		for _, station := range stations {
			// Check country filter
			if m.filters.CountryCode != "" && station.CountryCode != m.filters.CountryCode {
				continue
			}

			// Check genre filter (tags contain genre)
			if m.filters.Genre != "" {
				if !strings.Contains(strings.ToLower(station.Tags), strings.ToLower(m.filters.Genre)) {
					continue
				}
			}

			// Check language filter
			if m.filters.Language != "" && !strings.EqualFold(station.Language, m.filters.Language) {
				continue
			}

			// Check station name filter
			if m.filters.StationName != "" {
				if !strings.Contains(strings.ToLower(station.Name), strings.ToLower(m.filters.StationName)) {
					continue
				}
			}

			filtered = append(filtered, station)
		}

		return stationsLoadedMsg{stations: filtered}
	}
}

// tick returns a command that triggers metadata polling
func (m *Model) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// pollMetadata queries the player for current metadata
func (m *Model) pollMetadata() tea.Cmd {
	return func() tea.Msg {
		metadata, err := m.player.GetMetadata()
		if err != nil {
			// Silently fail - not critical
			return metadataUpdateMsg{song: ""}
		}

		return metadataUpdateMsg{song: metadata.Title}
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
type tickMsg struct{}
type metadataUpdateMsg struct {
	song string
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

	case tea.MouseMsg:
		// Handle mouse clicks to switch sections
		if msg.Type == tea.MouseLeft {
			return m.handleMouseClick(msg)
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

	case tickMsg:
		// Poll metadata if playing
		if m.isPlaying && m.player != nil {
			return m, tea.Batch(
				m.tick(),         // Schedule next tick
				m.pollMetadata(), // Poll current metadata
			)
		}
		// Still tick even if not playing (for responsiveness when playback starts)
		return m, m.tick()

	case metadataUpdateMsg:
		m.currentSong = msg.song
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

	default:
		return "Unknown view\n"
	}
}

// handleMouseClick handles mouse click events to switch sections
func (m *Model) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Don't process clicks if not ready or not in station list view
	if !m.ready || m.view != ViewStationList {
		return m, nil
	}

	// Calculate layout boundaries (matching renderMultiPanelLayout)
	topSectionHeight := int(float64(m.height) * 0.30)

	// Header takes about 2 lines
	headerHeight := 2

	// Check if click is in top section (Filters area)
	if msg.Y >= headerHeight && msg.Y < headerHeight+topSectionHeight {
		// Top section - check if left half (Filters) or right half (Now Playing)
		halfWidth := m.width / 2
		if msg.X < halfWidth {
			// Clicked on Filters section
			m.focusedSection = SectionFilters
		}
		// Right half is Now Playing - not focusable, so ignore
		return m, nil
	}

	// Check if click is in bottom section (Station List)
	if msg.Y >= headerHeight+topSectionHeight {
		// Clicked on Station List section
		m.focusedSection = SectionStationList
		return m, nil
	}

	return m, nil
}

// Close cleans up resources
func (m *Model) Close() error {
	if m.player != nil {
		return m.player.Close()
	}
	return nil
}

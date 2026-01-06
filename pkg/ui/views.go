package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrwhyte/rig/pkg/icon"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	nowPlayingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))
)

// StationItem implements list.Item for station list
type StationItem struct {
	station radiobrowser.Station
}

func (i StationItem) Title() string       { return i.station.Name }
func (i StationItem) Description() string {
	return fmt.Sprintf("%s • %s • %d kbps • %d clicks",
		i.station.Country,
		i.station.Codec,
		i.station.Bitrate,
		i.station.ClickCount)
}
func (i StationItem) FilterValue() string { return i.station.Name }

// initList initializes the station list
func (m *Model) initList() {
	items := make([]list.Item, len(m.stations))
	for i, station := range m.stations {
		items[i] = StationItem{station: station}
	}

	delegate := list.NewDefaultDelegate()
	m.stationList = list.New(items, delegate, 0, 0)
	m.stationList.Title = "Popular Radio Stations"
	m.stationList.SetShowStatusBar(true)
	m.stationList.SetFilteringEnabled(true)

	if m.width > 0 && m.height > 0 {
		m.stationList.SetSize(m.width, m.height-10)
	}
}

// renderStationList renders the station list view (uses multi-panel layout)
func (m *Model) renderStationList() string {
	return m.renderMultiPanelLayout()
}

// renderHelp renders the help view
func (m *Model) renderHelp() string {
	help := `
  ██████╗ ██╗ ██████╗     ███████╗███╗   ███╗
  ██╔══██╗██║██╔════╝     ██╔════╝████╗ ████║
  ██████╔╝██║██║  ███╗    █████╗  ██╔████╔██║
  ██╔══██╗██║██║   ██║    ██╔══╝  ██║╚██╔╝██║
  ██║  ██║██║╚██████╔╝    ██║     ██║ ╚═╝ ██║
  ╚═╝  ╚═╝╚═╝ ╚═════╝     ╚═╝     ╚═╝     ╚═╝

  The most beautiful terminal radio experience

  KEYBOARD SHORTCUTS

  Navigation:
    ↑/↓ or j/k     Navigate station list
    enter          Play selected station
    /              Filter/search stations

  Playback:
    space          Pause/resume playback
    s              Stop playback
    + or =         Increase volume
    - or _         Decrease volume

  General:
    r              Refresh station list
    ?              Toggle this help
    q or ctrl+c    Quit

  ABOUT

  rig.fm is powered by Radio Browser (radio-browser.info)
  A free, community-driven radio station database

  Press ? to return to the station list
`

	return help
}

// playStation starts playing a station
func (m *Model) playStation(station *radiobrowser.Station) (tea.Model, tea.Cmd) {
	// Stop current playback
	m.stopPlayback()

	// Track click
	go m.apiClient.TrackClick(station.StationUUID)

	// Start playback
	if err := m.player.Play(station.URLResolved); err != nil {
		m.err = fmt.Errorf("failed to play station: %w", err)
		return m, nil
	}

	m.nowPlaying = station
	m.isPlaying = true
	m.stationIcon = "" // Clear old icon

	// Load station icon in background
	return m, m.loadStationIcon(station.Favicon)
}

// loadStationIcon loads a station icon in the background
func (m *Model) loadStationIcon(faviconURL string) tea.Cmd {
	return func() tea.Msg {
		iconStr, err := icon.FetchAndRender(faviconURL)
		if err != nil {
			// Return placeholder on error
			iconStr, _ = icon.FetchAndRender("")
		}
		return iconLoadedMsg{iconStr}
	}
}

// stopPlayback stops current playback
func (m *Model) stopPlayback() {
	if m.player != nil {
		_ = m.player.Stop()
	}
	m.nowPlaying = nil
	m.isPlaying = false
	m.stationIcon = ""
}

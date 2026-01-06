package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// StationItem implements list.Item for station list
type StationItem struct {
	station radiobrowser.Station
}

func (i StationItem) Title() string { return i.station.Name }
func (i StationItem) Description() string {
	// Format tags
	tags := i.station.Tags
	if tags == "" {
		tags = "no tags"
	} else {
		// Limit to 25 chars for space
		if len(tags) > 25 {
			tags = tags[:22] + "..."
		}
	}

	return fmt.Sprintf("%s • %s • %s • %d kbps • %d clicks",
		i.station.Country,
		tags,
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

	// Disable default quit keys (q and esc) - we only want ctrl+c
	m.stationList.KeyMap.Quit.SetEnabled(false)
	m.stationList.KeyMap.ForceQuit.SetEnabled(false)
	m.stationList.KeyMap.CloseFullHelp.SetEnabled(false)

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
    ctrl+c         Quit

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

	return m, nil
}

// stopPlayback stops current playback
func (m *Model) stopPlayback() {
	if m.player != nil {
		_ = m.player.Stop()
	}
	m.nowPlaying = nil
	m.isPlaying = false
}

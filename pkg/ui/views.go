package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// StationItem implements list.Item for station list
type StationItem struct {
	station    radiobrowser.Station
	isFavorite bool
}

func (i StationItem) Title() string {
	name := i.station.Name
	// Add ★ if favorited
	if i.isFavorite {
		name = name + " ★"
	}
	return name
}
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
		isFavorite := false
		if m.favManager != nil {
			isFavorite = m.favManager.IsFavorite(station.StationUUID)
		}
		items[i] = StationItem{
			station:    station,
			isFavorite: isFavorite,
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(colorAccent).
		BorderLeftForeground(colorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(colorMuted).
		BorderLeftForeground(colorAccent)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(colorTitle)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(colorDim)
	m.stationList = list.New(items, delegate, 0, 0)

	// Style all built-in list chrome to match our palette
	m.stationList.Styles.Title = m.stationList.Styles.Title.
		Foreground(colorTitle).
		Background(lipgloss.NoColor{})
	m.stationList.Styles.TitleBar = m.stationList.Styles.TitleBar.
		Background(lipgloss.NoColor{})
	m.stationList.Styles.Filter.Focused.Prompt = m.stationList.Styles.Filter.Focused.Prompt.
		Foreground(colorAccent)
	m.stationList.Styles.Filter.Cursor.Color = colorTitle
	m.stationList.Styles.DefaultFilterCharacterMatch = m.stationList.Styles.DefaultFilterCharacterMatch.
		Foreground(colorAccent)
	m.stationList.Styles.StatusBar = m.stationList.Styles.StatusBar.
		Foreground(colorMuted)
	m.stationList.Styles.StatusBarActiveFilter = m.stationList.Styles.StatusBarActiveFilter.
		Foreground(colorAccent)
	m.stationList.Styles.StatusBarFilterCount = m.stationList.Styles.StatusBarFilterCount.
		Foreground(colorMuted)
	m.stationList.Styles.NoItems = m.stationList.Styles.NoItems.
		Foreground(colorMuted)
	m.stationList.Styles.ActivePaginationDot = m.stationList.Styles.ActivePaginationDot.
		Foreground(colorAccent)
	m.stationList.Styles.InactivePaginationDot = m.stationList.Styles.InactivePaginationDot.
		Foreground(colorBorder)
	m.stationList.Styles.HelpStyle = m.stationList.Styles.HelpStyle.
		Foreground(colorMuted)
	m.stationList.Styles.DividerDot = m.stationList.Styles.DividerDot.
		Foreground(colorBorder)

	// Disable the built-in title - we render our own in the panel
	m.stationList.Title = ""
	m.stationList.SetShowTitle(false)

	m.stationList.SetShowStatusBar(true)
	m.stationList.SetFilteringEnabled(true)

	// Update help to include paging instructions
	m.stationList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("left", "right"),
				key.WithHelp("←/→", "page"),
			),
		}
	}
	m.stationList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("left", "right"),
				key.WithHelp("←/→", "page up/down"),
			),
		}
	}

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

	m.playing = station
	m.isPlaying = true
	m.waveFrame = 0

	return m, m.waveTick()
}

// applyTheme sets the theme by index and refreshes all theme-dependent UI state.
func (m *Model) applyTheme(index int) {
	setTheme(index)
	// Refresh volume bar colors
	m.volumeBar.FullColor = colorAccent
	m.volumeBar.EmptyColor = colorBorder
	// Refresh list delegate styles
	m.initList()
}

// stopPlayback stops current playback
func (m *Model) stopPlayback() {
	if m.player != nil {
		_ = m.player.Stop()
	}
	m.playing = nil
	m.isPlaying = false
	m.currentSong = ""
	m.currentGenre = ""
}

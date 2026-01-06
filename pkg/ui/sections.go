package ui

// Section represents a UI section/panel
type Section int

const (
	SectionStationList Section = iota
	SectionNowPlaying
	SectionFilters
)

// sectionNames maps sections to their display names
var sectionNames = map[Section]string{
	SectionStationList: "Station List",
	SectionNowPlaying:  "Now Playing",
	SectionFilters:     "Filters",
}

// nextSection returns the next section in the cycle
func (s Section) next() Section {
	switch s {
	case SectionStationList:
		return SectionFilters
	case SectionFilters:
		return SectionStationList
	default:
		return SectionStationList
	}
}

// prevSection returns the previous section in the cycle
func (s Section) prev() Section {
	switch s {
	case SectionStationList:
		return SectionFilters
	case SectionFilters:
		return SectionStationList
	default:
		return SectionStationList
	}
}

// String returns the section name
func (s Section) String() string {
	if name, ok := sectionNames[s]; ok {
		return name
	}
	return "Unknown"
}

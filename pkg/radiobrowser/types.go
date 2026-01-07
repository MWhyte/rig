package radiobrowser

import (
	"strings"
	"time"
)

// RadioBrowserTime is a custom time type to handle Radio Browser's timestamp format
type RadioBrowserTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Radio Browser timestamps
func (t *RadioBrowserTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	// Handle empty string
	if s == "" || s == "null" {
		t.Time = time.Time{}
		return nil
	}

	// Try common Radio Browser formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}

	var err error
	for _, format := range formats {
		t.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return err
}

// Station represents a radio station from the Radio Browser API
type Station struct {
	StationUUID     string           `json:"stationuuid"`
	Name            string           `json:"name"`
	URL             string           `json:"url"`
	URLResolved     string           `json:"url_resolved"`
	Homepage        string           `json:"homepage"`
	Favicon         string           `json:"favicon"`
	Tags            string           `json:"tags"`
	Country         string           `json:"country"`
	CountryCode     string           `json:"countrycode"`
	State           string           `json:"state"`
	Language        string           `json:"language"`
	LanguageCodes   string           `json:"languagecodes"`
	Votes           int              `json:"votes"`
	Codec           string           `json:"codec"`
	Bitrate         int              `json:"bitrate"`
	LastCheckOK     int              `json:"lastcheckok"`
	LastCheckTime   RadioBrowserTime `json:"lastchecktime"`
	LastCheckOKTime RadioBrowserTime `json:"lastcheckoktime"`
	ClickTimestamp  RadioBrowserTime `json:"clicktimestamp"`
	ClickCount      int              `json:"clickcount"`
	ClickTrend      int              `json:"clicktrend"`
	GeoLat          float64          `json:"geo_lat"`
	GeoLong         float64          `json:"geo_long"`
	HasExtendedInfo bool             `json:"has_extended_info"`
}

// Country represents a country with station count
type Country struct {
	Name         string `json:"name"`
	ISO3166_1    string `json:"iso_3166_1"`
	StationCount int    `json:"stationcount"`
}

// Language represents a language with station count
type Language struct {
	Name         string `json:"name"`
	ISO639       string `json:"iso_639"`
	StationCount int    `json:"stationcount"`
}

// Tag represents a tag/genre with station count
type Tag struct {
	Name         string `json:"name"`
	StationCount int    `json:"stationcount"`
}

// Codec represents an audio codec with station count
type Codec struct {
	Name         string `json:"name"`
	StationCount int    `json:"stationcount"`
}

// SearchParams contains parameters for station search
type SearchParams struct {
	Name        string
	Country     string
	CountryCode string
	State       string
	Language    string
	Tag         string
	Codec       string
	Order       string // name, votes, clickcount, bitrate, codec
	Reverse     bool
	Offset      int
	Limit       int
	HideBroken  bool
}

// ClickResponse represents the response from a click tracking call
type ClickResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

// VoteResponse represents the response from a vote call
type VoteResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

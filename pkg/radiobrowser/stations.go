package radiobrowser

import (
	"fmt"
	"net/url"
	"strconv"
)

// SearchStations performs an advanced search with multiple filter criteria
func (c *Client) SearchStations(params SearchParams) ([]Station, error) {
	endpoint := "/json/stations/search"

	// Build query parameters
	query := url.Values{}

	if params.Name != "" {
		query.Set("name", params.Name)
	}
	if params.Country != "" {
		query.Set("country", params.Country)
	}
	if params.CountryCode != "" {
		query.Set("countrycode", params.CountryCode)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.Language != "" {
		query.Set("language", params.Language)
	}
	if params.Tag != "" {
		query.Set("tag", params.Tag)
	}
	if params.Codec != "" {
		query.Set("codec", params.Codec)
	}
	if params.Order != "" {
		query.Set("order", params.Order)
	}
	if params.Reverse {
		query.Set("reverse", "true")
	}
	if params.Offset > 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.HideBroken {
		query.Set("hidebroken", "true")
	}

	if len(query) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, query.Encode())
	}

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := unmarshalJSON(data, &stations); err != nil {
		return nil, err
	}

	return stations, nil
}

// SearchByName searches for stations by name
func (c *Client) SearchByName(name string) ([]Station, error) {
	endpoint := fmt.Sprintf("/json/stations/byname/%s", url.PathEscape(name))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := unmarshalJSON(data, &stations); err != nil {
		return nil, err
	}

	return stations, nil
}

// SearchByCountry searches for stations by country
func (c *Client) SearchByCountry(country string) ([]Station, error) {
	endpoint := fmt.Sprintf("/json/stations/bycountry/%s", url.PathEscape(country))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := unmarshalJSON(data, &stations); err != nil {
		return nil, err
	}

	return stations, nil
}

// SearchByTag searches for stations by tag/genre
func (c *Client) SearchByTag(tag string) ([]Station, error) {
	endpoint := fmt.Sprintf("/json/stations/bytag/%s", url.PathEscape(tag))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := unmarshalJSON(data, &stations); err != nil {
		return nil, err
	}

	return stations, nil
}

// SearchByLanguage searches for stations by language
func (c *Client) SearchByLanguage(language string) ([]Station, error) {
	endpoint := fmt.Sprintf("/json/stations/bylanguage/%s", url.PathEscape(language))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := unmarshalJSON(data, &stations); err != nil {
		return nil, err
	}

	return stations, nil
}

// GetStationByUUID retrieves a specific station by its UUID
func (c *Client) GetStationByUUID(uuid string) (*Station, error) {
	endpoint := fmt.Sprintf("/json/stations/byuuid/%s", url.PathEscape(uuid))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := unmarshalJSON(data, &stations); err != nil {
		return nil, err
	}

	if len(stations) == 0 {
		return nil, fmt.Errorf("station not found")
	}

	return &stations[0], nil
}

// SearchByUUIDs searches for stations by their UUIDs
func (c *Client) SearchByUUIDs(uuids []string) ([]Station, error) {
	if len(uuids) == 0 {
		return []Station{}, nil
	}

	// Fetch each station individually
	// This is reliable and works for a reasonable number of favorites
	stations := make([]Station, 0, len(uuids))

	for _, uuid := range uuids {
		station, err := c.GetStationByUUID(uuid)
		if err != nil {
			// Skip stations that can't be found (may have been removed from API)
			continue
		}
		stations = append(stations, *station)
	}

	return stations, nil
}

// GetTopStations retrieves the top stations by vote count
func (c *Client) GetTopStations(limit int) ([]Station, error) {
	params := SearchParams{
		Order:      "votes",
		Reverse:    true,
		Limit:      limit,
		HideBroken: true,
	}

	return c.SearchStations(params)
}

// GetPopularStations retrieves popular stations by click count
func (c *Client) GetPopularStations(limit int) ([]Station, error) {
	params := SearchParams{
		Order:      "clickcount",
		Reverse:    true,
		Limit:      limit,
		HideBroken: true,
	}

	return c.SearchStations(params)
}

// GetCountries retrieves all countries with station counts
func (c *Client) GetCountries() ([]Country, error) {
	endpoint := "/json/countries"

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var countries []Country
	if err := unmarshalJSON(data, &countries); err != nil {
		return nil, err
	}

	return countries, nil
}

// GetLanguages retrieves all languages with station counts
func (c *Client) GetLanguages() ([]Language, error) {
	endpoint := "/json/languages"

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var languages []Language
	if err := unmarshalJSON(data, &languages); err != nil {
		return nil, err
	}

	return languages, nil
}

// GetTags retrieves all tags/genres with station counts
func (c *Client) GetTags() ([]Tag, error) {
	endpoint := "/json/tags"

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	if err := unmarshalJSON(data, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetCodecs retrieves all codecs with station counts
func (c *Client) GetCodecs() ([]Codec, error) {
	endpoint := "/json/codecs"

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var codecs []Codec
	if err := unmarshalJSON(data, &codecs); err != nil {
		return nil, err
	}

	return codecs, nil
}

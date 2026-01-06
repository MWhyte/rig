package radiobrowser

import (
	"fmt"
	"net/url"
)

// TrackClick registers a click/play event for a station.
// This should be called whenever a user starts playing a station.
// The API counts only one click per IP per station per day.
func (c *Client) TrackClick(stationUUID string) (*ClickResponse, error) {
	endpoint := fmt.Sprintf("/json/url/%s", url.PathEscape(stationUUID))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var response ClickResponse
	if err := unmarshalJSON(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Vote casts a vote for a station.
// The same IP can only vote for the same station once every 10 minutes.
func (c *Client) Vote(stationUUID string) (*VoteResponse, error) {
	endpoint := fmt.Sprintf("/json/vote/%s", url.PathEscape(stationUUID))

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var response VoteResponse
	if err := unmarshalJSON(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

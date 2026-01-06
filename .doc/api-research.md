# Radio Browser API Research

## Overview

Radio Browser is a free, open-source community radio station database with a public API. It provides access to thousands of internet radio stations worldwide with comprehensive metadata.

## Radio Browser API

### Key Features

- **Free & Open Source**: No API keys or authentication required
- **Multiple Output Formats**: JSON, XML, CSV, M3U, PLS, XSPF, TTL
- **Comprehensive Data**: Station name, URL, genre, country, language, bitrate, codec, etc.
- **Community-Driven**: Users can add stations, vote, and track clicks
- **High Availability**: Multiple server mirrors with DNS-based discovery

### Server Discovery

Instead of hardcoding a single server, the API recommends:
1. DNS lookup of `all.api.radio-browser.info` to get all available servers
2. Randomize the server list
3. Use the first server for requests
4. Fall back to others if needed

Example servers:
- `https://de1.api.radio-browser.info`
- `https://nl1.api.radio-browser.info`
- `https://at1.api.radio-browser.info`

### Important Endpoints

#### Search Stations
- `/json/stations/byname/{searchterm}` - Search by station name
- `/json/stations/bycountry/{country}` - Search by country
- `/json/stations/bytag/{tag}` - Search by tag/genre
- `/json/stations/bylanguage/{language}` - Search by language
- `/json/stations/byuuid/{uuid}` - Get specific station
- `/json/stations/search` - Advanced search with multiple filters

#### List Metadata
- `/json/countries` - All countries with station counts
- `/json/languages` - All languages with station counts
- `/json/tags` - All tags/genres with station counts
- `/json/codecs` - All audio codecs

#### Station Interaction
- `/json/url/{stationuuid}` - Track click (call when user starts playback)
- `/json/vote/{stationuuid}` - Vote for station (once per 10min per IP)

### API Requirements

1. **User-Agent Header**: Must send descriptive User-Agent (e.g., "rig.fm/1.0")
2. **Click Tracking**: Should call `/json/url/{uuid}` when user plays a station
3. **Rate Limiting**: Voting limited to once per 10 minutes per IP per station

### Response Fields

Key station fields:
- `stationuuid` - Unique identifier
- `name` - Station name
- `url`, `url_resolved` - Stream URLs
- `homepage` - Station website
- `favicon` - Station logo URL
- `tags` - Comma-separated genres
- `country`, `countrycode` - Location
- `language`, `languagecodes` - Languages
- `codec` - Audio format (MP3, AAC, etc.)
- `bitrate` - Stream quality in kbps
- `votes` - User votes
- `clickcount` - Play count
- `lastcheckok` - Stream is currently working (1/0)
- `lastchecktime` - Last connectivity check

## goradios SDK Evaluation

**Repository**: https://gitlab.com/AgentNemo/goradios
**License**: MIT
**Last Updated**: February 2021 (abandoned)

### What It Provides

- Basic station search (by name, country, tag, language, codec)
- List operations (countries, states, languages, tags, codecs)
- Struct definitions for Station, Country, Language, etc.
- JSON unmarshaling helpers

### Limitations

1. **Incomplete**: Only ~60-70% of API implemented
2. **Hardcoded Server**: Uses `de1.api.radio-browser.info` instead of DNS discovery
3. **Missing Features**:
   - No voting support
   - No click tracking
   - No advanced search
   - No server failover
4. **No Maintenance**: Last commit in 2021, no tests, marked as TODO
5. **Poor Practices**: Doesn't follow API recommendations

### Example Usage

```go
import "gitlab.com/AgentNemo/goradios"

// Search stations by name
stations, err := goradios.FetchStations(goradios.Name, "jazz")

// Get countries
countries, err := goradios.FetchCountries()
```

## Decision: Build Our Own Client

**Recommendation**: Write a custom, lightweight API client for rig.fm

### Rationale

1. **Better Practices**: Implement DNS-based server discovery properly
2. **Completeness**: Add all features we need (voting, click tracking)
3. **Maintainability**: We control the code and can fix/extend it
4. **Learning**: Better understanding of the API
5. **Size**: The API is simple enough that a custom client isn't much work
6. **Quality**: We can add proper error handling, retries, and tests

### Implementation Plan

Create `pkg/radiobrowser/` with:

1. **client.go**: Core HTTP client with server discovery
2. **stations.go**: Station search and retrieval
3. **metadata.go**: Countries, languages, tags, codecs
4. **interaction.go**: Voting and click tracking
5. **types.go**: Struct definitions for API responses

### Features to Implement

**Phase 1 (MVP)**:
- Server discovery via DNS
- Search stations by name, country, tag
- Get station details
- Click tracking

**Phase 2 (Enhanced)**:
- Advanced search with filters
- List countries/languages/tags
- Voting support
- Caching for metadata

**Phase 3 (Polish)**:
- Retry logic with fallback servers
- Request timeout handling
- Rate limiting
- Comprehensive error messages

## References

- [Radio Browser API Documentation](https://docs.radio-browser.info/)
- [Radio Browser API Home](https://api.radio-browser.info/)
- [goradios Package](https://pkg.go.dev/gitlab.com/AgentNemo/goradios)
- [Radio Browser GitLab](https://gitlab.com/radiobrowser)

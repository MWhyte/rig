# rig.fm

**The most beautiful, feature-rich and user-friendly terminal radio CLI**

Listen to radio from the comfort of your terminal.

## Project Overview

rig.fm is a terminal-based internet radio player that brings the joy of radio listening to the command line. Unlike other terminal radio apps, rig.fm aims to provide an exceptional user experience with a beautiful interface and rich features.

## Tech Stack

- **Language**: Go
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Backend API**: [Radio Browser](https://www.radio-browser.info/users)
- **API Client**: Custom-built client in `pkg/radiobrowser` (goradios evaluated but incomplete)

## Project Goals

1. Create the most beautiful terminal radio interface
2. Provide rich features (search, favorites, history, etc.)
3. Deliver exceptional user experience
4. Make it easy and enjoyable to discover and listen to radio stations

## Architecture

### Directory Structure

```
rig/
├── .claude/          # Claude AI context and project notes
├── .doc/             # Important documentation and plans
├── cmd/
│   ├── rig/          # Main application entry point
│   ├── test-api/     # API client testing tool
│   └── test-player/  # Audio player testing tool
├── pkg/
│   ├── radiobrowser/ # Radio Browser API client
│   └── player/       # Audio player (MPV-based)
└── .github/          # CI/CD and releases (to be added later)
```

## Development Roadmap

### Phase 1: Foundation
- [x] Project scaffolding
- [x] Radio Browser API client (pkg/radiobrowser)
  - [x] DNS-based server discovery
  - [x] Station search (by name, tag, country, language)
  - [x] Advanced search with filters
  - [x] Click tracking and voting
  - [x] Metadata retrieval (countries, languages, tags, codecs)
- [x] Audio playback integration (pkg/player)
  - [x] MPV-based player with IPC control
  - [x] Play, pause, resume, stop controls
  - [x] Volume control
  - [x] Tested with live radio streams
- [x] Beautiful Bubble Tea TUI (pkg/ui)
  - [x] Multi-panel layout (station list, now playing, filters)
  - [x] Tab navigation between sections
  - [x] Focus management with visual indicators
  - [x] Station browser with list view
  - [x] Now playing panel with detailed info
  - [x] Filters panel (country, genre, language)
  - [x] Keyboard controls (tab, ↑/↓, enter, space, +/-, s, r, ?, q)
  - [x] Search/filter stations
  - [x] Help screen
  - [x] Styled with lipgloss borders and colors

### Phase 2: Core Features (MVP Complete!)
- [x] Station browsing (top 50 popular stations)
- [x] Playback controls
- [x] Now playing display
- [x] Volume controls
- [ ] Search by genre/country/name
- [ ] Browse categories

### Phase 3: Enhanced Features
- [ ] Favorites/bookmarks
- [ ] Listening history
- [ ] Keyboard shortcuts
- [ ] Station metadata display

### Phase 4: Polish
- [ ] Themes/customization
- [ ] Configuration file
- [ ] Help/documentation
- [ ] Error handling and recovery

## Notes

- Keep the codebase clean and well-organized
- Follow Go best practices and idioms
- Prioritize user experience in all decisions
- Make it fast, responsive, and delightful to use

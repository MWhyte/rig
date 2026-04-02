# 📻 rig.fm

**Internet radio in your terminal. Beautiful, fast, and keyboard-driven.**

[![Release](https://github.com/MWhyte/rig/actions/workflows/release.yml/badge.svg)](https://github.com/MWhyte/rig/actions/workflows/release.yml)
[![Lint](https://github.com/MWhyte/rig/actions/workflows/lint.yml/badge.svg)](https://github.com/MWhyte/rig/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/MWhyte/rig)](https://goreportcard.com/report/github.com/MWhyte/rig)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)

[Install](#installation) | [Features](#features) | [Usage](#usage) | [Contributing](CONTRIBUTING.md) | [License](#license)

## What is rig.fm?

rig.fm is a terminal radio player that lets you browse, search, and listen to thousands of internet radio stations without leaving the command line. Powered by the [Radio Browser](https://www.radio-browser.info/) directory, it gives you access to stations from around the world, all wrapped in a beautiful TUI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

No accounts. No ads. Just radio.

## Features

- 🌍 Browse thousands of radio stations worldwide
- 🔍 Search by genre, country, language, or station name
- 🎨 Beautiful terminal UI with multiple themes
- ⌨️ Keyboard-driven interface for fast navigation
- ⭐ Save your favourite stations
- 🎵 Now playing display with station metadata

### Search and play
<!-- Replace with actual gif -->
![Searching and playing a station](.docs/assets/search.gif)

### Themes
<!-- Replace with actual gif -->
![Switching themes](.docs/assets/themes.gif)

### Favourites
<!-- Replace with actual gif -->
![Managing favourites](.docs/assets/favourites.gif)


## Installation

rig.fm requires [mpv](https://mpv.io/) for audio playback. Homebrew installs it automatically. For other install methods, make sure mpv is available on your system.

### Homebrew (macOS and Linux)

```bash
brew install mwhyte/tap/rig
```

### Download binary

Grab the latest release for your platform from the [releases page](https://github.com/MWhyte/rig/releases).

### Go install

```bash
go install github.com/mrwhyte/rig/cmd/rig@latest
```

Requires Go 1.25 or later.

### Build from source

```bash
git clone https://github.com/MWhyte/rig.git
cd rig
go build -o rig ./cmd/rig
./rig
```

## Usage

```bash
rig
```

### Keyboard controls

| Key | Action |
|-----|--------|
| `Tab` | Switch between panels |
| `Up/Down` | Navigate station list |
| `Enter` | Play selected station |
| `Space` | Pause/Resume |
| `+` / `-` | Volume up/down |
| `s` | Search stations |
| `?` | Help |
| `q` | Quit |

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to get started.

## Sponsors

If you find rig.fm useful, consider sponsoring its development: [github.com/sponsors/MWhyte](https://github.com/sponsors/MWhyte)

## License

GNU Affero General Public License v3.0. See [LICENSE](LICENSE) for details.

# rig.fm

> The most beautiful, feature-rich and user-friendly terminal radio CLI

Listen to internet radio from the comfort of your terminal.

## Features

- Browse thousands of radio stations worldwide
- Search by genre, country, language, or station name
- Beautiful terminal UI built with Bubble Tea
- Save favourite stations
- Volume control and sleep timer
- Multiple themes
- Keyboard-driven interface

## Installation

```bash
go install github.com/mrwhyte/rig/cmd/rig@latest
```

## Usage

```bash
rig
```

## Development

### Prerequisites

- Go 1.22 or later

### Building from source

```bash
git clone https://github.com/mrwhyte/rig.git
cd rig
go build -o rig ./cmd/rig
./rig
```

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Radio Browser API](https://www.radio-browser.info/) - Radio station database


## Sponsors

If you find rig.fm useful, consider sponsoring its development: [github.com/sponsors/mrwhyte](https://github.com/sponsors/mrwhyte)

## License

GNU Affero General Public License v3.0 — see [LICENSE](LICENSE) for details.

Free to use and build on, provided derivative works are also open source and non-commercial. Commercial use requires explicit permission from the author.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

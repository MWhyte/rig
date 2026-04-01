# Contributing to rig.fm

Thanks for your interest in contributing to rig.fm. This guide will help you get set up and ready to contribute.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25 or later
- [mpv](https://mpv.io/) for audio playback
- A terminal emulator with true color support (recommended)

## Getting started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/rig.git
   cd rig
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build and run:
   ```bash
   go build -o rig ./cmd/rig
   ./rig
   ```

## Project structure

```
rig/
├── cmd/
│   └── rig/          # Application entry point
├── pkg/
│   ├── radiobrowser/ # Radio Browser API client
│   ├── player/       # Audio player (mpv-based)
│   └── ui/           # Bubble Tea TUI components
└── .github/          # CI/CD workflows
```

## Making changes

1. Create a branch for your work:
   ```bash
   git checkout -b your-feature-name
   ```
2. Make your changes
3. Test that the app builds and runs:
   ```bash
   go build -o rig ./cmd/rig
   ./rig
   ```
4. Commit your changes with a clear message
5. Push to your fork and open a pull request

## Pull requests

- Keep PRs focused on a single change
- Provide a clear description of what the PR does and why
- Include screenshots or recordings for UI changes
- Make sure the project builds without errors

## Reporting bugs

Open an issue with:
- Steps to reproduce
- Expected vs actual behavior
- Your OS and terminal emulator

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions small and focused
- Prefer clear names over comments

## Questions?

Open an issue or start a discussion. We're happy to help.

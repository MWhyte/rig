package ui

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrwhyte/rig/pkg/identifier"
)

// identifyDuration is how much audio we capture for one recognition attempt.
// Long enough for reliable matches, short enough not to bore the user.
const identifyDuration = 12 * time.Second

// identifyOverallTimeout caps the full operation including network latency
// and the Shazam rate limiter window.
const identifyOverallTimeout = 45 * time.Second

// identifyResultMsg carries the result of an identification attempt.
type identifyResultMsg struct {
	track *identifier.Track
	err   error
}

// isIdentifying reports whether the identify modal is in its loading state,
// i.e. an identification is in flight and we haven't received a result yet.
func (m *Model) isIdentifying() bool {
	return m.showIdentifyModal && m.identifyTrack == nil && m.identifyErr == nil
}

// startIdentify launches the async identification command. The capture,
// fingerprint, and Shazam round-trip all happen inside the returned tea.Cmd
// so the UI stays responsive. The created context is stored on the model so
// Esc/Enter can abort the in-flight goroutine instead of leaking it.
func (m *Model) startIdentify() tea.Cmd {
	if m.playing == nil {
		return nil
	}
	streamURL := m.playing.URLResolved
	ctx, cancel := context.WithTimeout(context.Background(), identifyOverallTimeout)
	m.identifyCancel = cancel
	return func() tea.Msg {
		track, err := identifier.IdentifyStreamFor(ctx, streamURL, identifyDuration)
		return identifyResultMsg{track: track, err: err}
	}
}

// openURL opens rawURL in the user's default browser. The scheme is
// validated to be http/https before exec'ing anything.
func openURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported url scheme: %q", parsed.Scheme)
	}

	ctx := context.Background()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", rawURL) //nolint:gosec // G204: URL scheme validated above
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", rawURL) //nolint:gosec // G204: URL scheme validated above
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", rawURL) //nolint:gosec // G204: URL scheme validated above
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// handleIdentifyModalInput handles keyboard input while the identify modal
// is showing — whether the spinner is running or a result is on screen.
func (m *Model) handleIdentifyModalInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyCtrlC:
		m.stopPlayback()
		return m, tea.Quit

	case keyEsc, keyEnter:
		m.resetIdentifyState()
		return m, nil

	case keyOpen:
		if m.identifyTrack != nil && m.identifyTrack.ShazamURL != "" {
			if err := openURL(m.identifyTrack.ShazamURL); err != nil {
				m.identifyErr = fmt.Errorf("open url: %w", err)
			}
		}
		return m, nil
	}
	return m, nil
}

// resetIdentifyState clears identify-related state and cancels any in-flight
// goroutine. The cancellation propagates to the HTTP request, the MP3
// decoder, and the Shazam rate limiter, so dismissed identifications stop
// consuming network and memory immediately.
func (m *Model) resetIdentifyState() {
	if m.identifyCancel != nil {
		m.identifyCancel()
		m.identifyCancel = nil
	}
	m.showIdentifyModal = false
	m.identifyTrack = nil
	m.identifyErr = nil
}

// renderIdentifyModal renders a centred modal showing either the spinner
// (while identifying) or the result (after completion).
func (m *Model) renderIdentifyModal() string {
	const modalWidth = 56
	const modalHeight = 11

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorTitle).
		Padding(0, 1).
		Render("♪ Identify Track")

	var content string
	switch {
	case m.isIdentifying():
		content = m.renderIdentifySpinner()
	case m.identifyErr != nil:
		content = m.renderIdentifyError()
	case m.identifyTrack != nil:
		content = m.renderIdentifyResult()
	default:
		content = "\n  (nothing to display)\n"
	}

	panel := lipgloss.JoinVertical(lipgloss.Left, title, content)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight).
		Render(panel)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m *Model) renderIdentifySpinner() string {
	return fmt.Sprintf(
		"\n  %s  %s\n\n  %s\n",
		lipgloss.NewStyle().Foreground(colorAccent).Render(m.identifySpinner.View()),
		lipgloss.NewStyle().Foreground(colorTitle).Render("Listening..."),
		lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("This takes about 12-15 seconds. esc to cancel."),
	)
}

func (m *Model) renderIdentifyError() string {
	// For codec mismatches we deliberately hide the technical decoder
	// string ("only layer3 ... want 1 got 3" etc) and just show the
	// friendly headline.
	if errors.Is(m.identifyErr, identifier.ErrUnsupportedCodec) {
		return fmt.Sprintf(
			"\n  %s\n\n  %s",
			lipgloss.NewStyle().Foreground(colorWarning).Render("Sorry, only MP3 streams are currently supported"),
			lipgloss.NewStyle().Foreground(colorDim).Render("enter/esc to close"),
		)
	}

	headline := "Couldn't identify the track"
	if errors.Is(m.identifyErr, identifier.ErrNoMatch) {
		headline = "No match found"
	}
	return fmt.Sprintf(
		"\n  %s\n\n  %s\n\n  %s",
		lipgloss.NewStyle().Foreground(colorWarning).Render(headline),
		lipgloss.NewStyle().Foreground(colorMuted).Render(m.identifyErr.Error()),
		lipgloss.NewStyle().Foreground(colorDim).Render("enter/esc to close"),
	)
}

func (m *Model) renderIdentifyResult() string {
	t := m.identifyTrack
	var b strings.Builder
	b.WriteString("\n  ")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTitle).Render(t.Title))
	b.WriteString("\n  ")
	b.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render(t.Artist))
	if t.Album != "" {
		b.WriteString("\n  ")
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(t.Album))
		if t.Year != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render(" · " + t.Year))
		}
	}
	b.WriteString("\n\n  ")
	if t.ShazamURL != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(colorDim).
			Render("press o to open in Shazam · enter/esc to close"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("enter/esc to close"))
	}
	return b.String()
}

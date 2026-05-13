// Package identifier identifies tracks playing on internet radio streams by
// tapping the stream's audio, computing a Shazam fingerprint locally, and
// querying Shazam's recognition endpoint.
package identifier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrwhyte/rig/pkg/identifier/shazam"
)

// DefaultSampleSeconds is the amount of audio captured for one recognition
// attempt. Shazam typically resolves a track from 8-12 seconds of audio.
const DefaultSampleSeconds = 12

// Track is what a successful identification returns to a caller.
type Track struct {
	Title     string
	Artist    string
	Album     string
	Year      string
	ShazamURL string // canonical Shazam track URL, empty if Shazam returned no key
	AppleID   string // Apple Music track ID, empty if absent
}

// ErrNoMatch is returned when Shazam responds successfully but cannot identify
// the audio (silence, talk, an unrecognised song).
var ErrNoMatch = errors.New("no shazam match")

// IdentifyStream captures DefaultSampleSeconds of audio from streamURL,
// fingerprints it, and asks Shazam to identify the track. It returns
// ErrNoMatch if Shazam responds with no matches.
func IdentifyStream(ctx context.Context, streamURL string) (*Track, error) {
	return IdentifyStreamFor(ctx, streamURL, DefaultSampleSeconds*time.Second)
}

// IdentifyStreamFor is the variant of IdentifyStream that lets the caller pick
// the capture duration. Shorter durations are faster but less accurate.
func IdentifyStreamFor(ctx context.Context, streamURL string, duration time.Duration) (*Track, error) {
	samples, sampleRate, err := CaptureMonoSamples(ctx, streamURL, duration)
	if err != nil {
		return nil, fmt.Errorf("capture: %w", err)
	}

	resampled := Resample(samples, sampleRate, FingerprintSampleRate)
	sig := shazam.ComputeSignature(FingerprintSampleRate, resampled)
	result, err := shazam.Identify(ctx, sig)
	if err != nil {
		return nil, fmt.Errorf("identify: %w", err)
	}
	if !result.Found {
		return nil, ErrNoMatch
	}

	return &Track{
		Title:     result.Title,
		Artist:    result.Artist,
		Album:     result.Album,
		Year:      result.Year,
		ShazamURL: result.ShazamURL(),
		AppleID:   result.AppleID,
	}, nil
}

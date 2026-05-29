// Command test-shazam taps a live radio stream URL, fingerprints ~12 seconds
// of audio, and asks Shazam to identify the track. It exists so the
// identification pipeline can be smoke-tested end-to-end without booting the
// TUI.
//
// Usage:
//
//	go run ./cmd/test-shazam <stream_url> [duration_seconds]
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mrwhyte/rig/pkg/identifier"
	"github.com/mrwhyte/rig/pkg/identifier/shazam"
)

func main() {
	os.Exit(run())
}

func run() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <stream_url> [duration_seconds]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return 2
	}

	streamURL := flag.Arg(0)
	duration := identifier.DefaultSampleSeconds * time.Second
	if flag.NArg() >= 2 {
		secs, err := strconv.Atoi(flag.Arg(1))
		if err != nil || secs <= 0 {
			fmt.Fprintf(os.Stderr, "invalid duration_seconds: %q\n", flag.Arg(1))
			return 2
		}
		duration = time.Duration(secs) * time.Second
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Generous overall timeout: capture + network round-trip + Shazam rate limiter.
	ctx, cancelTimeout := context.WithTimeout(ctx, duration+30*time.Second)
	defer cancelTimeout()

	fmt.Fprintf(os.Stderr, "tapping %s for %s...\n", streamURL, duration)

	captureStart := time.Now()
	samples, sampleRate, err := identifier.CaptureMonoSamples(ctx, streamURL, duration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "capture error: %v\n", err)
		return 1
	}
	captureWall := time.Since(captureStart)
	audioSeconds := float64(len(samples)) / float64(sampleRate)
	fmt.Fprintf(
		os.Stderr,
		"captured %d samples at %d Hz (%.2fs of audio, wall time %s)\n",
		len(samples), sampleRate, audioSeconds, captureWall.Round(time.Millisecond),
	)
	if audioSeconds < float64(duration/time.Second)*0.9 {
		fmt.Fprintf(os.Stderr, "warning: captured less audio than requested — stream may have closed early\n")
	}

	resampled := identifier.Resample(samples, sampleRate, identifier.FingerprintSampleRate)
	fmt.Fprintf(
		os.Stderr,
		"resampled to %d samples at %d Hz\n",
		len(resampled), identifier.FingerprintSampleRate,
	)

	identifyStart := time.Now()
	sig := shazam.ComputeSignature(identifier.FingerprintSampleRate, resampled)
	result, err := shazam.Identify(ctx, sig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "identify error: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "shazam round trip %s\n", time.Since(identifyStart).Round(time.Millisecond))

	if !result.Found {
		fmt.Fprintln(os.Stderr, "no match")
		return 1
	}

	fmt.Printf("Title:   %s\n", result.Title)
	fmt.Printf("Artist:  %s\n", result.Artist)
	if result.Album != "" {
		fmt.Printf("Album:   %s\n", result.Album)
	}
	if result.Year != "" {
		fmt.Printf("Year:    %s\n", result.Year)
	}
	if u := result.ShazamURL(); u != "" {
		fmt.Printf("Shazam:  %s\n", u)
	}
	if result.AppleID != "" {
		fmt.Printf("AppleID: %s\n", result.AppleID)
	}
	return 0
}

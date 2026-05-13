package identifier

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/hajimehoshi/go-mp3"
)

// ErrUnsupportedCodec is returned when the stream's Content-Type indicates an
// audio codec other than MP3. Track identification currently only supports
// MP3 streams; AAC and other codecs would need a different decoder.
var ErrUnsupportedCodec = errors.New("unsupported audio codec")

// streamReadBufBytes is the size of the chunk we read at a time from the
// decoder. go-mp3 emits 16-bit stereo little-endian PCM (4 bytes per frame),
// so 4096 is exactly 1024 frames per read.
const streamReadBufBytes = 4096

// CaptureMonoSamples opens streamURL, decodes the next `duration` worth of
// MP3 audio, downmixes to mono and returns the samples normalised to [-1, 1]
// alongside the decoder's native sample rate.
//
// The samples are NOT resampled here; if you intend to feed them to
// shazam.ComputeSignature, run them through Resample first so the rate
// matches what Shazam's matcher expects (FingerprintSampleRate).
func CaptureMonoSamples(ctx context.Context, streamURL string, duration time.Duration) (samples []float64, sampleRate int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, http.NoBody)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	// Ask the server not to interleave ICY metadata into the audio stream —
	// most respect this and it keeps the MP3 frame boundaries clean.
	req.Header.Set("Icy-MetaData", "0")
	req.Header.Set("User-Agent", "rig/identifier")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("open stream: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("stream returned %s", resp.Status)
	}

	// Reject obvious non-MP3 codecs up front so the user gets a clear
	// "unsupported codec" error instead of a confusing "EOF" from go-mp3.
	if ct := resp.Header.Get("Content-Type"); ct != "" && !isMP3ContentType(ct) {
		return nil, 0, fmt.Errorf("%w: %s (only MP3 is supported)", ErrUnsupportedCodec, ct)
	}

	// Live Shoutcast/Icecast streams hand us bytes mid-frame on connect,
	// which makes go-mp3 fail with errors like "is_pos was too big". Wrap
	// the body in a reader that scans forward to the first valid MP3 frame
	// header before passing data through.
	synced, err := newMP3SyncReader(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("sync mp3 stream: %w", err)
	}

	dec, err := mp3.NewDecoder(synced)
	if err != nil {
		// Errors from mp3.NewDecoder almost always mean the stream is not
		// an MP3 we can handle — wrong codec (AAC/Ogg), MPEG Layer I/II,
		// or unrecognisable bytes after our sync scan. Surface them all as
		// ErrUnsupportedCodec so the UI shows a clean message rather than
		// a technical decoder error string.
		return nil, 0, fmt.Errorf("%w: %s", ErrUnsupportedCodec, err.Error())
	}

	sampleRate = dec.SampleRate()
	if sampleRate <= 0 || sampleRate > 192000 {
		return nil, 0, fmt.Errorf("implausible stream sample rate: %d Hz", sampleRate)
	}

	framesNeeded := int(duration.Seconds() * float64(sampleRate))
	samples = make([]float64, 0, framesNeeded)

	buf := make([]byte, streamReadBufBytes)
	for len(samples) < framesNeeded {
		n, readErr := dec.Read(buf)
		if n > 0 {
			n -= n % 4 // align to whole stereo frames
			for i := 0; i+3 < n; i += 4 {
				left := int16(binary.LittleEndian.Uint16(buf[i : i+2]))    //nolint:gosec // G115: PCM frames are intentionally signed
				right := int16(binary.LittleEndian.Uint16(buf[i+2 : i+4])) //nolint:gosec // G115: PCM frames are intentionally signed
				samples = append(samples, (float64(left)+float64(right))/2/32768)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return nil, 0, fmt.Errorf("read mp3: %w", readErr)
		}
	}

	if len(samples) > framesNeeded {
		samples = samples[:framesNeeded]
	}
	if len(samples) == 0 {
		return nil, 0, errors.New("no audio decoded")
	}

	return samples, sampleRate, nil
}

// FingerprintSampleRate is the rate Shazam's matcher is tuned for. Feeding
// the algorithm at higher rates (e.g. 44100 Hz directly) is technically valid
// per the signature wire format, but produces sparse peaks in Shazam's
// 250-5500 Hz bands and fails to match real-world tracks. Resampling to
// 16 kHz before fingerprinting is what every working Shazam client does.
const FingerprintSampleRate = 16000

// Resample converts audio samples from srcRate to dstRate. For downsampling
// it averages over the source window to apply a crude anti-aliasing low-pass
// filter; for upsampling it uses linear interpolation. Quality is adequate
// for audio fingerprinting, where peak topology matters more than fidelity.
//
// If srcRate == dstRate, the input slice is returned unchanged.
func Resample(in []float64, srcRate, dstRate int) []float64 {
	if srcRate == dstRate || len(in) == 0 {
		return in
	}
	ratio := float64(srcRate) / float64(dstRate)
	outLen := int(float64(len(in)) / ratio)
	if outLen <= 0 {
		return nil
	}
	out := make([]float64, outLen)

	if ratio > 1 {
		// Downsampling: boxcar-averaged decimation.
		window := int(ratio + 0.5)
		if window < 2 {
			window = 2
		}
		for i := range out {
			srcStart := int(float64(i) * ratio)
			if srcStart >= len(in) {
				break
			}
			srcEnd := srcStart + window
			if srcEnd > len(in) {
				srcEnd = len(in)
			}
			sum := 0.0
			for j := srcStart; j < srcEnd; j++ {
				sum += in[j]
			}
			out[i] = sum / float64(srcEnd-srcStart)
		}
		return out
	}

	// Upsampling: linear interpolation.
	for i := range out {
		srcF := float64(i) * ratio
		srcI := int(srcF)
		frac := srcF - float64(srcI)
		switch {
		case srcI+1 < len(in):
			out[i] = in[srcI]*(1-frac) + in[srcI+1]*frac
		case srcI < len(in):
			out[i] = in[srcI]
		}
	}
	return out
}

// mp3SyncScanLimit is how many bytes we'll scan from the start of a stream
// looking for the first valid MP3 frame header before giving up.
const mp3SyncScanLimit = 1 << 16 // 64 KiB

// mp3SyncReader wraps an io.Reader and skips past any leading bytes that
// don't begin a valid MP3 frame header. Used to recover from connecting to
// a live Shoutcast/Icecast stream mid-frame.
type mp3SyncReader struct {
	br *bufio.Reader
}

// isMP3ContentType reports whether the HTTP Content-Type advertises an MP3
// stream. Servers vary in exact spelling (audio/mpeg, audio/mp3, audio/MPA).
func isMP3ContentType(ct string) bool {
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		mt = strings.ToLower(strings.TrimSpace(ct))
	}
	switch mt {
	case "audio/mpeg", "audio/mp3", "audio/x-mpeg", "audio/mpa", "audio/mpa-robust":
		return true
	}
	return false
}

// isValidMP3Header returns true if h looks like a real MP3 frame header.
// Loose sync (byte 0 == 0xFF, byte 1 top 3 bits set) gets a lot of false
// positives in ICY metadata and ID3 tags; we additionally reject the
// reserved version, reserved layer, "bad" bitrate index, and reserved
// sample-rate index, which never appear in real audio.
func isValidMP3Header(h []byte) bool {
	if len(h) < 3 {
		return false
	}
	if h[0] != 0xFF {
		return false
	}
	if h[1]&0xE0 != 0xE0 {
		return false
	}
	if h[1]&0x18 == 0x08 { // version bits == 01: reserved
		return false
	}
	if h[1]&0x06 == 0x00 { // layer bits == 00: reserved
		return false
	}
	if h[2]&0xF0 == 0xF0 { // bitrate index == 1111: bad
		return false
	}
	if h[2]&0x0C == 0x0C { // sample-rate index == 11: reserved
		return false
	}
	return true
}

// newMP3SyncReader scans bytes from r until it finds an MP3 frame header
// that passes isValidMP3Header, then returns a reader positioned at that
// byte. Bytes are inspected via Peek so they stay buffered for the decoder.
func newMP3SyncReader(r io.Reader) (*mp3SyncReader, error) {
	br := bufio.NewReaderSize(r, 4096)
	for scanned := 0; scanned < mp3SyncScanLimit; scanned++ {
		head, err := br.Peek(3)
		if err != nil {
			return nil, fmt.Errorf("peek while seeking sync: %w", err)
		}
		if isValidMP3Header(head) {
			return &mp3SyncReader{br: br}, nil
		}
		if _, err := br.ReadByte(); err != nil {
			return nil, fmt.Errorf("discard while seeking sync: %w", err)
		}
	}
	return nil, fmt.Errorf("no MP3 sync in first %d bytes", mp3SyncScanLimit)
}

// Read implements io.Reader by delegating to the buffered reader.
func (s *mp3SyncReader) Read(p []byte) (int, error) {
	return s.br.Read(p)
}

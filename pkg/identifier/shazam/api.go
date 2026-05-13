// This file is vendored from https://github.com/lukechampine/barbershop
// Original copyright (c) 2024 Luke Champine, MIT licensed (see LICENSE).
//
// Modifications for rig:
//   - Identify now takes a context.Context for cancellation.
//   - Result exposes the Shazam track Key so callers can build a Shazam URL.
//   - Replaced math/rand with math/rand/v2 for the user-agent picker.
//   - Replaced goto-based retry with a bounded loop.
//   - Added http.NewRequestWithContext and explicit error handling throughout.

package shazam

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// debugEnabled returns true when RIG_SHAZAM_DEBUG=1 in the environment. When
// enabled, postSignature dumps the raw request body and response body to
// stderr so we can diagnose why Shazam is returning no matches.
func debugEnabled() bool {
	return os.Getenv("RIG_SHAZAM_DEBUG") == "1"
}

// Shared rate limiter so we don't hammer Shazam and earn a ban.
var throttle = rate.NewLimiter(rate.Every(3*time.Second), 1)

// Result contains the outcome of attempting to identify a song.
type Result struct {
	Found   bool
	Artist  string
	Title   string
	Album   string
	Year    string
	AppleID string
	Key     string // Shazam track key, used to build https://www.shazam.com/track/<Key>
}

// ShazamURL returns the canonical Shazam track URL, or an empty string if the
// result does not carry a track key.
func (r Result) ShazamURL() string {
	if r.Key == "" {
		return ""
	}
	return "https://www.shazam.com/track/" + r.Key
}

const (
	endpointHost  = "amp.shazam.com"
	endpointPath  = "/discovery/v5/en/US/android/-/tag"
	endpointQuery = "?sync=true&webv3=true&sampling=true&connected=&shazamapiversion=v3&sharehub=true&video=v3"
)

// Identify sends the given signature to Shazam and returns the recognised
// track, or a Result with Found == false if no match was made.
func Identify(ctx context.Context, sig Signature) (Result, error) {
	body, err := buildRequestBody(sig)
	if err != nil {
		return Result{}, fmt.Errorf("build request body: %w", err)
	}

	reqURL := fmt.Sprintf(
		"http://%s%s/%s/%s%s",
		endpointHost,
		endpointPath,
		strings.ToUpper(uuid.NewString()),
		uuid.NewString(),
		endpointQuery,
	)

	respData, err := postSignature(ctx, reqURL, body)
	if err != nil {
		return Result{}, err
	}
	if len(respData.Matches) == 0 {
		return Result{Found: false}, nil
	}

	album, year := "", ""
	for _, section := range respData.Track.Sections {
		for _, meta := range section.Metadata {
			switch meta.Title {
			case "Album":
				album = meta.Text
			case "Released", "Sortie":
				year = meta.Text
			}
		}
	}
	appleID := ""
	for _, action := range respData.Track.Hub.Actions {
		if action.Name == "apple" && action.ID != "" {
			appleID = action.ID
			break
		}
	}

	return Result{
		Found:   true,
		Artist:  respData.Track.Subtitle,
		Title:   respData.Track.Title,
		Album:   album,
		Year:    year,
		AppleID: appleID,
		Key:     respData.Track.Key,
	}, nil
}

func buildRequestBody(sig Signature) ([]byte, error) {
	now := time.Now().UnixMilli()
	payload := map[string]any{
		"geolocation": map[string]any{
			"altitude":  300,
			"latitude":  45,
			"longitude": 2,
		},
		"signature": map[string]any{
			"samplems":  sig.numSamples / sig.sampleRate * 1000,
			"timestamp": now,
			"uri":       "data:audio/vnd.shazam.sig;base64," + base64.StdEncoding.EncodeToString(sig.encode()),
		},
		"timestamp": now,
		"timezone":  "Europe/Berlin",
	}
	return json.Marshal(payload)
}

type shazamResponse struct {
	Matches []struct {
		ID            string
		Offset        float64
		TimeSkew      float64
		FrequencySkew float64
	}
	Track struct {
		Title    string
		Subtitle string
		Key      string
		Hub      struct {
			Actions []struct {
				Name string
				ID   string
			}
		}
		Sections []struct {
			Type     string
			Metadata []struct {
				Title string
				Text  string
			}
		}
	}
}

const maxRetries = 3

func postSignature(ctx context.Context, reqURL string, body []byte) (*shazamResponse, error) {
	debug := debugEnabled()
	if debug {
		fmt.Fprintf(os.Stderr, "[shazam] POST %s\n", reqURL)
		fmt.Fprintf(os.Stderr, "[shazam] request body: %s\n", string(body))
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := throttle.Wait(ctx); err != nil {
			return nil, fmt.Errorf("throttle: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		ua := pickUserAgent()
		req.Header.Set("User-Agent", ua)
		req.Header.Set("Content-Language", "en_US")
		req.Header.Set("Content-Type", "application/json")
		if debug {
			fmt.Fprintf(os.Stderr, "[shazam] User-Agent: %s\n", ua)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if debug {
			fmt.Fprintf(os.Stderr, "[shazam] response status: %s\n", resp.Status)
			fmt.Fprintf(os.Stderr, "[shazam] response body (%d bytes): %s\n", len(respBody), string(respBody))
		}
		if readErr != nil {
			return nil, fmt.Errorf("read response: %w", readErr)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			var out shazamResponse
			if err := json.Unmarshal(respBody, &out); err != nil {
				return nil, fmt.Errorf("decode response: %w", err)
			}
			return &out, nil
		case http.StatusTooManyRequests:
			lastErr = fmt.Errorf("rate limited by shazam")
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(3 * time.Second):
			}
		default:
			return nil, fmt.Errorf("bad status: %s (%s)", resp.Status, string(respBody))
		}
	}
	return nil, lastErr
}

func pickUserAgent() string {
	// Picking a UA string is not a security-sensitive decision; math/rand/v2 is fine.
	return userAgents[rand.IntN(len(userAgents))] //nolint:gosec // G404: non-crypto use
}

// userAgents is a pool of Android Dalvik user agents used to look like a
// regular mobile Shazam client. Picked at random per request.
var userAgents = []string{
	"Dalvik/2.1.0 (Linux; U; Android 5.0.2; VS980 4G Build/LRX22G)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SM-T210 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-P905V Build/LMY47X)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; Vodafone Smart Tab 4G Build/KTU84P)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; SM-G360H Build/KTU84P)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0.2; SM-S920L Build/LRX22G)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0; Fire Pro Build/LRX21M)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0; SM-N9005 Build/LRX21V)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G920F Build/MMB29K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SM-G7102 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0; SM-G900F Build/LRX21T)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G928F Build/MMB29K)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-J500FN Build/LMY48B)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; Coolpad 3320A Build/LMY47V)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; SM-J110F Build/KTU84P)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SAMSUNG-SGH-I747 Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SAMSUNG-SM-T337A Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.3; SGH-T999 Build/JSS15J)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; D6603 Build/23.5.A.0.570)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-J700H Build/LMY48B)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC6600LVW Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-N910G Build/LMY47X)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-N910T Build/LMY47X)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; C6903 Build/14.4.A.0.157)",
	"Dalvik/1.6.0 (Linux; U; Android 4.2.2; GT-I9105P Build/JDQ39)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; GT-I9192 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-G531H Build/LMY48B)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; LGMS345 Build/LMY47V)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0.2; HTC One Build/LRX22G)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0.2; LG-D800 Build/LRX22G)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; SM-T113 Build/KTU84P)",
	"Dalvik/1.6.0 (Linux; U; Android 4.2.2; AndyWin Build/JDQ39E)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0; Lenovo A7000-a Build/LRX21M)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; LGL16C Build/KOT49I.L16CV11a)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; GT-I9500 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0.2; SM-A700FD Build/LRX22G)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SM-G130HN Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SM-N9005 Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.1.2; LG-E975T Build/JZO54K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; E1 Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; GT-N5100 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-A310F Build/LMY47X)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-J105H Build/LMY47V)",
	"Dalvik/1.6.0 (Linux; U; Android 4.3; GT-I9305T Build/JSS15J)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; android Build/JDQ39)",
	"Dalvik/1.6.0 (Linux; U; Android 4.2.1; HS-U970 Build/JOP40D)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; SM-T561 Build/KTU84P)",
	"Dalvik/1.6.0 (Linux; U; Android 4.2.2; GT-P3110 Build/JDQ39)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G925T Build/MMB29K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; HUAWEI Y221-U22 Build/HUAWEIY221-U22)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-G530T1 Build/LMY47X)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-G920I Build/LMY47X)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; Vodafone Smart ultra 6 Build/LMY47V)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; XT1080 Build/SU6-7.7)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; ASUS MeMO Pad 7 Build/KTU84P)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SM-G800F Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; GT-N7100 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G925I Build/MMB29K)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; A0001 Build/MMB29X)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1; XT1045 Build/LPB23.13-61)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; LGMS330 Build/LMY47V)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; Z970 Build/KTU84P)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0; SM-N900P Build/LRX21V)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; T1-701u Build/HuaweiMediaPad)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1; HTCD100LVWPP Build/LMY47O)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G935R4 Build/MMB29M)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G930V Build/MMB29M)",
	"Dalvik/2.1.0 (Linux; U; Android 5.0.2; ZTE Blade Q Lux Build/LRX22G)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; GT-I9060I Build/KTU84P)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; LGUS992 Build/MMB29M)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G900P Build/MMB29M)",
	"Dalvik/1.6.0 (Linux; U; Android 4.1.2; SGH-T999L Build/JZO54K)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-N910V Build/LMY47X)",
	"Dalvik/2.1.0 (Linux; U; Android 5.1.1; SM-P601 Build/LMY47X)",
	"Dalvik/1.6.0 (Linux; U; Android 4.2.2; GT-S7272 Build/JDQ39)",
	"Dalvik/1.6.0 (Linux; U; Android 4.3; SAMSUNG-SGH-I747 Build/JSS15J)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-G930F Build/MMB29K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; HTC_PO582 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0; HUAWEI MT7-TL10 Build/HuaweiMT7-TL10)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0; LG-H811 Build/MRA58K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; SM-N7505 Build/KOT49H)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0; LG-H815 Build/MRA58K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.2; LenovoA3300-HV Build/KOT49H)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; SM-G360G Build/KTU84P)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; GT-I9300I Build/KTU84P)",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; SM-J700T Build/MMB29K)",
	"Dalvik/1.6.0 (Linux; U; Android 4.2.2; SM-T217S Build/JDQ39)",
	"Dalvik/1.6.0 (Linux; U; Android 4.4.4; SAMSUNG-SM-N900A Build/KTU84P)",
}

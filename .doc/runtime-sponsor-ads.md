# Runtime Sponsor Ads with Embedded Defaults

## Context
Moving from fully embedded ads to a hybrid model. Three permanent "house" ads are embedded in the binary (mwhyte.dev, radio-browser.info, sponsor CTA). Community sponsor ads are fetched at runtime from this repo's `/ads` directory via GitHub API, loaded in the background so startup isn't blocked. If the fetch fails, the app still runs with the embedded defaults.

## Design

### Two-tier ad system
- **Embedded (permanent):** 3 house ads baked into the binary via `//go:embed`. These always rotate, even offline.
- **Remote (sponsors):** `.txt` files in `/ads` at the repo root, fetched from GitHub API on startup. Merged into the rotation alongside the embedded ads.

### Directory structure
```
rig/
├── pkg/ui/ads/          # Embedded house ads (permanent, 3 files)
│   ├── mwhyte.txt
│   ├── radiobrowser.txt
│   └── sponsor-cta.txt
├── pkg/ui/ads.go        # Embed loader + remote fetch logic
├── ads/                 # Community sponsor ads (fetched at runtime)
│   ├── README.md        # Instructions for sponsors on format/sizing
│   └── *.txt            # Sponsor ad files (added via PR)
```

### Fetch flow
1. App starts → `sponsorAds` initialized with embedded house ads (instant, no network)
2. `Init()` fires `fetchRemoteAds()` as a background command
3. `fetchRemoteAds()` calls GitHub API: `GET /repos/mrwhyte/rig/contents/ads`
4. For each `.txt` file returned, fetch its raw content
5. On success, send `remoteAdsLoadedMsg{ads []string}` back to Update()
6. Update() appends remote ads to `sponsorAds`, rotation continues seamlessly
7. On failure (network error, API error), silently ignore — embedded ads keep running

### Files to modify
- **`pkg/ui/ads.go`** — keep `loadAds()` for embedded, add `fetchRemoteAds()` returning a `tea.Cmd` that hits GitHub API
- **`pkg/ui/model.go`** — add `remoteAdsLoadedMsg` type, handle it in `Update()` to append remote ads to `sponsorAds` and restart rotation if needed, add `fetchRemoteAds()` call to `Init()`

### Files to create
- **`ads/README.md`** — instructions for sponsors on ad format, dimensions, and PR process

### Files to rename
- `pkg/ui/ads/sample1.txt` → keep as the rig.fm branding ad (already exists, user modified)
- `pkg/ui/ads/sample2.txt` → radio-browser.info ad (already exists)
- `pkg/ui/ads/sample3.txt` → sponsor CTA ad (already exists)
- `pkg/ui/ads/sample4.txt` → mwhyte.dev ad (already exists, user modified)
- Remove extras so only 3 permanent ads remain: mwhyte.dev, radio-browser.info, sponsor CTA

### No changes needed
- `pkg/ui/layout.go` — renderSponsorsPanel already works with the `sponsorAds` slice, no changes needed
- Wipe animation, rotation timing — all unchanged

## Verification
- `go build -o rig cmd/rig/main.go`
- Run with internet: should show 3 embedded ads immediately, then remote ads appear in rotation after a few seconds
- Run without internet (or before repo is public): should show only 3 embedded ads, no errors
- Add a test `.txt` file to `/ads` in the repo, verify it appears after restart

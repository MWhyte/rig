package ui

import (
	"embed"
	"strings"
)

//go:embed ads/*.txt
var adsFS embed.FS

// loadAds reads all .txt files from the embedded ads directory.
func loadAds() []string {
	entries, err := adsFS.ReadDir("ads")
	if err != nil {
		return nil
	}

	var ads []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		data, err := adsFS.ReadFile("ads/" + entry.Name())
		if err != nil {
			continue
		}
		content := string(data)
		if len(strings.TrimSpace(content)) > 0 {
			ads = append(ads, content)
		}
	}
	return ads
}

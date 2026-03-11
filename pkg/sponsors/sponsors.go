package sponsors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// gistURL is the raw URL of the public Gist containing sponsors.json.
// Update this after creating the Gist (see setup instructions in CLAUDE.md).
const gistURL = "https://gist.githubusercontent.com/mrwhyte/1c23d0635fc400008be3dcea8f8d068c/raw/rig-sponsors.json"

const cacheTTL = 24 * time.Hour

// Sponsor is a single GitHub Sponsors backer.
type Sponsor struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

// SponsorList is the full payload stored in the Gist and local cache.
type SponsorList struct {
	UpdatedAt time.Time `json:"updated_at"`
	Sponsors  []Sponsor `json:"sponsors"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Load returns sponsors from cache if fresh, otherwise fetches from the Gist.
// Falls back to stale cache when the network is unavailable.
func Load() ([]Sponsor, error) {
	cached, cacheErr := readCache()
	if cacheErr == nil && time.Since(cached.FetchedAt) < cacheTTL {
		return cached.Sponsors, nil
	}

	fetched, err := fetch()
	if err != nil {
		if cacheErr == nil {
			return cached.Sponsors, nil
		}
		return nil, fmt.Errorf("sponsors: %w", err)
	}

	list := &SponsorList{
		UpdatedAt: fetched.UpdatedAt,
		Sponsors:  fetched.Sponsors,
		FetchedAt: time.Now(),
	}
	_ = writeCache(list)
	return list.Sponsors, nil
}

func fetch() (*SponsorList, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(gistURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var list SponsorList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &list, nil
}

func cachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "rig", "sponsors.json"), nil
}

func readCache() (*SponsorList, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list SponsorList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func writeCache(list *SponsorList) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

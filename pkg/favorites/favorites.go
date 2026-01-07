package favorites

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FavoriteStation represents a favorited station
type FavoriteStation struct {
	StationUUID string    `json:"stationuuid"`
	Name        string    `json:"name"`
	URLResolved string    `json:"url_resolved"`
	AddedAt     time.Time `json:"added_at"`
}

// Manager handles favorites persistence
type Manager struct {
	filePath string
	stations map[string]FavoriteStation // Map by UUID for fast lookup
}

// NewManager creates a favorites manager
func NewManager() (*Manager, error) {
	// Get config dir: ~/.config/rig/
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	rigDir := filepath.Join(configDir, "rig")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(rigDir, "favorites.json")

	m := &Manager{
		filePath: filePath,
		stations: make(map[string]FavoriteStation),
	}

	// Load existing favorites
	if err := m.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

// IsFavorite checks if a station is favorited
func (m *Manager) IsFavorite(stationUUID string) bool {
	_, exists := m.stations[stationUUID]
	return exists
}

// Toggle adds or removes a station from favorites
func (m *Manager) Toggle(stationUUID, name, urlResolved string) error {
	if m.IsFavorite(stationUUID) {
		delete(m.stations, stationUUID)
	} else {
		m.stations[stationUUID] = FavoriteStation{
			StationUUID: stationUUID,
			Name:        name,
			URLResolved: urlResolved,
			AddedAt:     time.Now(),
		}
	}

	return m.Save()
}

// GetAll returns all favorite stations
func (m *Manager) GetAll() []FavoriteStation {
	favs := make([]FavoriteStation, 0, len(m.stations))
	for _, station := range m.stations {
		favs = append(favs, station)
	}
	return favs
}

// Load reads favorites from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var stored struct {
		Stations []FavoriteStation `json:"stations"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return err
	}

	// Build map for fast lookup
	m.stations = make(map[string]FavoriteStation)
	for _, station := range stored.Stations {
		m.stations[station.StationUUID] = station
	}

	return nil
}

// Save writes favorites to disk
func (m *Manager) Save() error {
	// Convert map to slice for JSON
	stations := make([]FavoriteStation, 0, len(m.stations))
	for _, station := range m.stations {
		stations = append(stations, station)
	}

	data := struct {
		Stations []FavoriteStation `json:"stations"`
	}{
		Stations: stations,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, jsonData, 0644)
}

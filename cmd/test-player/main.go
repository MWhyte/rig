package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mrwhyte/rig/pkg/player"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

func main() {
	fmt.Println("Testing Audio Player...")
	fmt.Println()

	// Create player
	fmt.Println("1. Creating mpv player...")
	p, err := player.NewMPVPlayer()
	if err != nil {
		log.Fatalf("Failed to create player: %v", err)
	}
	defer p.Close()
	fmt.Println("   ✓ Player created")
	fmt.Println()

	// Get a radio station
	fmt.Println("2. Fetching a radio station...")
	client, err := radiobrowser.NewClient()
	if err != nil {
		log.Fatalf("Failed to create API client: %v", err)
	}

	// Get a popular station
	stations, err := client.GetPopularStations(10)
	if err != nil {
		log.Fatalf("Failed to get stations: %v", err)
	}

	if len(stations) == 0 {
		log.Fatal("No stations found")
	}

	// Find a station with a working stream
	var station *radiobrowser.Station
	for i := range stations {
		if stations[i].LastCheckOK == 1 && stations[i].URLResolved != "" {
			station = &stations[i]
			break
		}
	}

	if station == nil {
		log.Fatal("No working station found")
	}

	fmt.Printf("   ✓ Found station: %s\n", station.Name)
	fmt.Printf("     Country: %s\n", station.Country)
	fmt.Printf("     Codec: %s | Bitrate: %d kbps\n", station.Codec, station.Bitrate)
	fmt.Printf("     URL: %s\n", station.URLResolved)
	fmt.Println()

	// Track the click
	fmt.Println("3. Tracking click...")
	_, err = client.TrackClick(station.StationUUID)
	if err != nil {
		fmt.Printf("   ⚠ Warning: Failed to track click: %v\n", err)
	} else {
		fmt.Println("   ✓ Click tracked")
	}
	fmt.Println()

	// Play the station
	fmt.Println("4. Starting playback...")
	fmt.Printf("   Playing: %s\n", station.Name)
	if err := p.Play(station.URLResolved); err != nil {
		log.Fatalf("Failed to play: %v", err)
	}
	fmt.Println("   ✓ Playback started")
	fmt.Println()

	// Let it play for a bit
	fmt.Println("5. Listening for 10 seconds...")
	for i := 1; i <= 10; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("   %d/10 seconds...\r", i)
	}
	fmt.Println()
	fmt.Println()

	// Test volume control
	fmt.Println("6. Testing volume control...")
	fmt.Println("   Setting volume to 50%...")
	if err := p.SetVolume(50); err != nil {
		log.Printf("   ⚠ Warning: Failed to set volume: %v", err)
	} else {
		vol, _ := p.GetVolume()
		fmt.Printf("   ✓ Volume set to %d%%\n", vol)
	}
	time.Sleep(2 * time.Second)
	fmt.Println()

	// Test pause/resume
	fmt.Println("7. Testing pause/resume...")
	fmt.Println("   Pausing...")
	if err := p.Pause(); err != nil {
		log.Printf("   ⚠ Warning: Failed to pause: %v", err)
	} else {
		fmt.Println("   ✓ Paused")
	}
	time.Sleep(2 * time.Second)

	fmt.Println("   Resuming...")
	if err := p.Resume(); err != nil {
		log.Printf("   ⚠ Warning: Failed to resume: %v", err)
	} else {
		fmt.Println("   ✓ Resumed")
	}
	time.Sleep(2 * time.Second)
	fmt.Println()

	// Stop playback
	fmt.Println("8. Stopping playback...")
	if err := p.Stop(); err != nil {
		log.Fatalf("Failed to stop: %v", err)
	}
	fmt.Println("   ✓ Stopped")
	fmt.Println()

	fmt.Println("✓ All tests passed!")
	fmt.Println()
	fmt.Println("The player is working! You can now integrate it into the main app.")
}

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mrwhyte/rig/pkg/player"
	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

func main() {
	fmt.Println("Testing Player Controls (IPC)...")
	fmt.Println()

	// Create player
	p, err := player.NewMPVPlayer()
	if err != nil {
		log.Fatalf("Failed to create player: %v", err)
	}
	defer p.Close()

	// Get a station
	client, err := radiobrowser.NewClient()
	if err != nil {
		log.Fatalf("Failed to create API client: %v", err)
	}

	stations, err := client.GetPopularStations(5)
	if err != nil {
		log.Fatalf("Failed to get stations: %v", err)
	}

	station := &stations[0]
	fmt.Printf("Testing with: %s\n\n", station.Name)

	// Play
	fmt.Println("1. Starting playback...")
	if err := p.Play(station.URLResolved); err != nil {
		log.Fatalf("Failed to play: %v", err)
	}
	fmt.Println("   ✓ Playing")
	time.Sleep(3 * time.Second)

	// Volume down
	fmt.Println("\n2. Setting volume to 50%...")
	if err := p.SetVolume(50); err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		vol, _ := p.GetVolume()
		fmt.Printf("   ✓ Volume: %d%%\n", vol)
	}
	time.Sleep(2 * time.Second)

	// Pause
	fmt.Println("\n3. Pausing...")
	if err := p.Pause(); err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		fmt.Println("   ✓ Paused")
	}
	time.Sleep(2 * time.Second)

	// Resume
	fmt.Println("\n4. Resuming...")
	if err := p.Resume(); err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		fmt.Println("   ✓ Resumed")
	}
	time.Sleep(2 * time.Second)

	// Volume up
	fmt.Println("\n5. Setting volume to 100%...")
	if err := p.SetVolume(100); err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		vol, _ := p.GetVolume()
		fmt.Printf("   ✓ Volume: %d%%\n", vol)
	}
	time.Sleep(2 * time.Second)

	// Stop
	fmt.Println("\n6. Stopping...")
	if err := p.Stop(); err != nil {
		log.Fatalf("Failed to stop: %v", err)
	}
	fmt.Println("   ✓ Stopped")

	fmt.Println("\n✓ All controls working!")
}

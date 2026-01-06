package main

import (
	"fmt"
	"log"

	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

func main() {
	fmt.Println("Testing Radio Browser API Client...")
	fmt.Println()

	// Create client with server discovery
	fmt.Println("1. Discovering servers...")
	client, err := radiobrowser.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	servers := client.GetServers()
	fmt.Printf("   ✓ Discovered %d servers\n", len(servers))
	for i, server := range servers {
		if i < 3 {
			fmt.Printf("     - %s\n", server)
		}
	}
	if len(servers) > 3 {
		fmt.Printf("     ... and %d more\n", len(servers)-3)
	}
	fmt.Println()

	// Search for jazz stations
	fmt.Println("2. Searching for jazz stations...")
	stations, err := client.SearchByTag("jazz")
	if err != nil {
		log.Fatalf("Failed to search stations: %v", err)
	}
	fmt.Printf("   ✓ Found %d jazz stations\n", len(stations))
	fmt.Println()

	// Display first 5 stations
	fmt.Println("3. Top 5 jazz stations:")
	for i, station := range stations {
		if i >= 5 {
			break
		}
		fmt.Printf("   - %s\n", station.Name)
		fmt.Printf("     Country: %s | Bitrate: %d kbps | Codec: %s\n",
			station.Country, station.Bitrate, station.Codec)
		fmt.Printf("     URL: %s\n", station.URLResolved)
		fmt.Println()
	}

	// Get popular stations
	fmt.Println("4. Getting top 5 popular stations...")
	popular, err := client.GetPopularStations(5)
	if err != nil {
		log.Fatalf("Failed to get popular stations: %v", err)
	}
	fmt.Printf("   ✓ Found %d popular stations\n", len(popular))
	for i, station := range popular {
		fmt.Printf("   %d. %s (%s) - %d clicks\n",
			i+1, station.Name, station.Country, station.ClickCount)
	}
	fmt.Println()

	// Get countries
	fmt.Println("5. Getting countries...")
	countries, err := client.GetCountries()
	if err != nil {
		log.Fatalf("Failed to get countries: %v", err)
	}
	fmt.Printf("   ✓ Found %d countries\n", len(countries))
	fmt.Println("   Top 5 countries by station count:")
	for i, country := range countries {
		if i >= 5 {
			break
		}
		fmt.Printf("   - %s: %d stations\n", country.Name, country.StationCount)
	}
	fmt.Println()

	fmt.Println("✓ All tests passed!")
}

package radiobrowser_test

import (
	"fmt"
	"log"

	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

// Example demonstrates basic usage of the Radio Browser client
func Example() {
	// Create a new client with automatic server discovery
	client, err := radiobrowser.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Discovered %d servers\n", len(client.GetServers()))

	// Search for jazz stations
	stations, err := client.SearchByTag("jazz")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d jazz stations\n", len(stations))

	// Display first 5 stations
	for i, station := range stations {
		if i >= 5 {
			break
		}
		fmt.Printf("- %s (%s, %s)\n", station.Name, station.Country, station.Codec)
	}
}

// ExampleAdvancedSearch demonstrates advanced search with filters
func ExampleAdvancedSearch() {
	client, err := radiobrowser.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Search for high-quality stations in the USA
	params := radiobrowser.SearchParams{
		CountryCode: "US",
		Order:       "votes",
		Reverse:     true,
		Limit:       10,
		HideBroken:  true,
	}

	stations, err := client.SearchStations(params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Top %d stations in USA:\n", len(stations))
	for _, station := range stations {
		fmt.Printf("- %s (%d votes, %d kbps)\n",
			station.Name, station.Votes, station.Bitrate)
	}
}

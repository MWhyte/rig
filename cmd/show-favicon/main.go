package main

import (
	"fmt"
	"log"

	"github.com/mrwhyte/rig/pkg/radiobrowser"
)

func main() {
	client, err := radiobrowser.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Search for BBC World Service
	stations, err := client.SearchByName("BBC World Service")
	if err != nil {
		log.Fatal(err)
	}

	if len(stations) == 0 {
		fmt.Println("No stations found")
		return
	}

	// Show first result
	station := stations[0]
	fmt.Printf("Station: %s\n", station.Name)
	fmt.Printf("Favicon: %s\n", station.Favicon)
	fmt.Printf("URL: %s\n", station.URLResolved)
}

package main

import "fmt"

// Function that displays values in a pretty format
func prettyDisplay(matchedMap map[int]matchedSatelliteContainer) {
	fmt.Printf("\n\n")
	for NCID, records := range matchedMap {
		fmt.Printf("Satellite: %v\n", NCID)
		for _, rec := range records.record {
			if rec.entered {
				fmt.Printf("Entered visible range at: %v\n", rec.time)
			} else {
				fmt.Printf("Left visible range at: \t  %v\n", rec.time)
			}
		}
		fmt.Println("")
	}
}

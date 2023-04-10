package main

import (
	"github.com/kahara/go-pskreporter-spot"
	"time"
)

func main() {
	spotter := spot.NewSpotter("localhost:4739", "N0CALL", "JJ00OG", "Dipole", "fakespot v0", "", spot.SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart, nil)
	for i := 0; i < 100; i++ {
		spotter.Feed(spot.NewSpot("N1CALL", "II00OG", 50313650, -3, 2, "FT8", 1, uint32(time.Now().UTC().Unix())))
	}
	spotter.Close()
}

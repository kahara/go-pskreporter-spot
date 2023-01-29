package main

import (
	spot "github.com/kahara/go-pskreporter-spot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	ReceiverCallsign = "N0CALL"
	ReceiverLocator  = "JJ00AA"
	ReceiverAntenna  = "Dipole"
	ReceiverSoftware = "go-pskreporter-spot fakespot v0"
)

//var ( // Because there are no const arrays
//	SenderCallsigns = []string{
//		"N1CALL",
//		"N2CALL",
//		"N3CALL",
//		"N4CALL",
//		"N5CALL",
//	}
//
//	var SenderAntennas = []string {
//		"Dipole",
//		"Vertical",
//		"Bedsprings",
//	}
//
//	var SenderModes = []string {
//		"FT8",
//		"FT4",
//	}
//
//	var SenderSoftware = []string {
//
//	}
//)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	spotter := spot.NewSpotter("localhost:4739", ReceiverCallsign, ReceiverLocator, ReceiverAntenna, ReceiverSoftware, "", spot.SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart)

	log.Info().Msgf("%+v", spotter)

	for {
		time.Sleep(5 * time.Second)
		spotter.Feed(spot.NewSpot("N1CALL", "IH37OG", 50314350, 23, 8, "FT8", 1, uint32(time.Now().UTC().Unix())))
		spotter.Feed(spot.NewSpot("N2CALL", "II38OG", 50314450, 0, 9, "FT4", 1, uint32(time.Now().UTC().Unix())))
		spotter.Feed(spot.NewSpot("N3CALL", "IJ39OG", 50314550, -42, 10, "FT8", 1, uint32(time.Now().UTC().Unix())))
	}
}

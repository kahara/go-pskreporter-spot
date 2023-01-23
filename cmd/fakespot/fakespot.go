package main

import (
	spot "github.com/kahara/go-pskreporter-spot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	ReceiverCallsign = "N0CALL"
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

	spotter := spot.NewSpotter()

	log.Info().Msgf("%+v", spotter)

	for {
		time.Sleep(1 * time.Second)
		spotter.Feed(spot.Spot{
			Flowstart: 0,
			Sender:    spot.Station{},
			Receiver:  spot.Station{},
			Frequency: 0,
			SNR:       0,
			IMD:       0,
			Software:  "",
			Antenna:   "",
			Mode:      "",
			Source:    0,
			ID:        "",
		})
	}
}

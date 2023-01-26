package spot

import (
	"github.com/rs/zerolog/log"
	"time"
)

const (
	QueueSize  = 10000
	MaxSpots   = 3
	LingerTime = time.Duration(3 * time.Second)
)

// From https://pskreporter.info/pskdev.html
// IPFIX attribute IDs in parenthesis.

type Spotter struct {
	Receiver  Station
	Software  string // (30351.8) "The name and version of the decoding software"
	Antenna   string // (30351.9) "A freeform description of the receiving antenna"
	ID        string // (30351.12) "Random string that identifies the sender. This may be used in the future as a primitive form of security."
	queue     chan Spot
	lastFlush time.Time
	done      chan bool
}

func NewSpotter() *Spotter {
	// Compose a Spotter
	spotter := Spotter{
		queue:     make(chan Spot, QueueSize),
		lastFlush: time.Now(),
		done:      make(chan bool, 1),
	}

	// Flush Spots if there are many of them, or if some time has passed since last flush
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				if len(spotter.queue) >= MaxSpots || (time.Now().Sub(spotter.lastFlush) >= LingerTime && len(spotter.queue) > 0) {
					spotter.flush()
					spotter.lastFlush = time.Now()
				}
			case <-spotter.done: // Attempt to shut down cleanly when done
				spotter.flush()
				return
			}
		}
	}()

	return &spotter
}

// Feed in a Spot to be sent later
func (s *Spotter) Feed(spot Spot) {
	s.queue <- spot
}

// Send Spots
func (s *Spotter) flush() {
	log.Debug().Msg("Flushing spots")

	// FIXME include headers in the first few packets
	// (RFC 5103 says it SHOULD always be sent when the transport is UDP, but PSK Reporter has a different preference.)

	// FIXME and relatedly, handle network error situations
}

func (s *Spotter) Close() {
	s.done <- true
	// FIXME add logic to handle the shutdown cleanly
}

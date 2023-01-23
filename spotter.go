package spot

import (
	"github.com/rs/zerolog/log"
	"time"
)

const (
	IANAEnterpriseID = "30351"
	QueueSize        = 10000
	MaxSpots         = 3
	LingerTime       = time.Duration(3 * time.Second)
)

type Spotter struct {
	queue     chan Spot
	lastFlush time.Time
}

func NewSpotter() *Spotter {
	spotter := Spotter{
		queue:     make(chan Spot, QueueSize),
		lastFlush: time.Now(),
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				// Flush if there are many spots, or if some time has passed since last time
				if len(spotter.queue) >= MaxSpots || (time.Now().Sub(spotter.lastFlush) >= LingerTime && len(spotter.queue) > 0) {
					spotter.flush()
					spotter.lastFlush = time.Now()
				}
			}
		}
	}()

	return &spotter
}

func (s *Spotter) Feed(spot Spot) {
	s.queue <- spot
}

func (s *Spotter) flush() {
	log.Debug().Msg("Flushing spots")
}

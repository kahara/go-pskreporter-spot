package spot

import (
	"github.com/dchest/uniuri"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net"
	"time"
)

const (
	QueueSize                        = 10000
	MaxSpots                         = 3
	LingerTime                       = time.Duration(3 * time.Second)
	InitialHeaderProbability float32 = 4.0
	HeaderProbabilityBackoff float32 = 0.65
	HeaderProbabilityLimit   float32 = 0.1
)

// From https://pskreporter.info/pskdev.html
// IPFIX attribute IDs in parenthesis.

type Spotter struct {
	receiver             Station
	antennaInformation   string // (30351.9) "A freeform description of the receiving antenna"
	decoderSoftware      string // (30351.8) "The name and version of the decoding software"
	persistentIdentifier string // (30351.12) "Random string that identifies the sender. This may be used in the future as a primitive form of security."
	randomIdentifier     uint32
	sequenceNumber       uint32
	headerProbability    float32
	spotKind             int
	queue                chan *Spot
	lastFlush            time.Time
	hostport             string
	done                 chan bool
}

func NewSpotter(hostport string, callsign string, locator string, antennaInformation string, decoderSoftware string, persistentIdentifier string, spotKind int) *Spotter {
	// Compose a Spotter
	spotter := Spotter{
		receiver: Station{
			callsign,
			locator,
		},
		antennaInformation:   antennaInformation,
		decoderSoftware:      decoderSoftware,
		persistentIdentifier: persistentIdentifier,
		randomIdentifier:     rand.Uint32(), // "needed to deal with nasty cases of residential NAT/PAT gateways and DHCP"
		sequenceNumber:       0,
		headerProbability:    InitialHeaderProbability,
		spotKind:             spotKind,
		queue:                make(chan *Spot, QueueSize),
		lastFlush:            time.Now(),
		hostport:             hostport,
		done:                 make(chan bool, 1),
	}

	// Generate a random 30351.12 "persistentIdentifier" if none was provided
	if spotter.persistentIdentifier == "" {
		spotter.persistentIdentifier = uniuri.New()
	}

	// Flush Spots if there are many of them, or if some time has passed since last flush
	go func() {
		const (
			InitialDelay = 100
			Backoff      = 2
			Limit        = 10000
		)

		var (
			err   error
			delay time.Duration = InitialDelay
			conn  net.Conn
		)

		for {
			// Prepare UDP "connection"
			for {
				conn, err = net.Dial("udp", spotter.hostport)
				if err != nil {
					log.Err(err).Msg("")
					time.Sleep(delay * time.Millisecond)
					delay *= Backoff
					if delay > Limit {
						delay = Limit
					}
					continue
				} else {
					break
				}
			}

			// Send an initial packet which may contain just the descriptors
			err = spotter.flush(conn)
			if err != nil {
				break
			}

			// Start sending periodically
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ticker.C:
					if len(spotter.queue) >= MaxSpots || (time.Now().Sub(spotter.lastFlush) >= LingerTime && len(spotter.queue) > 0) {
						err = spotter.flush(conn)
						if err != nil {
							break
						}
						spotter.lastFlush = time.Now()
					}
				case <-spotter.done: // Attempt to shut down cleanly when done
					err = spotter.flush(conn)
					if err != nil {
						break
					}
					return
				}
			}
		}
	}()

	return &spotter
}

// Feed in a Spot to be sent later
func (s *Spotter) Feed(spot *Spot) {
	s.queue <- spot
}

// Send Spots
func (s *Spotter) flush(conn net.Conn) error {
	log.Debug().Msg("Flushing spots")

	var (
		err         error
		descriptors []byte
		records     []byte
		datagram    []byte
	)

	// Get receiver and sender records, if any
	records = IPFIXRecords(s)

	// Include descriptors with steadily decreasing probability, down to a limit
	// (RFC 5103 says they SHOULD always be sent when transport is UDP, but PSK Reporter has a different preference.)
	if rand.Float32() < s.headerProbability {
		descriptors = IPFIXDescriptors(s)
	}

	if s.headerProbability > HeaderProbabilityLimit {
		s.headerProbability *= HeaderProbabilityBackoff
	} else {
		s.headerProbability = HeaderProbabilityLimit
	}

	// Combine everything into a packet
	datagram = IPFIX(s.sequenceNumber, s.randomIdentifier, descriptors, records)

	// Send packet
	_, err = conn.Write(datagram)
	if err != nil {
		return err
	}

	s.sequenceNumber += 1

	return nil
}

func (s *Spotter) Close() {
	s.done <- true
	// FIXME add logic to handle the shutdown cleanly
}

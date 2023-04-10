package spot

import (
	"github.com/dchest/uniuri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	QueueSize                        = 10000
	MaxSpots                         = 25
	LingerTime                       = time.Duration(300 * time.Second)
	InitialHeaderProbability float32 = 4.0
	HeaderProbabilityBackoff float32 = 0.65
	HeaderProbabilityLimit   float32 = 0.1
	IPv4MaxPayloadBytes              = 576 - 60 - 8 - 20 // Minimum MTU - IP header - UDP header - additional headroom
	IPv6MaxPayloadBytes              = 1280 - 40 - 8     // TODO Verify that this really is a reasonable assumption
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
	ipfixDescriptors     []byte
	queue                chan *Spot
	lastFlush            time.Time
	hostport             string
	maxPayloadBytes      int
	packetMetric         *prometheus.CounterVec
	done                 chan bool
	doneAck              chan bool
}

func NewSpotter(hostport string, callsign string, locator string, antennaInformation string, decoderSoftware string, persistentIdentifier string, spotKind int, packetMetric *prometheus.CounterVec) *Spotter {
	// For randomIdentifier
	rand.Seed(time.Now().UnixNano())

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
		ipfixDescriptors:     []byte{},
		queue:                make(chan *Spot, QueueSize),
		lastFlush:            time.Now(),
		hostport:             hostport,
		maxPayloadBytes:      0,
		packetMetric:         packetMetric,
		done:                 make(chan bool, 1),
		doneAck:              make(chan bool, 1),
	}

	// Construct IPFIX descriptors
	if spotter.antennaInformation == "" {
		spotter.ipfixDescriptors = append(spotter.ipfixDescriptors, ReceiverDescriptor_CallsignLocatorSoftware...)
	} else {
		spotter.ipfixDescriptors = append(spotter.ipfixDescriptors, ReceiverDescriptor_CallsignLocatorSoftwareAntenna...)
	}

	if spotter.spotKind == SpotKind_CallsignFrequencyModeSourceFlowstart {
		spotter.ipfixDescriptors = append(spotter.ipfixDescriptors, SenderDescriptor_CallsignFrequencyModeSourceFlowstart...)
	} else if spotter.spotKind == SpotKind_CallsignFrequencyModeSourceLocatorFlowstart {
		spotter.ipfixDescriptors = append(spotter.ipfixDescriptors, SenderDescriptor_CallsignFrequencyModeSourceLocatorFlowstart...)
	} else if spotter.spotKind == SpotKind_CallsignFrequencySNRIMDModeSourceFlowstart {
		spotter.ipfixDescriptors = append(spotter.ipfixDescriptors, SenderDescriptor_CallsignFrequencySNRIMDModeSourceFlowstart...)
	} else if spotter.spotKind == SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart {
		spotter.ipfixDescriptors = append(spotter.ipfixDescriptors, SenderDescriptor_CallsignFrequencySNRIMDModeSourceLocatorFlowstart...)
	}

	// Make some hopefully correct assumptions about how many bytes can be crammed into each packet without hitting MTU
	if strings.Count(spotter.hostport, ":") == 1 {
		spotter.maxPayloadBytes = IPv4MaxPayloadBytes
	} else {
		spotter.maxPayloadBytes = IPv6MaxPayloadBytes
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
				case <-spotter.done:
					// Attempt to shut down cleanly when done; this may or may not get everything written out in time
					_ = spotter.flush(conn)
					_ = conn.Close()
					spotter.doneAck <- true
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

	// Get receiver and sender records, if any
	records = IPFIXRecords(s, len(descriptors)+HeaderLength)

	// Combine everything into a packet
	datagram = IPFIX(s.sequenceNumber, s.randomIdentifier, descriptors, records)

	// Send packet
	// FIXME figure out how to handle potentially unsent data when writing fails
	_, err = conn.Write(datagram)
	if s.packetMetric != nil {
		s.packetMetric.WithLabelValues(conn.LocalAddr().Network(), conn.RemoteAddr().String()).Inc()
	}
	if err != nil {
		return err
	}

	// FIXME related to the above remark about failing writes
	s.sequenceNumber += 1

	return nil
}

func (s *Spotter) Close() {
	s.done <- true
	select {
	case <-s.doneAck:
		log.Debug().Str("callsign", s.receiver.Callsign).Msg("Connection to reporter closed")
	}
}

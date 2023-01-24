package spot

import (
	"github.com/rs/zerolog/log"
	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
	"time"
)

const (
	IANAEnterpriseID = "30351"
	QueueSize        = 10000
	MaxSpots         = 3
	LingerTime       = time.Duration(3 * time.Second)
)

type Spotter struct {
	queue       chan Spot
	lastFlush   time.Time
	exporter    *exporter.ExportingProcess
	templateID  uint16
	templateSet entities.Set
	done        chan bool
}

func NewSpotter() *Spotter {
	// Set up IPFIX export
	export, err := exporter.InitExportingProcess(exporter.ExporterInput{
		CollectorAddress:    "127.0.0.1:14739",
		CollectorProtocol:   "udp",
		ObservationDomainID: 0,
		TempRefTimeout:      0,
		TLSClientConfig:     nil,
		IsIPv6:              false,
		SendJSONRecord:      false,
		JSONBufferLen:       0,
		CheckConnInterval:   0,
	})
	if err != nil {
		log.Fatal().Err(err)
	}

	// Compose Spotter itself
	spotter := Spotter{
		queue:       make(chan Spot, QueueSize),
		lastFlush:   time.Now(),
		exporter:    export,
		templateID:  export.NewTemplateID(),
		templateSet: entities.NewSet(false),
		done:        make(chan bool, 1),
	}

	// Do the IPFIX boilerplate dance
	err = spotter.templateSet.PrepareSet(entities.Template, spotter.templateID)
	if err != nil {
		log.Fatal().Err(err)
	}

	element, err := registry.GetInfoElement("flowStartSeconds", registry.IANAEnterpriseID)
	if err != nil {
		log.Fatal().Err(err)
	}

	log.Info().Msgf("%+v", element)

	flowStartSeconds, _ := entities.DecodeAndCreateInfoElementWithValue(element, nil)

	spotter.templateSet.AddRecord([]entities.InfoElementWithValue{flowStartSeconds}, spotter.templateID)

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
				spotter.exporter.CloseConnToCollector()
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

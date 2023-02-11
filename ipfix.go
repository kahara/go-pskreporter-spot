package spot

import (
	"encoding/binary"
	"github.com/rs/zerolog/log"
	"time"
)

// See https://pskreporter.info/pskdev.html

const (
	HeaderLength = 16
	// TODO figure out what is the "theoretical" maximum size of sender record,
	// TODO for making the decision to get an item from the queue
)

var (
	Header               = []byte{0x00, 0x0A} // "Version" in RFC 5101
	ReceiverRecordHeader = []byte{0x99, 0x92} // "Set ID" in RFC 5101(?)
	SenderRecordHeader   = []byte{0x99, 0x93}

	ReceiverDescriptor_CallsignLocatorSoftware = []byte{
		0x00, 0x03, 0x00, 0x24, 0x99, 0x92, 0x00, 0x03, 0x00, 0x00,
		0x80, 0x02, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x04, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x08, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x00, 0x00,
	}
	ReceiverDescriptor_CallsignLocatorSoftwareAntenna = []byte{
		0x00, 0x03, 0x00, 0x2C, 0x99, 0x92, 0x00, 0x04, 0x00, 0x00,
		0x80, 0x02, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x04, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x08, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x09, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x00, 0x00,
	}
	SenderDescriptor_CallsignFrequencyModeSourceFlowstart = []byte{
		0x00, 0x02, 0x00, 0x2C, 0x99, 0x93, 0x00, 0x05,
		0x80, 0x01, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x05, 0x00, 0x04, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0A, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0B, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x00, 0x96, 0x00, 0x04,
	}
	SenderDescriptor_CallsignFrequencyModeSourceLocatorFlowstart = []byte{
		0x00, 0x02, 0x00, 0x34, 0x99, 0x93, 0x00, 0x06,
		0x80, 0x01, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x05, 0x00, 0x04, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0A, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0B, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x03, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x00, 0x96, 0x00, 0x04,
	}
	SenderDescriptor_CallsignFrequencySNRIMDModeSourceFlowstart = []byte{
		0x00, 0x02, 0x00, 0x3C, 0x99, 0x93, 0x00, 0x07,
		0x80, 0x01, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x05, 0x00, 0x04, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x06, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x07, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0A, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0B, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x00, 0x96, 0x00, 0x04,
	}
	SenderDescriptor_CallsignFrequencySNRIMDModeSourceLocatorFlowstart = []byte{
		0x00, 0x02, 0x00, 0x44, 0x99, 0x93, 0x00, 0x08,
		0x80, 0x01, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x05, 0x00, 0x04, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x06, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x07, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0A, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x0B, 0x00, 0x01, 0x00, 0x00, 0x76, 0x8F,
		0x80, 0x03, 0xFF, 0xFF, 0x00, 0x00, 0x76, 0x8F,
		0x00, 0x96, 0x00, 0x04,
	}
)

func IPFIX(sequenceNumber uint32, observationDomain uint32, descriptors []byte, records []byte) []byte {
	var (
		ipfix  []byte
		header [16]byte
	)

	// Construct an IPFIX header with message length, timestamp, sequence number, and observation domain
	header[0] = Header[0]
	header[1] = Header[1]
	binary.BigEndian.PutUint16(header[2:], uint16(HeaderLength+len(descriptors)+len(records)))
	binary.BigEndian.PutUint32(header[4:], uint32(time.Now().UTC().Unix()))
	binary.BigEndian.PutUint32(header[8:], sequenceNumber)
	binary.BigEndian.PutUint32(header[12:], observationDomain)

	// Concatenate everything
	ipfix = append(ipfix, header[:]...)   // Header is an array...
	ipfix = append(ipfix, descriptors...) // ...and these are slices
	ipfix = append(ipfix, records...)

	return ipfix
}

func IPFIXDescriptors(spotter *Spotter) []byte {
	return spotter.ipfixDescriptors
}

func IPFIXRecords(spotter *Spotter, spent int) []byte {
	var (
		records          []byte
		payloadBytesLeft = spotter.maxPayloadBytes - spent - 3 - 3 // Leave margin for paddings, too
		header           [4]byte
		receiverRecord   []byte
		senderRecords    []byte
		length           = 0
		padding          = 0
	)

	// Receiver record; callsign, locator, decoderSoftware, (optionally) antennaInformation
	// FIXME limit the strings' lengths
	receiverRecord = append(receiverRecord, uint8(len(spotter.receiver.Callsign)))
	receiverRecord = append(receiverRecord, []byte(spotter.receiver.Callsign)...)
	receiverRecord = append(receiverRecord, uint8(len(spotter.receiver.Locator)))
	receiverRecord = append(receiverRecord, []byte(spotter.receiver.Locator)...)
	receiverRecord = append(receiverRecord, uint8(len(spotter.decoderSoftware)))
	receiverRecord = append(receiverRecord, []byte(spotter.decoderSoftware)...)
	if spotter.antennaInformation != "" {
		receiverRecord = append(receiverRecord, uint8(len(spotter.antennaInformation)))
		receiverRecord = append(receiverRecord, []byte(spotter.antennaInformation)...)
	}

	length = len(header) + len(receiverRecord)
	padding = 4 - (length % 4)
	length += padding

	header[0] = ReceiverRecordHeader[0]
	header[1] = ReceiverRecordHeader[1]
	binary.BigEndian.PutUint16(header[2:], uint16(length))
	records = append(records, header[:]...)
	records = append(records, receiverRecord...)

	// Add padding for 4-byte alignment
	for i := 0; i < padding; i++ {
		records = append(records, 0)
	}

	payloadBytesLeft = payloadBytesLeft - length

	// Sender records
Senders:
	for {
		var (
			spot         *Spot
			senderRecord []byte
		)

		select {
		case s := <-spotter.queue:
			spot = s
			log.Info().Msgf("%+v", spot)
		default:
			break Senders
		}

		// Callsign and frequency
		senderRecord = append(senderRecord, uint8(len(spot.sender.Callsign)))
		senderRecord = append(senderRecord, []byte(spot.sender.Callsign)...)
		senderRecord = append(senderRecord, []byte{0, 0, 0, 0}...)
		binary.BigEndian.PutUint32(senderRecord[len(senderRecord)-4:], uint32(spot.frequency))

		// Add noise and distortion if these are supposed to be available
		if spotter.spotKind == SpotKind_CallsignFrequencySNRIMDModeSourceFlowstart || spotter.spotKind == SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart {
			senderRecord = append(senderRecord, byte(spot.snr))
			senderRecord = append(senderRecord, byte(spot.imd))
		}

		// Mode and source
		senderRecord = append(senderRecord, uint8(len(spot.mode)))
		senderRecord = append(senderRecord, []byte(spot.mode)...)
		senderRecord = append(senderRecord, byte(spot.informationSource))

		// Locator
		if spotter.spotKind == SpotKind_CallsignFrequencyModeSourceLocatorFlowstart || spotter.spotKind == SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart {
			senderRecord = append(senderRecord, uint8(len(spot.sender.Locator)))
			senderRecord = append(senderRecord, []byte(spot.sender.Locator)...)
		}

		// Beginning of transmission
		senderRecord = append(senderRecord, []byte{0, 0, 0, 0}...)
		binary.BigEndian.PutUint32(senderRecord[len(senderRecord)-4:], spot.flowStartSeconds)

		// If it starts to look like adding more would make the packet's size go over MTU, put the spot back into queue
		// TODO see comment about "theoretical" maximum size of sender record earlier in the file; this could be smarter
		if (len(header) + len(receiverRecord)) > payloadBytesLeft {
			spotter.queue <- spot
			log.Info().Msg("skipping")
			break Senders
		} else {
			senderRecords = append(senderRecords, senderRecord...)
		}
	}

	length = len(header) + len(senderRecords)
	padding = 4 - (length % 4)
	length += padding

	header[0] = SenderRecordHeader[0]
	header[1] = SenderRecordHeader[1]
	binary.BigEndian.PutUint16(header[2:], uint16(length))
	records = append(records, header[:]...)
	records = append(records, senderRecords...)

	// Pad the sender records, too
	for i := 0; i < padding; i++ {
		records = append(records, 0)
	}

	return records
}

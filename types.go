package spot

// From https://pskreporter.info/pskdev.html
// IPFIX attribute IDs in parenthesis.

type Station struct {
	Callsign string // (30351.{1,2}) "The callsign of the {sender,receiver} of the transmission"
	Locator  string // (30351.{3,4}) "The locator of the {sender,receiver} of the transmission"
}

type Sender struct {
	Station
}

type Receiver struct {
	Station
}

type Spot struct {
	Flowstart uint32 // (150) "The time of the transmission (absolute seconds since 1/1/1970)"
	Sender    Sender
	Receiver  Receiver
	Frequency uint64 // (30351.5) "The frequency of the transmission in Hertz"
	SNR       int8   // (30351.6) "The signal to noise ration of the transmission. Normally 1 byte"
	IMD       uint8  // (30351.7) "The intermodulation distortion of the transmission. Normally 1 byte."
	Software  string // (30351.8) "The name and version of the decoding software"
	Antenna   string // (30351.9) "A freeform description of the receiving antenna"
	Mode      string // (30351.10) "The mode of the communication. One of the ADIF values for MODE or SUBMODE"
	Source    uint8  // (30351.11) "Identifies the source of the record. The bottom 2 bits have the following meaning: 1 = Automatically Extracted. 2 = From a Call Log (QSO). 3 = Other Manual Entry. The 0x80 bit indicates that this record is a test transmission. Normally 1 byte."
	ID        string // (30351.12) "Random string that identifies the sender. This may be used in the future as a primitive form of security."
}

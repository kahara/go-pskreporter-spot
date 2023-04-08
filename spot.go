package spot

// From https://pskreporter.info/pskdev.html
// IPFIX attribute IDs in parenthesis.

const (
	SpotKind_CallsignFrequencyModeSourceFlowstart = iota
	SpotKind_CallsignFrequencyModeSourceLocatorFlowstart
	SpotKind_CallsignFrequencySNRIMDModeSourceFlowstart
	SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart
)

type Station struct {
	Callsign string // (30351.{1,2}) "The callsign of the {sender,receiver} of the transmission"
	Locator  string // (30351.{3,4}) "The locator of the {sender,receiver} of the transmission"
}

type Spot struct {
	sender            Station
	frequency         uint64 // (30351.5) "The frequency of the transmission in Hertz"
	snr               int8   // (30351.6) "The signal to noise ration of the transmission. Normally 1 byte"
	imd               uint8  // (30351.7) "The intermodulation distortion of the transmission. Normally 1 byte."
	mode              string // (30351.10) "The mode of the communication. One of the ADIF values for MODE or SUBMODE"
	informationSource uint8  // (30351.11) "Identifies the source of the record. The bottom 2 bits have the following meaning: 1 = Automatically Extracted. 2 = From a Call Log (QSO). 3 = Other Manual Entry. The 0x80 bit indicates that this record is a test transmission. Normally 1 byte."
	flowStartSeconds  uint32 // (150) "The time of the transmission (absolute seconds since 1/1/1970)"
}

func NewSpot(callsign string, locator string, frequency uint64, snr int8, imd uint8, mode string, informationSource uint8, flowStartSeconds uint32) *Spot {
	return &Spot{
		sender: Station{
			callsign,
			locator,
		},
		frequency:         frequency,
		snr:               snr,
		imd:               imd,
		mode:              mode,
		informationSource: informationSource,
		flowStartSeconds:  flowStartSeconds,
	}
}

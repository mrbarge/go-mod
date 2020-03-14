package module
import (
	"encoding/binary"
	"fmt"
	"errors"
)

type FastTracker struct {
	title string
	author string
	version uint16
	patternSize uint16
	restartPos uint16
	flags uint16
	tempo uint16
	bpm uint16
	orderTable []byte
	instruments []STInstrument
	patterns []Pattern
	Module
}

func (m *FastTracker) Type() FileFormat {
	return FASTTRACKER
}

func (m *FastTracker) Load(data []byte) (error) {
	m.title = filterNulls(string(data[17:37]))
	m.author = filterNulls(string(data[38:58]))
	m.version = binary.LittleEndian.Uint16(data[58:60])
	//headerSize := binary.LittleEndian.Uint32(data[60:64])
	m.patternSize = binary.LittleEndian.Uint16(data[64:66])
	m.restartPos = binary.LittleEndian.Uint16(data[66:68])
	//numChannels := binary.LittleEndian.Uint16(data[68:70])
	//numPatterns := binary.LittleEndian.Uint16(data[70:72])
	//numInstruments := binary.LittleEndian.Uint16(data[72:74])
	m.flags = binary.LittleEndian.Uint16(data[74:76])
	m.tempo = binary.LittleEndian.Uint16(data[76:78])
	m.bpm = binary.LittleEndian.Uint16(data[78:80])
	m.orderTable = data[80:336]



	instrumentCount := binary.LittleEndian.Uint16(data[34:36])

	return errors.New("Unsupported")
}

func (m *FastTracker) Play() {
	fmt.Printf("Playing FT..\n")
}

func (m *FastTracker) Title() (string) {
	return m.title
}

func (m *FastTracker) Instruments() []Instrument {
	r := make([]Instrument, len(m.instruments))
	for i := range m.instruments {
		r[i] = m.instruments[i]
	}
	return r
}


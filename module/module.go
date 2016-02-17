package module

import (
	"io/ioutil"
	"fmt"
	"os"
	"encoding/binary"
	"errors"
	"github.com/alexcesaro/log/stdlog"
)

const (
	MOD = iota
	SCREAMTRACKER
	FASTTRACKER
)

type Pattern struct {
	data []byte
}

type Instrument struct {
	name string
	length int64
	finetune int8
	volume int8
	repeatOffset uint16
	repeatLength uint16
	data []byte
}

type Module interface {
	Load(data []byte) (error)
	Title() (string)
	Play()
}

type ProTracker struct {
	title string
	numChannels int8
	songLength int8
	restartPos int8
	patternTable [128]int8
	instruments [31]Instrument
	patterns []Pattern
	Module
}

type ScreamTracker struct {
	title string
	Module
}

type FastTracker struct {
	title string
	Module
}

func (i *Instrument) Load(data []byte) (error) {
	i.name = string(data[0:22])
	// Length is stored as number of words in the PT format, but we'll store it as number of bytes
	i.length = int64(binary.BigEndian.Uint16(data[22:24])) * 2
	i.finetune = int8(data[24])
	i.volume = int8(data[25])
	i.repeatOffset = binary.BigEndian.Uint16(data[25:27])
	i.repeatLength = binary.BigEndian.Uint16(data[27:29])

	return nil
}

func (m *ProTracker) Load(data []byte) error {
	name := string(data[0:20])
	length := len(data)
	m.title = name

	// Load sample metadata
	offset := int(20)
	// Todo: If there's no magic number we should assume only 15 samples, not 31
	for i := 0; i < 31; i++ {
		sampleMeta := data[offset:offset+30]
		instrument := Instrument{}
		instrument.Load(sampleMeta)
		offset += 30
		m.instruments[i] = instrument
	}

	m.songLength = int8(data[offset])
	offset++
	m.restartPos = int8(data[offset])
	offset++
	for i := 0; i < 128; i++ {
		m.patternTable[i] = int8(data[offset+i])
	}
	offset += 128
	magicNumber := string(data[offset:offset+4])
	offset += 4

	// Validate our magic number against known possible values and hence number of channels
	if (magicNumber == "M.K." || magicNumber == "FLT4" || magicNumber == "M!K!") {
		m.numChannels = 4
	} else if (magicNumber == "6CHN") {
		m.numChannels = 6
	} else if (magicNumber == "8CHN" || magicNumber == "FLT8") {
		m.numChannels = 8
	} else {
		// Assume 4 channels and that the magic number starting offset is actually
		// the start of the pattern data, so let's roll back
		m.numChannels = 4
		offset -= 4
	}

	// Start reading the pattern data
	numPatterns := m.NumPatterns()
	stdlog.GetFromFlags().Debugf("Starting to read %d patterns starting at offset %d",numPatterns,offset)
	m.patterns = make([]Pattern, numPatterns, numPatterns)
	for i := 0; i < numPatterns; i++ {
		pattern := Pattern{}
		// Sanity check for enough data remaining in the buffer
		if (offset + 1024) > length {
			errtxt := fmt.Sprintf("Exceeded remaining data length on pattern %d",i)
			return errors.New(errtxt)
		}
		pattern.data = make([]byte, 1024)
		copy(pattern.data, data[offset:offset+1024])
		offset += 1024
		m.patterns[i] = pattern
	}

	// Start reading the sample data
	for i := 0; i < len(m.instruments); i++ {
		instrument := m.instruments[i]
		stdlog.GetFromFlags().Debugf("Loading sample %d of length %d (%s)",i,instrument.length,instrument.name)
		instrument.data = make([]byte, instrument.length)
		// Sanity check for enough data remaining in the buffer
		if (offset + int(instrument.length) > length) {
			errtxt := fmt.Sprintf("Exceeded remaining data length on sample %d (%s)",i,instrument.name)
			return errors.New(errtxt)
		}
		offset += int(instrument.length)
		copy(instrument.data, data[offset:offset+int(instrument.length)])
	}

	stdlog.GetFromFlags().Debugf("I'm done at offset %d and length was %d",offset,length)
	return nil
}

func (m *ProTracker) SongLength() int8 {
	return m.songLength
}

func (m *ProTracker) PatternTable() [128]int8 {
	return m.patternTable
}

func (m *ProTracker) NumPatterns() int {
	numPatterns := int(m.patternTable[0])
	for i := 0; i < len(m.patternTable); i++ {
		if int(m.patternTable[i]) > numPatterns {
			numPatterns = int(m.patternTable[i])
		}
	}
	return numPatterns
}
func (m *ProTracker) Title() (string) {
	return m.title
}

func (m *ProTracker) Play() {
	fmt.Printf("Playing PT..\n")
}

func (m *ScreamTracker) Load(data []byte) (error) {
	m.title = "nothing yet ST"
	return nil
}
func (m *ScreamTracker) Play() {
	fmt.Printf("Playing ST..\n")
}
func (m *ScreamTracker) Title() (string) {
	return m.title
}
func (m *FastTracker) Load(data []byte) (error) {
	m.title = "nothing yet FT"
	return nil
}
func (m *FastTracker) Play() {
	fmt.Printf("Playing FT..\n")
}
func (m *FastTracker) Title() (string) {
	return m.title
}

// Magic Numbers
// S3M: hex position 25: 00 00 00 1A 10 00 00
// XM: 0-17, "Extended Module:" 45 78 74 65 6E 64 65 64 20 6D 6F 64 75 6C 65 3A 20
// Mod:

func Load(modFile string) (Module,error) {

	f, err := os.Open(modFile)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	f.Close()


	// Determine file type
	if (string(data[44:48]) == "SCRM") {
		m := &ScreamTracker{}
		err = m.Load(data)
		return m, err
	}

	if (string(data[0:17]) == "Extended Module: ") {
		m := &FastTracker{}
		err = m.Load(data)
		return m, err
	}

	m := &ProTracker{}
	err = m.Load(data)

	return m, err

}

package module

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"errors"
)

type ProTracker struct {
	title string
	numChannels int8
	songLength int8
	restartPos int8
	sequenceTable [128]int8
	instruments []Instrument
	samples []PTSample
	patterns []Pattern
	Module
}

func (m *ProTracker) Type() FileFormat {
	return PROTRACKER
}

func (m *ProTracker) Load(data []byte) error {
	name := string(data[0:20])
	length := len(data)
	m.title = name

	// Load sample metadata
	offset := int(20)
	// Todo: If there's no magic number we should assume only 15 samples, not 31

	sampleDatas := make([][]byte, 0)

	for i := 0; i < 31; i++ {
		sampleMeta := data[offset:offset+30]
		sampleDatas = append(sampleDatas, sampleMeta)
		offset += 30
	}

	m.songLength = int8(data[offset])
	offset++
	m.restartPos = int8(data[offset])
	offset++
	for i := 0; i < 128; i++ {
		m.sequenceTable[i] = int8(data[offset+i])
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
	slog.Debug("Starting to read patterns", "num-patterns", numPatterns, "offset", offset)
	m.patterns = make([]Pattern, numPatterns, numPatterns)
	for i := 0; i < numPatterns; i++ {
		pattern := Pattern{rows:make([]Row,64), numChannels:m.numChannels}
		// Sanity check for enough data remaining in the buffer
		for j := 0; j < pattern.NumRows(); j++ {
			row := Row{notes:make([]Note, m.numChannels)}
			for k := 0; k < pattern.NumChannels(); k++ {
				n := Note{}
				n.Load(data[offset:offset+4])
				row.notes[k] = n
				offset += 4
			}
			pattern.SetRow(j,row)
		}
		/*
			if (offset + 1024) > length {
				errtxt := fmt.Sprintf("Exceeded remaining data length on pattern %d",i)
				return errors.New(errtxt)
			}*/
		//		pattern.data = make([]byte, 1024)
		//		copy(pattern.data, data[offset:offset+1024])
		//offset += 1024
		m.patterns[i] = pattern
	}

	// Start reading the sample data
	for i, sampleData := range sampleDatas {
		sample := PTSample{}
		sample.name = filterNulls(string(sampleData[0:22]))
		// Length is stored as number of words in the PT format, but we'll store it as number of bytes
		sample.length = int64(binary.BigEndian.Uint16(sampleData[22:24])) * 2
		sample.finetune = int8(sampleData[24])
		sample.volume = int8(sampleData[25])
		sample.repeatOffset = binary.BigEndian.Uint16(sampleData[25:27])
		sample.repeatLength = binary.BigEndian.Uint16(sampleData[27:29])
		sample.data = data[offset:offset+int(sample.length)]
		// Sanity check for enough data remaining in the buffer
		if (offset + int(sample.length) > length) {
			errtxt := fmt.Sprintf("Exceeded remaining data length on sample %d (%s)",i, sample.name)
			return errors.New(errtxt)
		}
		m.samples = append(m.samples, sample)
		offset += int(sample.length)
	}

	//for i := 0; i < len(m.samples); i++ {
	//	sample := m.samples[i]
	//	slog.Debug("Loading sample", "index", i, "offset", offset, "length", sample.length, "name", sample.name)
	//}
	//
	slog.Debug("Done loading", "offset", offset, "length", length)
	return nil
}

func (m *ProTracker) GetSample(i int) (PTSample,error) {
	if (i < 0 || i > len(m.samples)) {
		return PTSample{},errors.New("Invalid sample")
	} else {
		return m.samples[i],nil
	}
}

func (m *ProTracker) Instruments() []Instrument {
	return []Instrument{}
}

func (m *ProTracker) Samples() []Sample {
	r := make([]Sample, len(m.samples))
	for i := range m.samples {
		r[i] = m.samples[i]
	}
	return r
}

func (m *ProTracker) SongLength() int8 {
	return m.songLength
}

func (m *ProTracker) SequenceTable() [128]int8 {
	return m.sequenceTable
}

func (m *ProTracker) Patterns() []Pattern {
	return m.patterns
}

func (m *ProTracker) NumPatterns() int {
	numPatterns := int(m.sequenceTable[0])
	for i := 0; i < len(m.sequenceTable); i++ {
		if int(m.sequenceTable[i]) > numPatterns {
			numPatterns = int(m.sequenceTable[i])
		}
	}
	// Increment num patterns because we count from zero, so we actually have highest-pattern-id+1 patterns.
	numPatterns++
	return numPatterns
}
func (m *ProTracker) Title() (string) {
	return m.title
}

func (m *ProTracker) Play() {
	fmt.Printf("Playing PT..\n")
}

func (m *ProTracker) GetPattern(patternNumber int8) (Pattern,error) {
	if (patternNumber < 0 || patternNumber > 127) {
		return Pattern{},errors.New("Pattern index out of range.")
	} else {
		return m.patterns[patternNumber], nil
	}
}
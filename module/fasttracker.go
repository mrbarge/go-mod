package module

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	instruments []FTInstrument
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
	numPatterns := binary.LittleEndian.Uint16(data[70:72])
	numInstruments := binary.LittleEndian.Uint16(data[72:74])
	m.flags = binary.LittleEndian.Uint16(data[74:76])
	m.tempo = binary.LittleEndian.Uint16(data[76:78])
	m.bpm = binary.LittleEndian.Uint16(data[78:80])
	m.orderTable = data[80:336]

	offset := 336
	for i := 0; i < int(numPatterns); i++ {
		//hdrLength := binary.LittleEndian.Uint32(data[offset:offset+4])
		offset += 5		// skip packing type
		//numPatternRows := binary.LittleEndian.Uint16(data[offset:offset+2])
		offset += 2
		patternDataSize := binary.LittleEndian.Uint16(data[offset:offset+2])
		offset += 2
		offset += int(patternDataSize)
	}

	for i := 0; i < int(numInstruments); i++ {

		instrument := FTInstrument{}

		instOffset := offset
		instHeaderSize := binary.LittleEndian.Uint32(data[instOffset:instOffset+4])
		instOffset += 4
		instrument.name = string(data[instOffset:instOffset+22])
		instOffset += 23	// skip type
		instNumSamples :=  binary.LittleEndian.Uint16(data[instOffset:instOffset+2])
		instOffset += 2

		sampleSizes := make([]int, 0)
		for j := 0; j < int(instNumSamples); j++ {
			sampleHeaderSize := binary.LittleEndian.Uint32(data[instOffset:instOffset+4])
			sampleSizes = append(sampleSizes, int(sampleHeaderSize))
			instOffset += 96	// skip keymap assignments
			instOffset += 48	// skip volume envelope points
			instOffset += 48	// skip panning envelope points
		}

		instOffset += 38

		offset += int(instHeaderSize)

		// read sample datas
		for j := 0; j < int(instNumSamples); j++ {
			sample := FTSample{}

			sampleOffset := offset
			sample.length = binary.LittleEndian.Uint32(data[sampleOffset : sampleOffset+4])
			sampleOffset += 4
			sample.loopStart = binary.LittleEndian.Uint32(data[sampleOffset : sampleOffset+4])
			sampleOffset += 4
			sample.loopEnd = binary.LittleEndian.Uint32(data[sampleOffset : sampleOffset+4])
			sampleOffset += 4
			sample.volume = data[sampleOffset]
			sampleOffset += 1
			sample.finetune = data[sampleOffset]
			sampleOffset += 1
			sample.sampleType = data[sampleOffset]
			sampleOffset += 1
			sample.panning = data[sampleOffset]
			sampleOffset += 1
			sample.relativeNote = data[sampleOffset]
			sampleOffset += 1
			sample.dataType = data[sampleOffset]
			sampleOffset += 1
			sample.name = string(data[sampleOffset : sampleOffset+22])
			sampleOffset += 22

			instrument.samples = append(instrument.samples, sample)
			offset = sampleOffset
		}

		for j := 0; j < int(instNumSamples); j++ {
			sampleOffset := offset
			if ((1 << 4) & instrument.samples[j].sampleType) == 0 {
				fmt.Println(instrument.samples[j].length)
				codedSampleData := data[sampleOffset:sampleOffset+int(instrument.samples[j].length)]
				instrument.samples[j].data = decode8Bit(codedSampleData)
				offset += int(instrument.samples[j].length)
			} else {
				codedSampleData := data[sampleOffset:sampleOffset+int(instrument.samples[j].length)]
				instrument.samples[j].data = decode16Bit(codedSampleData, instrument.samples[j].length)
				offset += int(instrument.samples[j].length)
			}
		}

		m.instruments = append(m.instruments, instrument)
	}

	return nil

}

func decode8Bit(data []byte) []byte {

	r := make([]byte, len(data))
	old := int8(0)
	for i, v := range data {
		old += int8(v)
		r[i] = byte(old)
	}
	return r
}

func decode16Bit(data []byte, sampleLength uint32) []byte {
	r := make([]byte, len(data))
	old := int16(0)
	for _, v := range data {
		old += int16(v)

		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, old)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}

		r = append(r, buf.Bytes()...)
		//r[i] = byte(old)
	}
	return r
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

func (m *FastTracker) Samples() []Sample {
	r := make([]Sample, 0)
	for _, instrument := range m.instruments {
		for _, sample := range instrument.samples {
			r = append(r, sample)
		}
	}
	return r
}

func (m *FastTracker) NumPatterns() int {
	return len(m.patterns)
}


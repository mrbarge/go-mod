package module

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type ImpulseTracker struct {
	title string
	version uint16
	compat uint16
	flags uint16
	special uint16
	globalVolume uint8
	mixVolume uint8
	speed uint8
	tempo uint8
	panningSeparation uint8
	pitchWheelDepth uint8
	messageLength uint16
	messageOffset uint32
	orders []uint8
	instruments []ITInstrument
	samples []ITSample
	patterns []Pattern
	Module
}

func (m *ImpulseTracker) Type() FileFormat {
	return IMPULSETRACKER
}

func (m *ImpulseTracker) Load(data []byte) (error) {
	m.title = filterNulls(string(data[4:30]))
	numOrders := binary.LittleEndian.Uint16(data[32:34])
	numInstruments := binary.LittleEndian.Uint16(data[34:36])
	numSamples := binary.LittleEndian.Uint16(data[36:38])
	//numPatterns := int(binary.LittleEndian.Uint16(data[38:40]))
	m.version = binary.LittleEndian.Uint16(data[40:42])
	m.compat  = binary.LittleEndian.Uint16(data[42:44])
	m.flags = binary.LittleEndian.Uint16(data[44:46])
	m.special = binary.LittleEndian.Uint16(data[46:48])
	m.globalVolume = data[48]
	m.mixVolume = data[49]
	m.speed = data[50]
	m.tempo = data[51]
	m.panningSeparation = data[52]
	m.pitchWheelDepth = data[53]
	m.messageLength = binary.LittleEndian.Uint16(data[54:56])
	m.messageOffset = binary.LittleEndian.Uint32(data[56:60])

	// read order list
	offset := 192
	for i := offset; i < offset+int(numOrders); i++ {
		m.orders = append(m.orders, data[i])
	}
	offset = offset + int(numOrders)

	// read instrument offsets
	instrumentOffsets := make([]uint32, 0)
	for i := 0; i < int(numInstruments); i++ {
		ioff := binary.LittleEndian.Uint32(data[offset:offset+4])
		instrumentOffsets = append(instrumentOffsets, ioff)
		offset = offset + 4
	}

	// read sample offsets
	sampleOffsets := make([]uint32, 0)
	for i := 0; i < int(numSamples); i++ {
		soff := binary.LittleEndian.Uint32(data[offset:offset+4])
		sampleOffsets = append(sampleOffsets, soff)
		offset = offset + 4
	}

	// read instruments
	for i, v := range instrumentOffsets {
		offset = int(v)
		hdr := string(data[offset:offset+4])
		if hdr != "IMPI" {
			return errors.New(fmt.Sprintf("Invalid instrument %d read at offset %d", i, offset))
		}
		offset = offset + 4
		instrument := ITInstrument{}
		instrument.filename = filterNulls(string(data[offset:offset+12]))
		offset = offset + 12
		// skip flags
		offset = offset + 16
		instrument.name = filterNulls(string(data[offset:offset+26]))
		offset = offset + 26
		// skip flags
		offset = offset +6
		// skip note sample
		offset = offset + 240
		// skip envelopes
		offset = offset + 16

		m.instruments = append(m.instruments, instrument)
	}

	// read samples
	for i, v := range sampleOffsets {
		offset = int(v)
		hdr := string(data[offset:offset+4])
		if hdr != "IMPS" {
			return errors.New(fmt.Sprintf("Invalid sample %d read at offset %d", i, offset))
		}
		offset = offset + 4

		sample := ITSample{}
		sample.filename = filterNulls(string(data[offset:offset+12]))
		offset = offset + 12

		// skip flags
		offset = offset + 4
		sample.name = filterNulls(string(data[offset:offset+26]))
		offset = offset + 26
		// skip flags
		offset = offset + 2
		sample.length = binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4
		sample.loopStart = binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4
		sample.loopEnd = binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4
		sample.c5speed = binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4
		sample.sustainStart = binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4
		sample.sustainEnd = binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4

		samplePointer := binary.LittleEndian.Uint32(data[offset:offset+4])
		offset = offset + 4

		sample.data = data[samplePointer:samplePointer+sample.length]

		m.samples = append(m.samples, sample)
	}

	return nil
}

func (m *ImpulseTracker) Play() {
	fmt.Printf("Playing IT..\n")
}

func (m *ImpulseTracker) Title() (string) {
	return m.title
}

func (m *ImpulseTracker) Instruments() []Instrument {
	r := make([]Instrument, len(m.instruments))
	for i := range m.instruments {
		r[i] = m.instruments[i]
	}
	return r
}

func (m *ImpulseTracker) Samples() []Sample {
	r := make([]Sample, len(m.samples))
	for i := range m.samples {
		r[i] = m.samples[i]
	}
	return r
}

func (m *ImpulseTracker) NumPatterns() int {
	return len(m.patterns)
}

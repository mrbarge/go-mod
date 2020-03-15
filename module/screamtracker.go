package module

import (
	"encoding/binary"
	"errors"
	"fmt"
)

//	0	char[28]	title	Song title, must be null-terminated
//	28	UINT8	sig1	Signature byte, always 0x1A
//	29	UINT8	type	Song type, always 0x10 for S3M
//	30	UINT16LE	reserved	Always 0x0000
//	32	UINT16LE	orderCount	Number of entries in the order table, should be even
//	34 	UINT16LE	instrumentCount	Number of instruments in the song
//	36	UINT16LE	patternPtrCount	Number of pattern parapointers in the song
//	38	UINT16LE	flags	See below
//	40	UINT16LE	trackerVersion	upper four bits is tracker ID, lower 12 bits is tracker version
//	42	UINT16LE	sampleType	1=signed samples [deprecated], 2=unsigned samples
//	44	char[4]	sig2	Signature: "SCRM"
//	48	UINT8	globalVolume
//	49	UINT8	initialSpeed	Frames per row, can be changed later with A command
//	50	UINT8	initialTempo	Frames per second, can be changed later with T command
//	51	UINT8	masterVolume	bit 7: 1=stereo, 0=mono, bits 6-0: volume
//	52	UINT8	ultraClickRemoval	Number of channels to use for click removal on real GUS hardware
//	53	UINT8	defaultPan	252=read pan values in header, anything else ignores pan values in header
//	54	BYTE[8]	reserved	Unused, some trackers store data here
//	62	UINT16LE	ptrSpecial	Parapointer to additional data, if flags has bit 7 set
//	64	UINT8[32]	channelSettings	See below
//	96	UINT8[orderCount]	orderList	Which order patterns are played in
//	98	UINT16LE[instrumentCount]	ptrInstruments	List of parapointers to each instrument's data
//	100	UINT16LE[patternPtrCount]	ptrPatterns	List of parapointers to each pattern's data

type ScreamTracker struct {
	title string
	masterVolume uint8
	isStereo bool
	speed uint8
	tempo uint8
	volume uint8
	signature string
	sampleType SampleType
	samples []STSample
	patterns []Pattern
	orderList []uint8
	Module
}

type SampleType int
const (
	SIGNED = iota + 1
	UNSIGNED
)

func (m *ScreamTracker) Type() FileFormat {
	return SCREAMTRACKER
}

func (m *ScreamTracker) Load(data []byte) (error) {
	m.title = filterNulls(string(data[0:28]))

	orderCount := binary.LittleEndian.Uint16(data[32:34])
	instrumentCount := binary.LittleEndian.Uint16(data[34:36])
	patternPtrCount := binary.LittleEndian.Uint16(data[36:38])
	//flags := binary.LittleEndian.Uint32(data[38:40])
	m.sampleType = SampleType(binary.LittleEndian.Uint16(data[42:44]))
	m.signature = string(data[44:48])
	m.volume = uint8(data[48])
	m.speed = uint8(data[49])
	m.tempo = uint8(data[50])
	m.isStereo = ((data[51] & (1 << 7)) != 0)
	m.masterVolume = uint8((data[51] << 1) >> 1)

	// order list loading time
	for i := 96; i < 96+int(orderCount); i++ {
		patternNum := uint8(data[i])
		m.orderList = append(m.orderList, patternNum)
	}

	// instrument loading time
	startOffset := 96+int(orderCount)
	for i := 0; i < int(instrumentCount); i++ {
		// offset is parapointer, so multiply by 16
		instrumentOffset := int(binary.LittleEndian.Uint16(data[startOffset+(i*2):startOffset+2+(i*2)]) * 16)
		instrumentType := uint8(data[instrumentOffset])
		if instrumentType == 0 {
			// empty sample
			continue
		} else if instrumentType != 1 {
			return errors.New(fmt.Sprintf("Unsupported sample type %d at offset %d", instrumentType, instrumentOffset))
		}
		instrumentOffset = instrumentOffset + 1

		sample := STSample{}

		sample.filename = filterNulls(string(data[instrumentOffset:instrumentOffset+12]))
		instrumentOffset = instrumentOffset + 12

		sampleHighOffset := data[instrumentOffset]
		instrumentOffset += 1
		// another parapointer, so *16
		sample.sampleOffset = int(uint(sampleHighOffset << 16) | uint(data[instrumentOffset+1]) << 8 | uint(data[instrumentOffset]))*16
		instrumentOffset += 2

		//instrumentOffset := sampleHighOffset
		sample.length = binary.LittleEndian.Uint32(data[instrumentOffset:instrumentOffset+4])
		instrumentOffset += 4
		//instrumentOffset = instrumentOffset + 4

		sample.loopStart = binary.LittleEndian.Uint32(data[instrumentOffset:instrumentOffset+4])
		instrumentOffset += 4

		sample.loopEnd = binary.LittleEndian.Uint32(data[instrumentOffset:instrumentOffset+4])
		instrumentOffset += 4

		sample.volume = data[instrumentOffset]
		instrumentOffset += 2

		sample.pack = data[instrumentOffset]
		instrumentOffset += 1

		instrumentOffset += 1	//flags

		sample.c2spd = binary.LittleEndian.Uint32(data[instrumentOffset:instrumentOffset+4])
		instrumentOffset += 16 	// skip internal

		sample.name = filterNulls(string(data[instrumentOffset:instrumentOffset+28]))
		instrumentOffset += 28

		// validate we ended up at the right spot
		if string(data[instrumentOffset:instrumentOffset+4]) != "SCRS" {
			return errors.New(fmt.Sprintf("Signature missing or corrupt for sample %d", i))
		}

		// lastly set the sample data
		sample.data = data[sample.sampleOffset: sample.sampleOffset+int(sample.length)]

		m.samples = append(m.samples, sample)
	}
	for i := 0; i < int(patternPtrCount); i++ {
		pattern := Pattern{}
		m.patterns = append(m.patterns, pattern)
	}

	return nil
}

func (m *ScreamTracker) Play() {
}

func (m *ScreamTracker) Title() (string) {
	return m.title
}

func (m *ScreamTracker) Instruments() []Instrument {
	return []Instrument{}
}

func (m *ScreamTracker) Samples() []Sample {
	r := make([]Sample, len(m.samples))
	for i := range m.samples {
		r[i] = m.samples[i]
	}
	return r
}

func (m *ScreamTracker) NumPatterns() int {
	return len(m.patterns)
}


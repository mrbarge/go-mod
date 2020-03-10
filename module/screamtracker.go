package module

import (
	"encoding/binary"
	"errors"
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
//	98	UINT8[orderCount]	orderList	Which order patterns are played in
//	99	UINT16LE[instrumentCount]	ptrInstruments	List of parapointers to each instrument's data
//	101	UINT16LE[patternPtrCount]	ptrPatterns	List of parapointers to each pattern's data

type ScreamTracker struct {
	title string
	Module
}

func (m *ScreamTracker) Type() FileFormat {
	return SCREAMTRACKER
}

func (m *ScreamTracker) Load(data []byte) (error) {
	m.title = string(data[0:28])

	//numOrders := binary.LittleEndian.Uint32(data[32:33])
	//numInstruments :=binary.LittleEndian.Uint32(data[34:36])

}

func (m *ScreamTracker) Play() {
}

func (m *ScreamTracker) Title() (string) {
	return m.title
}

func (m *ScreamTracker) Instruments() []Instrument {
	return []Instrument{}
}

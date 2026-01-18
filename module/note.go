package module
import (
	"errors"
	"fmt"
	"strings"
)

type Note struct {
	key int
	instrument int
	period int
	effect int
	parameter int
}

// C-0   C#0   D-0   D#0   E-0   F-0   F#0   G-0   G#0   A-1  A#1  B-1 */
// 1712, 1616, 1524, 1440, 1356, 1280, 1208, 1140, 1076, 1016, 960, 907,
// 856,  808,  762,  720,  678,  640,  604,  570,  538,  508, 480, 453,
// 428,  404,  381,  360,  339,  320,  302,  285,  269,  254, 240, 226,
// 214,  202,  190,  180,  170,  160,  151,  143,  135,  127, 120, 113,
// 107,  101,   95,   90,   85,   80,   75,   71,   67,   63,  60,  56,

var periodLookup = []int {
	//C   C#    D     D#    E     F     F#    G     G#    A     A#   B
	1712, 1616, 1524, 1440, 1356, 1280, 1208, 1140, 1076, 1016, 960, 907, // 0ctave 0
    856,  808,  762,  720,  678,  640,  604,  570,  538,  508, 480, 453,  // 0ctave 1
    428,  404,  381,  360,  339,  320,  302,  285,  269,  254, 240, 226,  // 0ctave 2
    214,  202,  190,  180,  170,  160,  151,  143,  135,  127, 120, 113,  // 0ctave 3
    107,  101,   95,   90,   85,   80,   75,   71,   67,   63,  60,  56,  // 0ctave 4
}

var notes = []string {
	"C","C#","D","D#","E","F","F#","G","G#","A","A#","B",
}

//Info for each note:
//
//_____byte 1_____   byte2_    _____byte 3_____   byte4_
///                \ /      \  /                \ /      \
//0000          0000-00000000  0000          0000-00000000
//
//Upper four    12 bits for    Lower four    Effect command.
//bits of sam-  note period.   bits of sam-
//ple number.                  ple number.

func (n *Note) Load(data []byte) error {

	periodUpper := int(data[0] & 0x0F)
	periodLower := int(data[1])
	n.period = (periodUpper << 8) | periodLower

	instrumentUpper := int(data[0] & 0xF0)
	instrumentLower := int((data[2] & 0xF0) >> 4)
	n.instrument = instrumentUpper | instrumentLower

	n.effect = int(data[2] & 0x0F)
	n.parameter = int(data[3])

	return nil
}

func (n *Note) ToString() (string, error) {

	// First find the octave
	octave := 0
	for i := 0; i < len(periodLookup); i+=12 {
		if (n.period > periodLookup[i] ) {
			break
		} else {
			octave++
		}
	}
	octave--
	// Now find the note within that octave
	offset := octave * 12
	position := 0
	for i := offset; i < offset+12; i++ {
		if (periodLookup[i] == n.period) {
			position = i
			break
		}
	}

	if (position == -1) {
		return "",errors.New("unable to find note for period")
	}

	note,err := periodToString(position,octave)
	if (err != nil) {
		return "",err
	} else {
		return note,nil
	}
}

func periodToString(periodPos int, octave int) (string,error) {
	if (periodPos < 0) || (periodPos > len(periodLookup)) {
		return "",errors.New("Invalid period to convert to string")
	}  else {
		// Get the note pitch
		noteLookupPos := periodPos % 12
		pitch := notes[noteLookupPos]
		if (strings.HasSuffix(pitch, "#")) {
			return fmt.Sprintf("%s%d", pitch, octave),nil
		} else {
			return fmt.Sprintf("%s-%d", pitch, octave),nil
		}
	}
}

// Getters for exporting note data
func (n *Note) Period() int {
	return n.period
}

func (n *Note) Instrument() int {
	return n.instrument
}

func (n *Note) Effect() int {
	return n.effect
}

func (n *Note) Parameter() int {
	return n.parameter
}

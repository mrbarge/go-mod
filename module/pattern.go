package module
import "errors"

type Pattern struct {
	rows int8
	channels int8
	data []byte
}

func (p *Pattern) NumChannels() int8  {
	return p.channels
}

func (p *Pattern) NumRows() int8  {
	return p.rows
}

func (p *Pattern) GetNote(row int8, channel int8) (Note,error) {
	if (channel < 0 || channel > p.channels) {
		return Note{},errors.New("Invalid channel")
	}
	if (row < 0 || row > p.rows) {
		return Note{},errors.New("Invalid row")
	}

	// Calculate size of an individual row, each row has numChannels notes, each note is 4 bytes
	rowsize := p.channels * 4
	// Calculate offset into pattern data to find the note
	noteOffset := int(rowsize * row + (channel*4))

	// Sanity check of offset
	if (noteOffset > len(p.data)) {
		return Note{},errors.New("Unable to locate note offset.")
	}
	n := Note{}
	err := n.Load(p.data[noteOffset:noteOffset+4])

	// Number of
	return n,err
}


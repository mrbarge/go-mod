package module

import (
	"errors"
)

type Pattern struct {
	numChannels int8
	rows        []Row
	data        []byte
}

func (p *Pattern) NumChannels() int  {
	return int(p.numChannels)
}

func (p *Pattern) NumRows() int  {
	return len(p.rows)
}

func (p *Pattern) GetNote(row int, channel int) (Note,error) {
	if (channel < 0 || channel > p.NumChannels()) {
		return Note{},errors.New("Invalid channel")
	}
	if (row < 0 || row > p.NumRows()) {
		return Note{},errors.New("Invalid row")
	}

	// Calculate size of an individual row, each row has numChannels notes, each note is 4 bytes
	rowsize := p.NumChannels() * 4
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

func (p *Pattern) GetRow(row int) (*Row,error) {
	if (row < 0 || row > p.NumRows()) {
		return nil,errors.New("Invalid row index.")
	}
	return &p.rows[row],nil
}

func (p *Pattern) SetRow(idx int, r Row) error {
	if (idx < 0 || idx > len(p.rows)) {
		return errors.New("Invalid row index.")
	} else {
		p.rows[idx] = r
		return nil
	}
}
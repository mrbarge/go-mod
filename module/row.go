package module

type Row struct {
	notes []Note
}

// Notes returns the notes in this row
func (r *Row) Notes() []Note {
	return r.notes
}


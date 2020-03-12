package module

type STInstrument struct {
	name string
	filename string
	data []byte
	Instrument
}

func (i STInstrument) Name() string {
	return i.name
}

func (i STInstrument) Filename() string {
	return i.filename
}

func (i STInstrument) Load(data []byte) (error) {
	return nil
}

func (i STInstrument) Data() []byte {
	return i.data
}

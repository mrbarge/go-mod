package module

type STInstrument struct {
	name string
	filename string
	length uint32
	loopStart uint32
	loopEnd uint32
	sampleOffset int
	volume uint8
	pack uint8
	flags uint8
	c2spd uint32
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

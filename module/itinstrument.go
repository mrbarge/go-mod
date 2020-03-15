package module

type ITInstrument struct {
	name string
	filename string
	data []byte
	Instrument
}

func (i ITInstrument) Name() string {
	return i.name
}

func (i ITInstrument) Filename() string {
	return i.name
}

func (i ITInstrument) Load(data []byte) (error) {
	return nil
}

func (i ITInstrument) Data() []byte {
	return i.data
}

package module

type FTInstrument struct {
	name string
	data []byte
	Instrument
}

func (i FTInstrument) Name() string {
	return i.name
}

func (i FTInstrument) Filename() string {
	return i.name
}

func (i FTInstrument) Load(data []byte) (error) {
	return nil
}

func (i FTInstrument) Data() []byte {
	return i.data
}

package module

type FTInstrument struct {
	name string
	samples []FTSample
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
	if len(i.samples) > 0 {
		return i.samples[0].Data()
	} else {
		return nil
	}
}

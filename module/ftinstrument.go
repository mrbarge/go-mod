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

func (i FTInstrument) Samples() []FTSample {
	return i.samples
}

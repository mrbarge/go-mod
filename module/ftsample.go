package module

type FTSample struct {
	name string
	loopStart uint32
	loopEnd uint32
	length uint32
	volume uint8
	finetune uint8
	sampleType uint8
	panning uint8
	relativeNote uint8
	dataType uint8
	data []byte
}

func (i FTSample) Name() string {
	return i.name
}

func (i FTSample) Filename() string {
	return i.name
}

func (i FTSample) Data() []byte {
	return i.data
}

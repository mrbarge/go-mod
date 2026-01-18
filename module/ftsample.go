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

func (i FTSample) LoopStart() uint32 {
	return i.loopStart
}

func (i FTSample) LoopEnd() uint32 {
	return i.loopEnd
}

func (i FTSample) Length() uint32 {
	return i.length
}

func (i FTSample) Volume() uint8 {
	return i.volume
}

func (i FTSample) Finetune() uint8 {
	return i.finetune
}

func (i FTSample) SampleType() uint8 {
	return i.sampleType
}

func (i FTSample) Panning() uint8 {
	return i.panning
}

func (i FTSample) RelativeNote() uint8 {
	return i.relativeNote
}

func (i FTSample) DataType() uint8 {
	return i.dataType
}

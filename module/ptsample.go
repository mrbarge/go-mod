package module

type PTSample struct {
	name string
	length int64
	finetune int8
	volume int8
	repeatOffset uint16
	repeatLength uint16
	data []byte
	Sample
}

func (i PTSample) Name() string {
	return i.name
}

func (i PTSample) Filename() string {
	return i.name
}

func (i PTSample) Length() int64 {
	return i.length
}

func (i PTSample) Data() []byte {
	return i.data
}

func (i PTSample) Finetune() int8 {
	return i.finetune
}

func (i PTSample) Volume() int8 {
	return i.volume
}

func (i PTSample) RepeatOffset() uint16 {
	return i.repeatOffset
}

func (i PTSample) RepeatLength() uint16 {
	return i.repeatLength
}
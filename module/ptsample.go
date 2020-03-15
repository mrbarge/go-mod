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
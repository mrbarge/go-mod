package module

type STSample struct {
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
	Sample
}

func (i STSample) Name() string {
	return i.name
}

func (i STSample) Filename() string {
	return i.filename
}

func (i STSample) Load(data []byte) (error) {
	return nil
}

func (i STSample) Data() []byte {
	return i.data
}

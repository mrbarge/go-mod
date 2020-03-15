package module

type ITSample struct {
	name string
	filename string
	length uint32
	loopStart uint32
	loopEnd uint32
	sustainStart uint32
	sustainEnd uint32
	c5speed uint32
	data []byte
}

func (i ITSample) Name() string {
	return i.name
}

func (i ITSample) Filename() string {
	return i.name
}

func (i ITSample) Load(data []byte) (error) {
	return nil
}

func (i ITSample) Data() []byte {
	return i.data
}

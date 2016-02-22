package module

type Pattern struct {
	channels int8
	data []byte
}

func (p *Pattern) NumChannels() int8  {
	return p.channels
}

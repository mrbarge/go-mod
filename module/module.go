package module

import (
	"io/ioutil"
	"os"
)

type FileFormat int
const (
	PROTRACKER = iota
	SCREAMTRACKER
	FASTTRACKER
	IMPULSETRACKER
)

type Module interface {
	Load(data []byte) error
	Title() string
	Play()
	Type() FileFormat
	Instruments() []Instrument
	Samples() []Sample
	NumPatterns() int
}

type Sample interface {
	Name() string
	Filename() string
	Data() []byte
}

type Instrument interface {
	Name() string
	Filename() string
}

func Load(modFile string) (Module,error) {

	f, err := os.Open(modFile)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	f.Close()

	// Determine file type
	if (string(data[44:48]) == "SCRM") {
		m := &ScreamTracker{}
		err = m.Load(data)
		return m, err
	}

	if (string(data[0:17]) == "Extended Module: ") {
		m := &FastTracker{}
		err = m.Load(data)
		return m, err
	}

	if (string(data[0:4]) == "IMPM") {
		m := &ImpulseTracker{}
		err = m.Load(data)
		return m, err
	}

	m := &ProTracker{}
	err = m.Load(data)

	return m, err

}

package module

import (
	"archive/zip"
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
	Filename() string
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

	// assume a zipfile
	zf, err := zip.OpenReader(modFile)
	var data []byte
	if err == nil {
		// TODO: > 1 file per zip
		for _, file := range zf.File {
			fc, err := file.Open()
			defer fc.Close()
			if err != nil {
				return nil, err
			}
			data, err = ioutil.ReadAll(fc)
			if err != nil {
				return nil, err
			}
		}
	} else {
		f, err := os.Open(modFile)
		defer f.Close()
		data, err = ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
	}

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

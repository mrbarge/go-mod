package module

import (
	"io/ioutil"
	"os"
)

const (
	MOD = iota
	SCREAMTRACKER
	FASTTRACKER
)

type Module interface {
	Load(data []byte) (error)
	Title() (string)
	Play()
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

	m := &ProTracker{}
	err = m.Load(data)

	return m, err

}

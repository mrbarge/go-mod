package module
import (
	"fmt"
	"errors"
)

type FastTracker struct {
	title string
	Module
}

func (m *FastTracker) Load(data []byte) (error) {
	return errors.New("Unsupported")
}
func (m *FastTracker) Play() {
	fmt.Printf("Playing FT..\n")
}
func (m *FastTracker) Title() (string) {
	return m.title
}


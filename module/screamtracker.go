package module
import "errors"

type ScreamTracker struct {
	title string
	Module
}

func (m *ScreamTracker) Load(data []byte) (error) {
	return errors.New("Unsupported")
}
func (m *ScreamTracker) Play() {
}
func (m *ScreamTracker) Title() (string) {
	return m.title
}


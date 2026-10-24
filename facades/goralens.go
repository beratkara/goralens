package facades

import "github.com/beratkara/goralens"

func Goralens() *goralens.Recorder {
	return goralens.Instance()
}

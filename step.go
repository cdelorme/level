package level

import "os"

type step interface {
	Walk(string, os.FileInfo, error) error
	Execute() ([][]string, error)
}

var steps []step

func init() {
	steps = append(steps, &Data{})
}

package level6

import "os"

type step interface {
	Walk(string, os.FileInfo, error) error
	Execute() ([][]string, error)
}

var steps []step

func init() {
	steps = append(steps, &Data{})
	steps = append(steps, &Content{})
	steps = append(steps, &Image{})
	steps = append(steps, &Audio{})
	steps = append(steps, &Video{})
}

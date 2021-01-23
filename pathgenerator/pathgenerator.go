package pathgenerator

import (
	"gonum.org/v1/gonum/spatial/r3"
)

type PathGenerator interface {
	GeneratePath(from []r3.Vec, dest []r3.Vec) <-chan [][]r3.Vec
	Stop()
}

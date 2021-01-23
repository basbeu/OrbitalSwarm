package mapping

import (
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	staredVal = 1
	primedVal = 2
	coverVal  = 1
)

// https://en.wikipedia.org/wiki/Hungarian_algorithm
// https://www.researchgate.net/publication/290437481_Tutorial_on_Implementation_of_Munkres'_Assignment_Algorithm

type TargetsMapper interface {
	MapTargets(initials []r3.Vec, targets []r3.Vec) []r3.Vec
}

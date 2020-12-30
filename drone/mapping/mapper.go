package mapping

import (
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/spatial/r3"
)

// https://en.wikipedia.org/wiki/Hungarian_algorithm
// https://www.researchgate.net/publication/290437481_Tutorial_on_Implementation_of_Munkres'_Assignment_Algorithm

type TargetsMapper interface {
	MapTargets(initials []r3.Vec, targets []r3.Vec) map[string]r3.Vec
}

type hungarianMapper struct {
}

func NewHungarianMapper() *hungarianMapper {
	return &hungarianMapper{}
}

func (m *hungarianMapper) MapTargets(initials []r3.Vec, targets []r3.Vec) map[string]r3.Vec {
	//TODO
	return make(map[string]r3.Vec)
}

func (m *hungarianMapper) initMatrix(initials []r3.Vec, targets []r3.Vec) *mat.Dense {

	return nil
}

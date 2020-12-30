package mapping

import (
	"go.dedis.ch/cs438/orbitalswarm/utils"
	"gonum.org/v1/gonum/mat"
)

// https://en.wikipedia.org/wiki/Hungarian_algorithm
// https://www.researchgate.net/publication/290437481_Tutorial_on_Implementation_of_Munkres'_Assignment_Algorithm

type TargetsMapper interface {
	MapTargets(initials []utils.Vec3d, targets []utils.Vec3d) map[string]utils.Vec3d
}

type hungarianMapper struct {
}

func NewHungarianMapper() *hungarianMapper {
	return &hungarianMapper{}
}

func (m *hungarianMapper) MapTargets(initials []utils.Vec3d, targets []utils.Vec3d) map[string]utils.Vec3d {
	//TODO
	return make(map[string]utils.Vec3d)
}

func (m *hungarianMapper) initMatrix(initials []utils.Vec3d, targets []utils.Vec3d) *mat.Dense {

	return nil
}

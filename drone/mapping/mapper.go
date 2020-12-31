package mapping

import (
	"math"

	"gonum.org/v1/gonum/floats"

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
	if len(initials) != len(targets) {
		panic("Number of drones not equal to number of targets")
	}

	//TODO
	return make(map[string]r3.Vec)
}

// iniMatrix creates the nxn cost matrix
func (m *hungarianMapper) initMatrix(initials []r3.Vec, targets []r3.Vec) *mat.Dense {
	n := len(initials)
	matrix := mat.NewDense(n, n, nil)

	for i, drone := range initials {
		for j, target := range targets {
			dist := r3.Norm(drone.Sub(target))
			matrix.Set(i, j, math.Floor(dist))
		}
	}

	return matrix
}

// step01 : For each row of the matrix, find the smallest element and subtract it from every element in its row. Go to Step 2
func (m *hungarianMapper) step01(matrix *mat.Dense) {
	r, _ := matrix.Dims()
	for i := 0; i < r; i++ {
		row := matrix.RawRowView(i)
		min := floats.Min(row)
		floats.AddConst(-min, row)
	}
}

func (m *hungarianMapper) step02(matrix *mat.Dense) {
}

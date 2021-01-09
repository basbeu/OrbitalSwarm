package mapping

import (
	"math"

	"gonum.org/v1/gonum/floats"

	"gonum.org/v1/gonum/mat"
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

// step02 : Find a zero (Z) in the resulting matrix. If there is no starred zero in its row or column, star Z. Repeat for each element in the matrix. Go to Step 3. Return mask matrix
func (m *hungarianMapper) step02(matrix *mat.Dense) *mat.Dense {
	r, c := matrix.Dims()
	mask := mat.NewDense(r, c, nil)
	rowCover := mat.NewVecDense(r, nil)
	colCover := mat.NewVecDense(c, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if matrix.At(i, j) == 0 &&
				rowCover.AtVec(i) != coverVal &&
				colCover.AtVec(j) != coverVal {
				mask.Set(i, j, staredVal)
				rowCover.SetVec(i, coverVal)
				colCover.SetVec(j, coverVal)
			}
		}
	}
	return mask
}

// step03 : Cover each column containing a starred zero. If K columns are covered, the starred zeros describe a complete set of unique assignments. In this case, Go to DONE, otherwise, Go to Step 4.
func (m *hungarianMapper) step03(mask *mat.Dense) bool {
	r, c := mask.Dims()
	colCover := mat.NewVecDense(c, nil)
	colCount := 0
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if mask.At(i, j) == staredVal &&
				colCover.AtVec(j) != coverVal {
				colCover.SetVec(j, coverVal)
				colCount++
			}
		}
	}

	return colCount >= c
}

func (m *hungarianMapper) step04(matrix, mask *mat.Dense, rowCover, colCover *mat.VecDense) (int, int, bool) {
	done := false
	for !done {
		r, c, found := m.findUncoveredZero(matrix, rowCover, colCover)
		if !found {
			done = true
		} else {
			mask.Set(r, c, primedVal)
			cStar, found := m.findStarredInRow(mask, r)
			if found {
				rowCover.SetVec(r, coverVal)
				colCover.SetVec(cStar, 0)
			} else {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

func (m *hungarianMapper) findUncoveredZero(matrix *mat.Dense, rowCover, colCover *mat.VecDense) (int, int, bool) {
	r, c := matrix.Dims()
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if matrix.At(i, j) == 0 &&
				rowCover.AtVec(i) != coverVal &&
				colCover.AtVec(j) != coverVal {
				return i, j, true
			}
		}
	}
	return 0, 0, false
}

func (m *hungarianMapper) findStarredInRow(mask *mat.Dense, row int) (int, bool) {
	_, c := mask.Dims()
	for j := 0; j < c; j++ {
		if mask.At(row, j) == staredVal {
			return j, true
		}
	}

	return 0, false
}

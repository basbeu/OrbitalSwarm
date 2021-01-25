package mapping

import (
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/spatial/r3"
)

type hungarianMapper struct {
}

func NewHungarianMapper() *hungarianMapper {
	return &hungarianMapper{}
}

func (m *hungarianMapper) MapTargets(initials []r3.Vec, targets []r3.Vec) []r3.Vec {
	if len(initials) != len(targets) {
		panic("Number of drones not equal to number of targets")
	}

	matrix := m.initMatrix(initials, targets)
	mask := m.computeAssignment(matrix)

	return m.decodeAssignement(targets, mask)
}

// iniMatrix creates the nxn cost matrix
func (m *hungarianMapper) initMatrix(initials []r3.Vec, targets []r3.Vec) *mat.Dense {
	n := len(initials)
	matrix := mat.NewDense(n, n, nil)

	for i, drone := range initials {
		for j, target := range targets {
			dist := math.Pow(r3.Norm(drone.Sub(target))+25, 2)
			matrix.Set(i, j, dist)
		}
	}

	return matrix
}

func (m *hungarianMapper) computeAssignment(matrix *mat.Dense) *mat.Dense {
	var mask *mat.Dense
	var rowCover *mat.VecDense
	var colCover *mat.VecDense
	var rPrimed, cPrimed int

	r, c := matrix.Dims()
	rowCover = mat.NewVecDense(r, nil)
	colCover = mat.NewVecDense(c, nil)

	done := false
	step := 1
	for !done {
		switch step {
		case 1:
			m.step01(matrix)
			step = 2
		case 2:
			mask = m.step02(matrix)
			step = 3
		case 3:
			done = m.step03(mask, colCover)
			step = 4
		case 4:
			var found bool
			rPrimed, cPrimed, found = m.step04(matrix, mask, rowCover, colCover)
			if found {
				step = 5
			} else {
				step = 6
			}
		case 5:
			m.step05(mask, rowCover, colCover, rPrimed, cPrimed)
			step = 3
		case 6:
			m.step06(matrix, rowCover, colCover)
			step = 4
		}
	}
	return mask
}

func (m *hungarianMapper) decodeAssignement(targets []r3.Vec, mask *mat.Dense) []r3.Vec {
	r, c := mask.Dims()
	res := make([]r3.Vec, r)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if mask.At(i, j) == staredVal {
				res[i] = targets[j]
			}
		}
	}
	return res
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
func (m *hungarianMapper) step03(mask *mat.Dense, colCover *mat.VecDense) bool {
	r, c := mask.Dims()
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

// step04 :   Find a noncovered zero and prime it. If there is no starred zero in the row containing this primed zero, Go to Step 5. Otherwise, cover this row and uncover the column containing the starred zero. Continue in this manner until there are no uncovered zeros left. Save the smallest uncovered value and Go to Step 6.
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

// step05 :  Construct a series of alternating primed and starred zeros as follows. Let Z0 represent the uncovered primed zero found in Step 4. Let Z1 denote the starred zero in the column of Z0 (if any). Let Z2 denote the primed zero in the row of Z1 (there will always be one). Continue until the series terminates at a primed zero that has no starred zero in its column. Unstar each starred zero of the series, star each primed zero of the series, erase all primes and uncover every line in the matrix. Return to Step 3
func (m *hungarianMapper) step05(mask *mat.Dense, rowCover, colCover *mat.VecDense, rPrimed, cPrimed int) {
	r, c := mask.Dims()

	path := mat.NewDense(r*c, 2, nil)
	pathCounter := 0
	path.Set(0, 0, float64(rPrimed))
	path.Set(0, 1, float64(cPrimed))
	done := false
	for !done {
		cStarred := int(path.At(pathCounter, 1))
		rStarred, found := m.findStarredInCol(mask, cStarred)
		done = !found
		if found {
			pathCounter++
			path.Set(pathCounter, 0, float64(rStarred))
			path.Set(pathCounter, 1, float64(cStarred))

			rPrimed = rStarred
			cPrimed, _ = m.findPrimedInRow(mask, rPrimed)
			pathCounter++
			path.Set(pathCounter, 0, float64(rPrimed))
			path.Set(pathCounter, 1, float64(cPrimed))
		}
	}

	for i := 0; i <= pathCounter; i++ {
		rPath := int(path.At(i, 0))
		cPath := int(path.At(i, 1))

		if mask.At(rPath, cPath) == staredVal {
			mask.Set(rPath, cPath, 0)
		} else {
			mask.Set(rPath, cPath, staredVal)
		}
	}

	// Uncover every lines
	rowCover.Zero()
	colCover.Zero()

	// Erase primed
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if mask.At(i, j) == primedVal {
				mask.Set(i, j, 0)
			}
		}
	}
}

func (m *hungarianMapper) findStarredInCol(mask *mat.Dense, col int) (int, bool) {
	r, _ := mask.Dims()
	for i := 0; i < r; i++ {
		if mask.At(i, col) == staredVal {
			return i, true
		}
	}

	return 0, false
}

func (m *hungarianMapper) findPrimedInRow(mask *mat.Dense, row int) (int, bool) {
	_, c := mask.Dims()
	for j := 0; j < c; j++ {
		if mask.At(row, j) == primedVal {
			return j, true
		}
	}

	return 0, false
}

// step06 :    Add the value found in Step 4 to every element of each covered row, and subtract it from every element of each uncovered column. Return to Step 4 without altering any stars, primes, or covered lines.
func (m *hungarianMapper) step06(matrix *mat.Dense, rowCover, colCover *mat.VecDense) {
	smallest := m.findSmallestUncovered(matrix, rowCover, colCover)
	r, c := matrix.Dims()
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if rowCover.AtVec(i) == coverVal {
				matrix.Set(i, j, matrix.At(i, j)+smallest)
			}
			if colCover.AtVec(j) != coverVal {
				matrix.Set(i, j, matrix.At(i, j)-smallest)
			}
		}
	}
}

func (m *hungarianMapper) findSmallestUncovered(matrix *mat.Dense, rowCover, colCover *mat.VecDense) float64 {
	r, c := matrix.Dims()
	min := math.MaxFloat64
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if rowCover.AtVec(i) == 0 && colCover.AtVec(j) == 0 {
				min = math.Min(matrix.At(i, j), min)
			}
		}
	}
	return min
}

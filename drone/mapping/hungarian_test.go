package mapping

import (
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestInitMatrix(t *testing.T) {
	mapper := NewHungarianMapper()

	initialPos := []r3.Vec{
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 2, Y: 0, Z: 0},
	}
	targetPos := []r3.Vec{
		r3.Vec{X: 3, Y: 0, Z: 0},
		r3.Vec{X: 4, Y: 0, Z: 0},
	}

	expectedMatrix := mat.NewDense(2, 2, []float64{
		2, 3,
		1, 2,
	})

	matrix := mapper.initMatrix(initialPos, targetPos)

	require.True(t, mat.Equal(expectedMatrix, matrix))

	initialPos = []r3.Vec{
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 0, Y: 2, Z: 0},
		r3.Vec{X: 0, Y: 0, Z: 3},
	}
	targetPos = []r3.Vec{
		r3.Vec{X: 1, Y: 1, Z: 1},
		r3.Vec{X: 2, Y: 2, Z: 2},
		r3.Vec{X: 3, Y: 3, Z: 3},
	}

	expectedMatrix = mat.NewDense(3, 3, []float64{
		1, 3, 4,
		1, 2, 4,
		2, 3, 4,
	})

	matrix = mapper.initMatrix(initialPos, targetPos)

	require.True(t, mat.Equal(expectedMatrix, matrix))
}

func TestStep01(t *testing.T) {
	mapper := NewHungarianMapper()
	costMatrix := mat.NewDense(3, 3, []float64{
		1, 23, 45,
		34, 17, 12,
		104, 56, 64,
	})

	expectedMatrix := mat.NewDense(3, 3, []float64{
		0, 22, 44,
		22, 5, 0,
		48, 0, 8,
	})

	mapper.step01(costMatrix)
	require.True(t, mat.Equal(expectedMatrix, costMatrix))
}

func TestStep02(t *testing.T) {
	//TestCase 1
	mapper := NewHungarianMapper()
	step1Matrix := mat.NewDense(3, 3, []float64{
		0, 22, 44,
		22, 5, 0,
		48, 0, 8,
	})

	expectedMatrix := mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 0, 1,
		0, 1, 0,
	})

	mask := mapper.step02(step1Matrix)
	require.True(t, mat.Equal(expectedMatrix, mask))

	// TestCase 2
	step1Matrix = mat.NewDense(3, 3, []float64{
		0, 22, 0,
		22, 5, 0,
		48, 18, 0,
	})

	expectedMatrix = mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 0, 1,
		0, 0, 0,
	})

	mask = mapper.step02(step1Matrix)
	require.True(t, mat.Equal(expectedMatrix, mask))

	// TestCase 3
	step1Matrix = mat.NewDense(3, 3, []float64{
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
	})

	expectedMatrix = mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	})

	mask = mapper.step02(step1Matrix)
	require.True(t, mat.Equal(expectedMatrix, mask))
}

func TestStep03(t *testing.T) {
	//TestCase 1
	mapper := NewHungarianMapper()
	maskMatrix := mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 0, 1,
		0, 1, 0,
	})

	require.True(t, mapper.step03(maskMatrix))

	//TestCase 2
	maskMatrix = mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 0, 1,
		0, 0, 0,
	})

	require.True(t, !mapper.step03(maskMatrix))
}

func TestFindUncoveredZero(t *testing.T) {
	mapper := NewHungarianMapper()

	//TestCase 1
	matrix := mat.NewDense(3, 3, []float64{
		0, 22, 44,
		22, 5, 0,
		48, 0, 8,
	})
	rowCover := mat.NewVecDense(3, []float64{1, 0, 0})
	colCover := mat.NewVecDense(3, []float64{0, 1, 0})

	r, c, found := mapper.findUncoveredZero(matrix, rowCover, colCover)

	require.True(t, found)
	require.Equal(t, 1, r)
	require.Equal(t, 2, c)

	//TestCase 2
	rowCover = mat.NewVecDense(3, []float64{1, 1, 0})
	colCover = mat.NewVecDense(3, []float64{0, 1, 0})

	r, c, found = mapper.findUncoveredZero(matrix, rowCover, colCover)

	require.True(t, !found)
}

func TestFindStarredInRow(t *testing.T) {
	mapper := NewHungarianMapper()

	mask := mat.NewDense(3, 3, []float64{
		0, 0, 0,
		0, 1, 0,
		1, 0, 1,
	})

	//TestCase 1
	c, found := mapper.findStarredInRow(mask, 0)
	require.True(t, !found)

	//TestCase 2
	c, found = mapper.findStarredInRow(mask, 1)
	require.True(t, found)
	require.Equal(t, c, 1)

	//TestCase 3
	c, found = mapper.findStarredInRow(mask, 2)
	require.True(t, found)
	require.Equal(t, c, 0)
}

func TestStep04(t *testing.T) {
	//TestCase 1
	mapper := NewHungarianMapper()
	matrix := mat.NewDense(3, 3, []float64{
		0, 22, 44,
		22, 5, 0,
		48, 0, 8,
	})
	maskMatrix := mat.NewDense(3, 3, []float64{
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
	})

	rowCover := mat.NewVecDense(3, []float64{0, 0, 0})
	colCover := mat.NewVecDense(3, []float64{0, 0, 0})

	rPrimed, cPrimed, found := mapper.step04(matrix, maskMatrix, rowCover, colCover)

	require.True(t, found)
	require.True(t, mat.Equal(rowCover, mat.NewVecDense(3, []float64{0, 0, 0})))
	require.True(t, mat.Equal(colCover, mat.NewVecDense(3, []float64{0, 0, 0})))
	require.True(t, mat.Equal(maskMatrix, mat.NewDense(3, 3, []float64{
		2, 0, 0,
		0, 0, 0,
		0, 0, 0,
	})))
	require.Equal(t, 0, rPrimed)
	require.Equal(t, 0, cPrimed)

	//TestCase 2
	matrix = mat.NewDense(3, 3, []float64{
		0, 22, 0,
		22, 5, 0,
		48, 0, 8,
	})
	maskMatrix = mat.NewDense(3, 3, []float64{
		0, 0, 1,
		0, 0, 0,
		0, 0, 0,
	})

	rowCover = mat.NewVecDense(3, []float64{0, 0, 0})
	colCover = mat.NewVecDense(3, []float64{0, 0, 1})

	rPrimed, cPrimed, found = mapper.step04(matrix, maskMatrix, rowCover, colCover)

	require.True(t, found)
	require.True(t, mat.Equal(rowCover, mat.NewVecDense(3, []float64{1, 0, 0})))
	require.True(t, mat.Equal(colCover, mat.NewVecDense(3, []float64{0, 0, 0})))
	require.True(t, mat.Equal(maskMatrix, mat.NewDense(3, 3, []float64{
		2, 0, 1,
		0, 0, 2,
		0, 0, 0,
	})))
	require.Equal(t, 1, rPrimed)
	require.Equal(t, 2, cPrimed)
}
